package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLRUCache_Basic(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize: 3,
	})
	defer cache.Close()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, found := cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test cache miss
	_, found = cache.Get(ctx, "nonexistent")
	if found {
		t.Error("Expected cache miss for nonexistent key")
	}

	// Test statistics
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Sets != 1 {
		t.Errorf("Expected 1 set, got %d", stats.Sets)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize: 3,
	})
	defer cache.Close()

	// Fill cache to capacity
	cache.Set(ctx, "key1", "value1", 0)
	cache.Set(ctx, "key2", "value2", 0)
	cache.Set(ctx, "key3", "value3", 0)

	// Access key1 to make it most recently used
	cache.Get(ctx, "key1")

	// Add new item, should evict key2 (least recently used)
	cache.Set(ctx, "key4", "value4", 0)

	// Check that key2 was evicted
	_, found := cache.Get(ctx, "key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}

	// Check that other keys are still present
	_, found = cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected key1 to be present")
	}
	_, found = cache.Get(ctx, "key3")
	if !found {
		t.Error("Expected key3 to be present")
	}
	_, found = cache.Get(ctx, "key4")
	if !found {
		t.Error("Expected key4 to be present")
	}

	stats := cache.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
}

func TestLRUCache_TTL(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize:         10,
		CleanupInterval: 100 * time.Millisecond,
	})
	defer cache.Close()

	// Set item with short TTL
	cache.Set(ctx, "key1", "value1", 200*time.Millisecond)

	// Item should be present immediately
	_, found := cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected key1 to be present")
	}

	// Wait for expiration
	time.Sleep(250 * time.Millisecond)

	// Item should be expired
	_, found = cache.Get(ctx, "key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize: 10,
	})
	defer cache.Close()

	// Set and delete
	cache.Set(ctx, "key1", "value1", 0)
	err := cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should not find deleted key
	_, found := cache.Get(ctx, "key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}

	stats := cache.Stats()
	if stats.Deletes != 1 {
		t.Errorf("Expected 1 delete, got %d", stats.Deletes)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize: 10,
	})
	defer cache.Close()

	// Add multiple items
	cache.Set(ctx, "key1", "value1", 0)
	cache.Set(ctx, "key2", "value2", 0)
	cache.Set(ctx, "key3", "value3", 0)

	// Clear cache
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Check size
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// Check items are gone
	_, found := cache.Get(ctx, "key1")
	if found {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize: 100,
	})
	defer cache.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent sets
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				cache.Set(ctx, key, value, 0)
			}
		}(i)
	}

	// Concurrent gets
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Get(ctx, key)
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	cache.Set(ctx, "test", "value", 0)
	val, found := cache.Get(ctx, "test")
	if !found || val != "value" {
		t.Error("Cache not functional after concurrent operations")
	}
}

func TestLRUCache_MemoryLimit(t *testing.T) {
	ctx := context.Background()
	cache := NewLRUCache(Options{
		MaxSize:  100,
		MaxBytes: 1000, // Very small limit for testing
	})
	defer cache.Close()

	// Add items until memory limit is reached
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("This is a longer value to consume more memory: %d", i)
		cache.Set(ctx, key, value, 0)
	}

	// Check that total bytes doesn't exceed limit
	stats := cache.Stats()
	if stats.TotalBytes > 1000 {
		t.Errorf("Total bytes %d exceeds limit 1000", stats.TotalBytes)
	}

	// Check that evictions occurred
	if stats.Evictions == 0 {
		t.Error("Expected evictions due to memory limit")
	}
}

func TestStatistics_HitRate(t *testing.T) {
	tests := []struct {
		name     string
		hits     int64
		misses   int64
		expected float64
	}{
		{"All hits", 10, 0, 100.0},
		{"All misses", 0, 10, 0.0},
		{"Half and half", 5, 5, 50.0},
		{"No operations", 0, 0, 0.0},
		{"More hits", 75, 25, 75.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &Statistics{
				Hits:   tt.hits,
				Misses: tt.misses,
			}
			rate := stats.HitRate()
			if rate != tt.expected {
				t.Errorf("Expected hit rate %.2f, got %.2f", tt.expected, rate)
			}
		})
	}
}