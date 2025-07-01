package cache

import (
	"context"
	"sync"
	"time"

	"proxydav/pkg/types"
)

// MetadataCache implements an in-memory cache for file metadata
type MetadataCache struct {
	cache   map[string]*types.FileMetadata
	mutex   sync.RWMutex
	ttl     time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
	maxSize int
}

// New creates a new metadata cache with the specified TTL and max size
func New(ttl time.Duration, maxSize int) *MetadataCache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &MetadataCache{
		cache:   make(map[string]*types.FileMetadata),
		ttl:     ttl,
		ctx:     ctx,
		cancel:  cancel,
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Close gracefully stops the cache cleanup goroutine
func (c *MetadataCache) Close() {
	c.cancel()
}

// Get retrieves metadata from the cache
func (c *MetadataCache) Get(url string) *types.FileMetadata {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metadata, exists := c.cache[url]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Since(metadata.CachedAt) > c.ttl {
		return nil
	}

	return metadata
}

// Set stores metadata in the cache
func (c *MetadataCache) Set(url string, metadata *types.FileMetadata) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict items
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	metadata.CachedAt = time.Now()
	c.cache[url] = metadata
}

// evictOldest removes the oldest item from the cache
func (c *MetadataCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, metadata := range c.cache {
		if oldestKey == "" || metadata.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = metadata.CachedAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// cleanup periodically removes expired items from the cache
func (c *MetadataCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2) // Clean up twice per TTL period
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.removeExpired()
		}
	}
}

// removeExpired removes all expired items from the cache
func (c *MetadataCache) removeExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, metadata := range c.cache {
		if now.Sub(metadata.CachedAt) > c.ttl {
			delete(c.cache, key)
		}
	}
}

// Size returns the current size of the cache
func (c *MetadataCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// Clear removes all items from the cache
func (c *MetadataCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]*types.FileMetadata)
}
