package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entries  map[string]cacheEntry
	mu       *sync.Mutex
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
		mu:       &sync.Mutex{},
	}

	go cache.readLoop()

	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return val.val, true
}

func (c *Cache) readLoop() {
	ticker := time.NewTicker(c.interval)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, val := range c.entries {
			if now.Sub(val.createdAt) > c.interval {
				delete(c.entries, key)
			}
		}

		c.mu.Unlock()
	}
}
