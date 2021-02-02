package util

import (
	"math"
	"sync"
)

type cacheEntry struct {
	use   uint32
	value interface{}
}

// A simple-minded not-scalable LFU cache.
type Cache struct {
	entries map[string]*cacheEntry
	max     int
	lock    sync.Mutex
}

func NewCache(max int) *Cache {
	return &Cache{
		entries: map[string]*cacheEntry{},
		max:     max,
	}
}

func (c *Cache) Get(key string, fill func(string) interface{}) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.entries[key]
	if ok {
		if e.use != math.MaxUint32 {
			e.use++
		}
		return e.value
	}
	return c.put(key, fill(key))
}

func (c *Cache) Contains(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.entries[key]
	return ok
}

func (c *Cache) Put(key string, v interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.put(key, v)
}

func (c *Cache) put(key string, v interface{}) interface{} {
	if len(c.entries) == c.max {
		var m uint32
		var mk string
		for k, e := range c.entries {
			if mk == "" || e.use < m {
				mk = k
				m = e.use
			}
		}
		delete(c.entries, mk)
	}
	c.entries[key] = &cacheEntry{use: 1, value: v}
	return v
}
