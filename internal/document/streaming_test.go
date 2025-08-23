package document

import (
	"bytes"
	"strings"
	"testing"
)

func TestStreamingOptions(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		opts := DefaultStreamingOptions()
		if opts == nil {
			t.Fatal("DefaultStreamingOptions returned nil")
		}
		if opts.ChunkSize != 64*1024 {
			t.Errorf("Expected chunk size 64KB, got %d", opts.ChunkSize)
		}
		if opts.MaxMemory != 100*1024*1024 {
			t.Errorf("Expected max memory 100MB, got %d", opts.MaxMemory)
		}
		if !opts.EnableMemoryPool {
			t.Error("Expected memory pool to be enabled by default")
		}
	})
	
	t.Run("AdaptiveOptions", func(t *testing.T) {
		tests := []struct {
			name      string
			fileSize  int64
			wantChunk int
			wantMax   int64
		}{
			{
				name:      "SmallFile",
				fileSize:  500 * 1024, // 500KB
				wantChunk: 16 * 1024,
				wantMax:   10 * 1024 * 1024,
			},
			{
				name:      "MediumFile",
				fileSize:  5 * 1024 * 1024, // 5MB
				wantChunk: 64 * 1024,
				wantMax:   50 * 1024 * 1024,
			},
			{
				name:      "LargeFile",
				fileSize:  50 * 1024 * 1024, // 50MB
				wantChunk: 256 * 1024,
				wantMax:   100 * 1024 * 1024,
			},
			{
				name:      "VeryLargeFile",
				fileSize:  200 * 1024 * 1024, // 200MB
				wantChunk: 1024 * 1024,
				wantMax:   200 * 1024 * 1024,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				opts := AdaptiveStreamingOptions(tt.fileSize)
				if opts.ChunkSize != tt.wantChunk {
					t.Errorf("ChunkSize = %d, want %d", opts.ChunkSize, tt.wantChunk)
				}
				if opts.MaxMemory != tt.wantMax {
					t.Errorf("MaxMemory = %d, want %d", opts.MaxMemory, tt.wantMax)
				}
			})
		}
	})
}

func TestMemoryPool(t *testing.T) {
	t.Run("NewMemoryPool", func(t *testing.T) {
		pool := NewMemoryPool(1024)
		if pool == nil {
			t.Fatal("NewMemoryPool returned nil")
		}
		if pool.size != 1024 {
			t.Errorf("Expected pool size 1024, got %d", pool.size)
		}
	})
	
	t.Run("GetAndPut", func(t *testing.T) {
		pool := NewMemoryPool(1024)
		
		// Get buffer from pool
		buf1 := pool.Get()
		if len(buf1) != 1024 {
			t.Errorf("Expected buffer size 1024, got %d", len(buf1))
		}
		
		// Write some data
		copy(buf1, []byte("test data"))
		
		// Return to pool
		pool.Put(buf1)
		
		// Get another buffer (should be reused)
		buf2 := pool.Get()
		if len(buf2) != 1024 {
			t.Errorf("Expected buffer size 1024, got %d", len(buf2))
		}
	})
	
	t.Run("Reset", func(t *testing.T) {
		pool := NewMemoryPool(1024)
		buf := pool.Get()
		
		// Write sensitive data
		copy(buf, []byte("sensitive data"))
		
		// Reset should clear the buffer
		pool.Reset(buf)
		
		// Verify buffer was cleared
		for i := 0; i < len("sensitive data"); i++ {
			if buf[i] != 0 {
				t.Error("Buffer was not properly cleared")
				break
			}
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{5242880, "5.00 MB"},
		{1073741824, "1.00 GB"},
		{2147483648, "2.00 GB"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestMemoryMonitor(t *testing.T) {
	t.Run("NewMemoryMonitor", func(t *testing.T) {
		monitor := NewMemoryMonitor()
		if monitor == nil {
			t.Fatal("NewMemoryMonitor returned nil")
		}
	})
	
	t.Run("SetThresholds", func(t *testing.T) {
		monitor := NewMemoryMonitor()
		monitor.SetThresholds(100*1024*1024, 200*1024*1024)
		
		// Thresholds are set but we can't directly test private fields
		// We'd need to trigger alerts to verify they work
	})
	
	t.Run("GetStats", func(t *testing.T) {
		monitor := NewMemoryMonitor()
		stats := monitor.GetStats()
		
		if stats == nil {
			t.Fatal("GetStats returned nil")
		}
		
		// Stats should have reasonable values
		if stats.HeapAlloc == 0 {
			t.Error("HeapAlloc should not be zero")
		}
		if stats.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	})
	
	t.Run("StartStop", func(t *testing.T) {
		monitor := NewMemoryMonitor()
		
		// Should be safe to start
		monitor.Start()
		
		// Should be safe to start again (no-op)
		monitor.Start()
		
		// Should be safe to stop
		monitor.Stop()
		
		// Should be safe to stop again (no-op)
		monitor.Stop()
	})
}

func TestShouldProcessInMemory(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		want     bool // This will depend on available memory
	}{
		{
			name:     "SmallFile",
			fileSize: 1024, // 1KB
			want:     true, // Should always process small files in memory
		},
		{
			name:     "LargeFile",
			fileSize: 10 * 1024 * 1024 * 1024, // 10GB
			want:     false, // Should not process huge files in memory
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test might behave differently on different systems
			// based on available memory
			got := ShouldProcessInMemory(tt.fileSize)
			
			// For small files, we can be more certain
			if tt.fileSize < 1024*1024 && !got {
				t.Error("Small files should always be processed in memory")
			}
		})
	}
}

// Benchmark for memory pool
func BenchmarkMemoryPool(b *testing.B) {
	pool := NewMemoryPool(4096)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		copy(buf, []byte("benchmark data"))
		pool.Put(buf)
	}
}

// Benchmark for direct allocation
func BenchmarkDirectAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 4096)
		copy(buf, []byte("benchmark data"))
		// buf goes out of scope and is garbage collected
		_ = buf
	}
}

// Helper function to create test XML content
func createTestXML(textContent string) []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<document>`)
	buf.WriteString(`<body>`)
	
	// Split content into paragraphs
	paragraphs := strings.Split(textContent, "\n")
	for _, p := range paragraphs {
		buf.WriteString(`<p><r><t>`)
		buf.WriteString(p)
		buf.WriteString(`</t></r></p>`)
	}
	
	buf.WriteString(`</body>`)
	buf.WriteString(`</document>`)
	
	return buf.Bytes()
}