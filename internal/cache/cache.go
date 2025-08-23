package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache implementations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, bool)
	
	// Set stores a value in the cache with expiration
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	
	// Clear removes all values from the cache
	Clear(ctx context.Context) error
	
	// Size returns the number of items in the cache
	Size() int
	
	// Stats returns cache statistics
	Stats() *Statistics
}

// Statistics holds cache performance metrics
type Statistics struct {
	Hits       int64     // Number of cache hits
	Misses     int64     // Number of cache misses
	Sets       int64     // Number of set operations
	Deletes    int64     // Number of delete operations
	Evictions  int64     // Number of evictions
	Size       int       // Current number of items
	MaxSize    int       // Maximum capacity
	TotalBytes int64     // Estimated memory usage
	LastReset  time.Time // Last statistics reset time
}

// HitRate calculates the cache hit rate percentage
func (s *Statistics) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// Options configures cache behavior
type Options struct {
	MaxSize      int           // Maximum number of items (0 = unlimited)
	MaxBytes     int64         // Maximum memory usage in bytes (0 = unlimited)
	DefaultTTL   time.Duration // Default TTL for items
	CleanupInterval time.Duration // Interval for expired item cleanup
}

// DefaultOptions returns sensible default cache options
func DefaultOptions() Options {
	return Options{
		MaxSize:      1000,
		MaxBytes:     100 * 1024 * 1024, // 100MB
		DefaultTTL:   1 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	}
}