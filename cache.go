package ikuaisdk

import (
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type ResponseCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	cache := &ResponseCache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	go cache.cleanupExpired()

	return cache
}

func (c *ResponseCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

func (c *ResponseCache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ResponseCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *ResponseCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

func (c *ResponseCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Client) GetCachedResponse(key string) (interface{}, bool) {
	if c.cache == nil {
		return nil, false
	}
	return c.cache.Get(key)
}

func (c *Client) SetCachedResponse(key string, data interface{}) {
	if c.cache == nil {
		return
	}
	c.cache.Set(key, data)
}

func (c *Client) ClearResponseCache() {
	if c.cache == nil {
		return
	}
	c.cache.Clear()
}

func (c *Client) EnableCache(ttl time.Duration) {
	c.cache = NewResponseCache(ttl)
}

func (c *Client) DisableCache() {
	if c.cache != nil {
		c.cache.Clear()
		c.cache = nil
	}
}
