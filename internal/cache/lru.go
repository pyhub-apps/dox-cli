package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// item represents a cache entry
type item struct {
	key        string
	value      interface{}
	expiration time.Time
	size       int64
}

// isExpired checks if the item has expired
func (i *item) isExpired() bool {
	if i.expiration.IsZero() {
		return false
	}
	return time.Now().After(i.expiration)
}

// LRUCache implements a Least Recently Used cache with TTL support
type LRUCache struct {
	mu        sync.RWMutex
	items     map[string]*list.Element
	lruList   *list.List
	options   Options
	stats     Statistics
	stopCleanup chan struct{}
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(options Options) *LRUCache {
	if options.MaxSize <= 0 {
		options.MaxSize = 1000
	}
	
	cache := &LRUCache{
		items:     make(map[string]*list.Element),
		lruList:   list.New(),
		options:   options,
		stats:     Statistics{LastReset: time.Now(), MaxSize: options.MaxSize},
		stopCleanup: make(chan struct{}),
	}
	
	// Start cleanup goroutine if interval is set
	if options.CleanupInterval > 0 {
		go cache.cleanupExpired()
	}
	
	return cache
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	element, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}
	
	item := element.Value.(*item)
	
	// Check if expired
	if item.isExpired() {
		c.removeElement(element)
		c.stats.Misses++
		return nil, false
	}
	
	// Move to front (most recently used)
	c.lruList.MoveToFront(element)
	c.stats.Hits++
	
	return item.value, true
}

// Set stores a value in the cache with expiration
func (c *LRUCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Calculate expiration time
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else if c.options.DefaultTTL > 0 {
		expiration = time.Now().Add(c.options.DefaultTTL)
	}
	
	// Estimate size (simplified - in production, use more accurate sizing)
	size := int64(len(key)) + estimateSize(value)
	
	// Check if key already exists
	if element, exists := c.items[key]; exists {
		// Update existing item
		item := element.Value.(*item)
		c.stats.TotalBytes -= item.size
		item.value = value
		item.expiration = expiration
		item.size = size
		c.stats.TotalBytes += size
		c.lruList.MoveToFront(element)
	} else {
		// Add new item
		newItem := &item{
			key:        key,
			value:      value,
			expiration: expiration,
			size:       size,
		}
		
		// Check capacity and evict if necessary
		for c.options.MaxSize > 0 && c.lruList.Len() >= c.options.MaxSize {
			c.evictOldest()
		}
		
		// Check memory limit
		for c.options.MaxBytes > 0 && c.stats.TotalBytes+size > c.options.MaxBytes {
			if c.lruList.Len() == 0 {
				break // Can't evict anything
			}
			c.evictOldest()
		}
		
		element := c.lruList.PushFront(newItem)
		c.items[key] = element
		c.stats.TotalBytes += size
	}
	
	c.stats.Sets++
	c.stats.Size = c.lruList.Len()
	
	return nil
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if element, exists := c.items[key]; exists {
		c.removeElement(element)
		c.stats.Deletes++
	}
	
	return nil
}

// Clear removes all values from the cache
func (c *LRUCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*list.Element)
	c.lruList.Init()
	c.stats.TotalBytes = 0
	c.stats.Size = 0
	
	return nil
}

// Size returns the number of items in the cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() *Statistics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	stats := c.stats
	stats.Size = c.lruList.Len()
	return &stats
}

// Close stops the cleanup goroutine
func (c *LRUCache) Close() {
	close(c.stopCleanup)
}

// removeElement removes an element from the cache (must be called with lock held)
func (c *LRUCache) removeElement(element *list.Element) {
	item := element.Value.(*item)
	delete(c.items, item.key)
	c.lruList.Remove(element)
	c.stats.TotalBytes -= item.size
	c.stats.Size = c.lruList.Len()
}

// evictOldest removes the least recently used item (must be called with lock held)
func (c *LRUCache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest != nil {
		c.removeElement(oldest)
		c.stats.Evictions++
	}
}

// cleanupExpired periodically removes expired items
func (c *LRUCache) cleanupExpired() {
	ticker := time.NewTicker(c.options.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired items from the cache
func (c *LRUCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var toRemove []*list.Element
	
	// Find expired items
	for element := c.lruList.Back(); element != nil; element = element.Prev() {
		item := element.Value.(*item)
		if item.isExpired() {
			toRemove = append(toRemove, element)
		}
	}
	
	// Remove expired items
	for _, element := range toRemove {
		c.removeElement(element)
		c.stats.Evictions++
	}
}

// estimateSize estimates the memory size of a value
func estimateSize(v interface{}) int64 {
	switch val := v.(type) {
	case string:
		return int64(len(val))
	case []byte:
		return int64(len(val))
	case int, int32, int64, uint, uint32, uint64:
		return 8
	case bool:
		return 1
	default:
		// For complex types, this is a rough estimate
		// In production, use a more sophisticated approach
		return 100
	}
}