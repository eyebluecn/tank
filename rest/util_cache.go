package rest

import (
	"errors"
	"fmt"
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
	duration time.Duration
	//创建时间
	createTime time.Time
	//最后访问时间
	accessTime time.Time
	//访问次数
	count int64
	// 在删除缓存项之前调用的回调函数
	deleteCallback func(key interface{})
}

//新建一项缓存
func NewCacheItem(key interface{}, duration time.Duration, data interface{}) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:            key,
		duration:       duration,
		createTime:     t,
		accessTime:     t,
		count:          0,
		deleteCallback: nil,
		data:           data,
	}
}

//手动获取一下，保持该项
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessTime = time.Now()
	item.count++
}

//返回生命周期
func (item *CacheItem) Duration() time.Duration {
	return item.duration
}

//返回访问时间。可能并发，加锁
func (item *CacheItem) AccessTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessTime
}

//返回创建时间
func (item *CacheItem) CreateTime() time.Time {
	return item.createTime
}

//返回访问时间。可能并发，加锁
func (item *CacheItem) Count() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.count
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
func (item *CacheItem) SetDeleteCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.deleteCallback = f
}

// 统一管理缓存项的表
type CacheTable struct {
	sync.RWMutex

	//所有缓存项
	items map[interface{}]*CacheItem
	// 触发缓存清理的定时器
	cleanupTimer *time.Timer
	// 缓存清理周期
	cleanupInterval time.Duration
	// 获取一个不存在的缓存项时的回调函数
	loadData func(key interface{}, args ...interface{}) *CacheItem
	// 向缓存表增加缓存项时的回调函数
	addedCallback func(item *CacheItem)
	// 从缓存表删除一个缓存项时的回调函数
	deleteCallback func(item *CacheItem)
}

// 返回缓存中存储有多少项
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
func (table *CacheTable) SetAddedCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addedCallback = f
}

// 删除时的回调函数
func (table *CacheTable) SetDeleteCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

//终结检查，被自调整的时间触发
func (table *CacheTable) checkExpire() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after %v for table", table.cleanupInterval)
	} else {
		table.log("Expiration check installed for table")
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
		duration := item.duration
		accessTime := item.accessTime
		item.RUnlock()

		// 0永久有效
		if duration == 0 {
			continue
		}
		if now.Sub(accessTime) >= duration {
			//缓存项已经过期
			_, e := table.Delete(key)
			if e != nil {
				table.log("删除缓存项时出错 %v", e.Error())
			}
		} else {
			//查找最靠近结束生命周期的项目
			if smallestDuration == 0 || duration-now.Sub(accessTime) < smallestDuration {
				smallestDuration = duration - now.Sub(accessTime)
			}
		}
	}

	// 为下次清理设置间隔，自触发机制
	table.Lock()
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go SafeMethod(table.checkExpire)
		})
	}
	table.Unlock()
}

// 添加缓存项
func (table *CacheTable) Add(key interface{}, duration time.Duration, data interface{}) *CacheItem {
	item := NewCacheItem(key, duration, data)

	// 将缓存项放入表中
	table.Lock()
	table.log("Adding item with key %v and lifespan of %v to table", key, duration)
	table.items[key] = item

	// 取出需要的东西，释放锁
	expDur := table.cleanupInterval
	addedItem := table.addedCallback
	table.Unlock()

	// 有回调函数便执行回调
	if addedItem != nil {
		addedItem(item)
	}

	// 如果我们没有设置任何心跳检查定时器或者找一个即将迫近的项目
	if duration > 0 && (expDur == 0 || duration < expDur) {
		table.checkExpire()
	}

	return item
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
	deleteCallback := table.deleteCallback
	table.RUnlock()

	// 调用删除回调函数
	if deleteCallback != nil {
		deleteCallback(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.deleteCallback != nil {
		r.deleteCallback(key)
	}

	table.Lock()
	defer table.Unlock()
	table.log("Deleting item with key %v created on %v and hit %v times from table", key, r.createTime, r.count)
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

	item := NewCacheItem(key, lifeSpan, data)
	table.log("Adding item with key %v and lifespan of %v to table", key, lifeSpan)
	table.items[key] = item

	// 取出需要的内容，释放锁
	expDur := table.cleanupInterval
	addedItem := table.addedCallback
	table.Unlock()

	// 添加回调函数
	if addedItem != nil {
		addedItem(item)
	}

	// 触发过期检查
	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.checkExpire()
	}
	return true
}

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
			table.Add(key, item.duration, item.data)
			return item, nil
		}

		return nil, errors.New("无法加载到缓存值")
	}

	//没有找到任何东西，返回nil.
	return nil, nil
}

// 删除缓存表中的所有项目
func (table *CacheTable) Truncate() {
	table.Lock()
	defer table.Unlock()

	table.log("Truncate table")

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
		p[i] = CacheItemPair{k, v.count}
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
func (table *CacheTable) log(format string, v ...interface{}) {
	LOGGER.Info(format, v...)
}

//新建一个缓存Table
func NewCacheTable() *CacheTable {
	return &CacheTable{
		items: make(map[interface{}]*CacheItem),
	}
}
