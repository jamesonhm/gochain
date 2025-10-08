package yahoo

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu       sync.RWMutex
	entries  map[string]cacheEntry
	stopChan chan struct{}
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		entries:  make(map[string]cacheEntry),
		stopChan: make(chan struct{}),
	}
	go cache.reapLoop(interval)
	return &cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.entries[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return data.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.pruneEntries(interval)
		}
	}
}

func (c *Cache) pruneEntries(interval time.Duration) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.entries {
		if now.Sub(v.createdAt) > interval {
			delete(c.entries, k)
		}
	}
}
