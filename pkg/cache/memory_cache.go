package cache

import "sync"

type MemoryCache struct {
	cache map[string]interface{}
	lock  sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: make(map[string]interface{}),
	}
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, found := m.cache[key]
	return val, found
}

func (m *MemoryCache) Set(key string, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[key] = value
}
