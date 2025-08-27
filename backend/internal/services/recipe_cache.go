package services

import (
	"sync"
	"time"
)

// CacheEntry represents a cached recipe generation result
type CacheEntry struct {
	Data      *GenerationResult
	ExpiresAt time.Time
}

// RecipeCache implements an in-memory cache for recipe generation results
type RecipeCache struct {
	mu      sync.RWMutex
	data    map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
	cleanup *time.Ticker
}

// NewRecipeCache creates a new recipe cache
func NewRecipeCache(maxSize int, ttl time.Duration) *RecipeCache {
	cache := &RecipeCache{
		data:    make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		cleanup: time.NewTicker(ttl / 4), // Clean up every 1/4 of TTL
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from the cache
func (c *RecipeCache) Get(key string) *GenerationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Don't delete here to avoid write lock, let cleanup goroutine handle it
		return nil
	}

	// Return a copy to avoid race conditions
	result := *entry.Data
	return &result
}

// Set stores a value in the cache
func (c *RecipeCache) Set(key string, value *GenerationResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to make space
	if len(c.data) >= c.maxSize {
		c.evictOldest()
	}

	// Store the entry
	c.data[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a key from the cache
func (c *RecipeCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear removes all entries from the cache
func (c *RecipeCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
}

// Size returns the current number of entries in the cache
func (c *RecipeCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// GetStats returns cache statistics
func (c *RecipeCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	expiredCount := 0

	for _, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_entries":   len(c.data),
		"expired_entries": expiredCount,
		"active_entries":  len(c.data) - expiredCount,
		"max_size":        c.maxSize,
		"ttl_seconds":     c.ttl.Seconds(),
	}
}

// evictOldest removes the oldest entry to make space
func (c *RecipeCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// cleanupExpired removes expired entries periodically
func (c *RecipeCache) cleanupExpired() {
	for range c.cleanup.C {
		c.mu.Lock()
		now := time.Now()

		for key, entry := range c.data {
			if now.After(entry.ExpiresAt) {
				delete(c.data, key)
			}
		}

		c.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (c *RecipeCache) Stop() {
	if c.cleanup != nil {
		c.cleanup.Stop()
	}
}

// Keys returns all cache keys (for debugging)
func (c *RecipeCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}

	return keys
}

// HasKey checks if a key exists in the cache
func (c *RecipeCache) HasKey(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return false
	}

	return time.Now().Before(entry.ExpiresAt)
}
