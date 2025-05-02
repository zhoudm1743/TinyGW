package io

import (
	"sync"
)

var IConcurrentMap = NewConcurrentMap()

// ConcurrentMap 是一个并发安全的 map[string][]byte 结构
type ConcurrentMap struct {
	mu    sync.RWMutex
	items map[string][]byte
}

// NewConcurrentMap 创建一个新的 ConcurrentMap
func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		items: make(map[string][]byte),
	}
}

// Get 从 map 中获取指定 key 的值
func (cm *ConcurrentMap) Get(key string) ([]byte, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	value, exists := cm.items[key]
	return value, exists
}

// Set 将指定的 key 和 value 存入 map
func (cm *ConcurrentMap) Set(key string, value []byte) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.items[key] = value
}

// Delete 从 map 中删除指定的 key
func (cm *ConcurrentMap) Delete(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.items, key)
}

// Len 返回 map 中元素的数量
func (cm *ConcurrentMap) Len() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.items)
}

// Keys 返回 map 中所有的 key
func (cm *ConcurrentMap) Keys() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	keys := make([]string, 0, len(cm.items))
	for key := range cm.items {
		keys = append(keys, key)
	}
	return keys
}
