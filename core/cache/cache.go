package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type item struct {
	value      any
	expiration int64 // unix timestamp, 0 means no expire
}

type entry struct {
	key  string
	item *item
}

type Cache struct {
	mu     sync.RWMutex
	items  map[string]*list.Element
	lru    *list.List // *entry
	maxCap int
	stop   chan struct{}
	hit    uint64
	miss   uint64
}

// NewCache 创建一个高性能LRU内存缓存实例，maxCap<=0表示无限容量
func NewCache(maxCap int) *Cache {
	c := &Cache{
		items:  make(map[string]*list.Element),
		lru:    list.New(),
		maxCap: maxCap,
		stop:   make(chan struct{}),
	}
	go c.gc()
	return c
}

// Set 设置缓存，duration<=0 表示永不过期
func (c *Cache) Set(key string, value any, duration time.Duration) {
	exp := int64(0)
	if duration > 0 {
		exp = time.Now().Add(duration).UnixNano()
	}
	c.mu.Lock()
	if ele, ok := c.items[key]; ok {
		ent := ele.Value.(*entry)
		ent.item.value = value
		ent.item.expiration = exp
		c.lru.MoveToFront(ele)
	} else {
		ent := &entry{key: key, item: &item{value: value, expiration: exp}}
		ele := c.lru.PushFront(ent)
		c.items[key] = ele
		if c.maxCap > 0 && c.lru.Len() > c.maxCap {
			c.removeOldest()
		}
	}
	c.mu.Unlock()
}

// SetTyped 泛型安全设置
func SetTyped[T any](c *Cache, key string, value T, duration time.Duration) {
	c.Set(key, value, duration)
}

// SetBatch 批量设置
func (c *Cache) SetBatch(data map[string]any, duration time.Duration) {
	for k, v := range data {
		c.Set(k, v, duration)
	}
}

// Get 获取缓存，存在且未过期返回 value, true，否则 nil, false
func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	ele, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		atomic.AddUint64(&c.miss, 1)
		return nil, false
	}
	ent := ele.Value.(*entry)
	if ent.item.expiration > 0 && time.Now().UnixNano() > ent.item.expiration {
		c.removeElement(ele)
		c.mu.Unlock()
		atomic.AddUint64(&c.miss, 1)
		return nil, false
	}
	c.lru.MoveToFront(ele)
	val := ent.item.value
	c.mu.Unlock()
	atomic.AddUint64(&c.hit, 1)
	return val, true
}

// GetTyped 泛型安全获取
func GetTyped[T any](c *Cache, key string) (T, bool) {
	v, ok := c.Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	val, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return val, true
}

// GetBatch 批量获取
func (c *Cache) GetBatch(keys []string) map[string]any {
	res := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := c.Get(k); ok {
			res[k] = v
		}
	}
	return res
}

// Delete 删除缓存
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	if ele, ok := c.items[key]; ok {
		c.removeElement(ele)
	}
	c.mu.Unlock()
}

// DeleteBatch 批量删除
func (c *Cache) DeleteBatch(keys []string) {
	c.mu.Lock()
	for _, k := range keys {
		if ele, ok := c.items[k]; ok {
			c.removeElement(ele)
		}
	}
	c.mu.Unlock()
}

// Stats 缓存统计
func (c *Cache) Stats() (hit, miss, total uint64) {
	c.mu.RLock()
	total = uint64(len(c.items))
	c.mu.RUnlock()
	hit = atomic.LoadUint64(&c.hit)
	miss = atomic.LoadUint64(&c.miss)
	return
}

// removeOldest 移除最久未使用
func (c *Cache) removeOldest() {
	ele := c.lru.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// removeElement 从链表和map移除
func (c *Cache) removeElement(ele *list.Element) {
	ent := ele.Value.(*entry)
	delete(c.items, ent.key)
	c.lru.Remove(ele)
}

// gc 自动清理过期项
func (c *Cache) gc() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			c.mu.Lock()
			for ele := c.lru.Back(); ele != nil; {
				prev := ele.Prev()
				ent := ele.Value.(*entry)
				if ent.item.expiration > 0 && now > ent.item.expiration {
					c.removeElement(ele)
				}
				ele = prev
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Close 停止自动清理
func (c *Cache) Close() {
	close(c.stop)
}
