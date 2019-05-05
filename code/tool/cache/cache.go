package cache

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

//cache2go https://github.com/muesli/cache2go
type Item struct {
	//read write lock
	sync.RWMutex
	key  interface{}
	data interface{}
	// cache duration.
	duration time.Duration
	// create time
	createTime time.Time
	//last access time
	accessTime time.Time
	//visit times
	count int64
	// callback after deleting
	deleteCallback func(key interface{})
}

//create item.
func NewItem(key interface{}, duration time.Duration, data interface{}) *Item {
	t := time.Now()
	return &Item{
		key:            key,
		duration:       duration,
		createTime:     t,
		accessTime:     t,
		count:          0,
		deleteCallback: nil,
		data:           data,
	}
}

//keep alive
func (item *Item) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessTime = time.Now()
	item.count++
}

func (item *Item) Duration() time.Duration {
	return item.duration
}

func (item *Item) AccessTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessTime
}

func (item *Item) CreateTime() time.Time {
	return item.createTime
}

func (item *Item) Count() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.count
}

func (item *Item) Key() interface{} {
	return item.key
}

func (item *Item) Data() interface{} {
	return item.data
}

func (item *Item) SetDeleteCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.deleteCallback = f
}

// table for managing cache items
type Table struct {
	sync.RWMutex

	//all cache items
	items map[interface{}]*Item
	// trigger cleanup
	cleanupTimer *time.Timer
	// cleanup interval
	cleanupInterval time.Duration
	loadData        func(key interface{}, args ...interface{}) *Item
	// callback after adding.
	addedCallback func(item *Item)
	// callback after deleting
	deleteCallback func(item *Item)
}

func (table *Table) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

func (table *Table) Foreach(trans func(key interface{}, item *Item)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

func (table *Table) SetDataLoader(f func(interface{}, ...interface{}) *Item) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

func (table *Table) SetAddedCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.addedCallback = f
}

func (table *Table) SetDeleteCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

func (table *Table) RunWithRecovery(f func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("occur error %v \r\n", err)
		}
	}()

	f()
}

func (table *Table) checkExpire() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after %v for table", table.cleanupInterval)
	} else {
		table.log("Expiration check installed for table")
	}

	// in order to not take the lock. use temp items.
	items := table.items
	table.Unlock()

	//in order to make timer more precise, update now every loop.
	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range items {
		//take out our things, in order not to take the lock.
		item.RLock()
		duration := item.duration
		accessTime := item.accessTime
		item.RUnlock()

		// 0 means valid.
		if duration == 0 {
			continue
		}
		if now.Sub(accessTime) >= duration {
			//cache item expired.
			_, e := table.Delete(key)
			if e != nil {
				table.log("occur error while deleting %v", e.Error())
			}
		} else {
			//find the most possible expire item.
			if smallestDuration == 0 || duration-now.Sub(accessTime) < smallestDuration {
				smallestDuration = duration - now.Sub(accessTime)
			}
		}
	}

	//trigger next clean
	table.Lock()
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.RunWithRecovery(table.checkExpire)
		})
	}
	table.Unlock()
}

// add item
func (table *Table) Add(key interface{}, duration time.Duration, data interface{}) *Item {
	item := NewItem(key, duration, data)

	table.Lock()
	table.log("Adding item with key %v and lifespan of %d to table", key, duration)
	table.items[key] = item

	expDur := table.cleanupInterval
	addedItem := table.addedCallback
	table.Unlock()

	if addedItem != nil {
		addedItem(item)
	}

	//find the most possible expire item.
	if duration > 0 && (expDur == 0 || duration < expDur) {
		table.checkExpire()
	}

	return item
}

func (table *Table) Delete(key interface{}) (*Item, error) {
	table.RLock()
	r, ok := table.items[key]
	if !ok {
		table.RUnlock()
		return nil, errors.New(fmt.Sprintf("no item with key %s", key))
	}

	deleteCallback := table.deleteCallback
	table.RUnlock()

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
	table.log("Deleting item with key %v created on %s and hit %d times from table", key, r.createTime, r.count)
	delete(table.items, key)

	return r, nil
}

//check exist.
func (table *Table) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

//if exist, return false. if not exist add a key and return true.
func (table *Table) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewItem(key, lifeSpan, data)
	table.log("Adding item with key %v and lifespan of %d to table", key, lifeSpan)
	table.items[key] = item

	expDur := table.cleanupInterval
	addedItem := table.addedCallback
	table.Unlock()

	if addedItem != nil {
		addedItem(item)
	}

	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.checkExpire()
	}
	return true
}

func (table *Table) Value(key interface{}, args ...interface{}) (*Item, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()

	if ok {
		//update visit count and visit time.
		r.KeepAlive()
		return r, nil
	}

	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.duration, item.data)
			return item, nil
		}

		return nil, errors.New("cannot load item")
	}

	return nil, nil
}

// truncate a table.
func (table *Table) Truncate() {
	table.Lock()
	defer table.Unlock()

	table.log("Truncate table")

	table.items = make(map[interface{}]*Item)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

//support table sort
type ItemPair struct {
	Key         interface{}
	AccessCount int64
}

type ItemPairList []ItemPair

func (p ItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ItemPairList) Len() int           { return len(p) }
func (p ItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

//return most visited.
func (table *Table) MostAccessed(count int64) []*Item {
	table.RLock()
	defer table.RUnlock()

	p := make(ItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = ItemPair{k, v.count}
		i++
	}
	sort.Sort(p)

	var r []*Item
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

// print log.
func (table *Table) log(format string, v ...interface{}) {
	//fmt.Printf(format+"\r\n", v)
}

func NewTable() *Table {
	return &Table{
		items: make(map[interface{}]*Item),
	}
}
