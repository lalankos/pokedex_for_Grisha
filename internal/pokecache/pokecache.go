package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu        sync.Mutex
	entries   map[string]cacheEntry
	interval  time.Duration
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
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
	entry, found := c.entries[key]
	if !found {
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) reapLoop() {
	for {
		time.Sleep(c.interval)
		c.mu.Lock()
		threshold := time.Now().Add(-c.interval)
		for key, entry := range c.entries {
			if entry.createdAt.Before(threshold) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}