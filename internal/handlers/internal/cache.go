package internal

import (
	"sync"
	"time"
)

// ComponentCache provides caching for UI components
type ComponentCache struct {
	cache map[string]cacheEntry
	mutex sync.RWMutex
}

type cacheEntry struct {
	html      string
	timestamp time.Time
	ttl       time.Duration
}

// NewComponentCache creates a new component cache
func NewComponentCache() *ComponentCache {
	return &ComponentCache{
		cache: make(map[string]cacheEntry),
	}
}

// Get retrieves a cached component
func (c *ComponentCache) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return "", false
	}

	// Check if entry is expired
	if time.Since(entry.timestamp) > entry.ttl {
		// Remove expired entry
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.cache, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return "", false
	}

	return entry.html, true
}

// Set stores a component in cache
func (c *ComponentCache) Set(key string, html string, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = cacheEntry{
		html:      html,
		timestamp: time.Now(),
		ttl:       ttl,
	}
}

// Clear removes all cached entries
func (c *ComponentCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]cacheEntry)
}

// ClearExpired removes expired entries
func (c *ComponentCache) ClearExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.Sub(entry.timestamp) > entry.ttl {
			delete(c.cache, key)
		}
	}
}

// GenerateCacheKey generates a cache key for components
func GenerateCacheKey(componentType string, params map[string]string) string {
	key := componentType
	for k, v := range params {
		key += ":" + k + "=" + v
	}
	return key
}
