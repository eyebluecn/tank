package rest

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

//缓存项
//主要借鉴了cache2go https://github.com/muesli/cache2go
type CacheItem struct {
	sync.RWMutex //读写锁

	//缓存键
	key interface{}
	//缓存值
	data interface{}
	// 缓存项的生命期
	lifeSpan time.Duration
	//创建时间
	createdOn time.Time
	//最后访问时间
	accessedOn time.Time
	//访问次数
	accessCount int64
	// 在删除缓存项之前调用的回调函数
	aboutToExpire func(key interface{})
}

//新建一项缓存
func CreateCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) CacheItem {
	t := time.Now()
	return CacheItem{
		key:           key,
		lifeSpan:      lifeSpan,
		createdOn:     t,
		accessedOn:    t,
		accessCount:   0,
		aboutToExpire: nil,
		data:          data,
	}
}

//手动获取一下，保持该项
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

//返回生命周期
func (item *CacheItem) LifeSpan() time.Duration {
	return item.lifeSpan
}

//返回访问时间。可能并发，加锁
func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

//返回创建时间
func (item *CacheItem) CreatedOn() time.Time {
	return item.createdOn
}

//返回访问时间。可能并发，加锁
func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

//返回key值
func (item *CacheItem) Key() interface{} {
	return item.key
}

//返回数据
func (item *CacheItem) Data() interface{} {
	return item.data
}

//设置回调函数
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = f
}

// 统一管理缓存项的表
type CacheTable struct {
	sync.RWMutex

	//缓存表名
	name string
	//所有缓存项
	items map[interface{}]*CacheItem
	// 触发缓存清理的定时器
	cleanupTimer *time.Timer
	// 缓存清理周期
	cleanupInterval time.Duration
	// 该缓存表的日志
	logger *log.Logger
	// 获取一个不存在的缓存项时的回调函数
	loadData func(key interface{}, args ...interface{}) *CacheItem
	// 向缓存表增加缓存项时的回调函数
	addedItem func(item *CacheItem)
	// 从缓存表删除一个缓存项时的回调函数
	aboutToDeleteItem func(item *CacheItem)
}

// 返回当缓存中存储有多少项
func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// 遍历所有项
func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

// SetDataLoader配置一个数据加载的回调，当尝试去请求一个不存在的key的时候调用
func (table *CacheTable) SetDataLoader(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

// 添加时的回调函数
func (table *CacheTable) SetAddedItemCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addedItem = f
}

// 删除时的回调函数
func (table *CacheTable) SetAboutToDeleteItemCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.aboutToDeleteItem = f
}

// 设置缓存表需要使用的log
func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

//终结检查，被自调整的时间触发
func (table *CacheTable) expirationCheck() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after", table.cleanupInterval, "for table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	// 为了不抢占锁，采用临时的items.
	items := table.items
	table.Unlock()

	//为了定时器更准确，我们需要在每一个循环中更新‘now’，不确定是否是有效率的。
	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range items {
		// 取出我们需要的东西，为了不抢占锁
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		// 0永久有效
		if lifeSpan == 0 {
			continue
		}
		if now.Sub(accessedOn) >= lifeSpan {
			//缓存项已经过期
			_, e := table.Delete(key)
			if e != nil {
				table.log("删除缓存项时出错 ", e.Error())
			}
		} else {
			//查找最靠近结束生命周期的项目
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	// 为下次清理设置间隔，自触发机制
	table.Lock()
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()
}

// 添加缓存项
func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item := CreateCacheItem(key, lifeSpan, data)

	// 将缓存项放入表中
	table.Lock()
	table.log("Adding item with key", key, "and lifespan of", lifeSpan, "to table", table.name)
	table.items[key] = &item

	// 取出需要的东西，释放锁
	expDur := table.cleanupInterval
	addedItem := table.addedItem
	table.Unlock()

	// 有回调函数便执行回调
	if addedItem != nil {
		addedItem(&item)
	}

	// 如果我们没有设置任何心跳检查定时器或者找一个即将迫近的项目
	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.expirationCheck()
	}

	return &item
}

// 从缓存中删除项
func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	if !ok {
		table.RUnlock()
		return nil, errors.New(fmt.Sprintf("没有找到%s对应的记录", key))
	}

	// 取出要用到的东西，释放锁
	aboutToDeleteItem := table.aboutToDeleteItem
	table.RUnlock()

	// 调用删除回调函数
	if aboutToDeleteItem != nil {
		aboutToDeleteItem(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.aboutToExpire != nil {
		r.aboutToExpire(key)
	}

	table.Lock()
	defer table.Unlock()
	table.log("Deleting item with key", key, "created on", r.createdOn, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)

	return r, nil
}

//单纯的检查某个键是否存在
func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

//如果存在，返回false. 如果不存在,就去添加一个键，并且返回true
func (table *CacheTable) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := CreateCacheItem(key, lifeSpan, data)
	table.log("Adding item with key", key, "and lifespan of", lifeSpan, "to table", table.name)
	table.items[key] = &item

	// 取出需要的内容，释放锁
	expDur := table.cleanupInterval
	addedItem := table.addedItem
	table.Unlock()

	// 添加回调函数
	if addedItem != nil {
		addedItem(&item)
	}

	// 触发过期检查
	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.expirationCheck()
	}
	return true
}

// Get an item from the cache and mark it to be kept alive. You can pass
// additional arguments to your DataLoader callback function.
//从缓存中返回一个被标记的并保持活性的值。你可以传附件的参数到DataLoader回调函数
func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()

	if ok {
		// 更新访问次数和访问时间
		r.KeepAlive()
		return r, nil
	}

	// 有加载数据的方式，就通过loadData函数去加载进来
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.lifeSpan, item.data)
			return item, nil
		}

		return nil, errors.New("无法加载到缓存值")
	}

	return nil, errors.New(fmt.Sprintf("没有找到%s对应的记录", key))
}

// 删除缓存表中的所有项目
func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)

	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

//辅助table中排序，统计的
type CacheItemPair struct {
	Key         interface{}
	AccessCount int64
}

type CacheItemPairList []CacheItemPair

func (p CacheItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p CacheItemPairList) Len() int           { return len(p) }
func (p CacheItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

// 返回缓存表中被访问最多的项目
func (table *CacheTable) MostAccessed(count int64) []*CacheItem {
	table.RLock()
	defer table.RUnlock()

	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{k, v.accessCount}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
	c := int64(0)
	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}


// 打印日志
func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		fmt.Println(v...)
		return
	}

	table.logger.Println(v...)
}

var (
	cacheTableMap   = make(map[string]*CacheTable)
	cacheTableMutex sync.RWMutex
)

//统一管理所有的缓存表，如果没有就返回一个新的。
func Cache(table string) *CacheTable {
	cacheTableMutex.RLock()
	t, ok := cacheTableMap[table]
	cacheTableMutex.RUnlock()

	if !ok {
		t = &CacheTable{
			name:  table,
			items: make(map[interface{}]*CacheItem),
		}

		cacheTableMutex.Lock()
		cacheTableMap[table] = t
		cacheTableMutex.Unlock()
	}

	return t
}
