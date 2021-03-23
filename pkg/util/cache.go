// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
