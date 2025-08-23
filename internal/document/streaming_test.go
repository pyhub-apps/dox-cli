package document

import (
	"archive/zip"
	"bytes"
	"os"
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
	
	// Split content into smaller chunks to simulate real Word documents
	// Real Word documents break text into smaller elements
	const maxChunkSize = 100 // Characters per text element
	
	words := strings.Fields(textContent)
	currentChunk := ""
	
	for _, word := range words {
		if len(currentChunk)+len(word)+1 > maxChunkSize && currentChunk != "" {
			// Write current chunk as a paragraph
			buf.WriteString(`<p><r><t>`)
			buf.WriteString(currentChunk)
			buf.WriteString(`</t></r></p>`)
			currentChunk = word
		} else {
			if currentChunk != "" {
				currentChunk += " "
			}
			currentChunk += word
		}
	}
	
	// Write any remaining content
	if currentChunk != "" {
		buf.WriteString(`<p><r><t>`)
		buf.WriteString(currentChunk)
		buf.WriteString(`</t></r></p>`)
	}
	
	buf.WriteString(`</body>`)
	buf.WriteString(`</document>`)
	
	return buf.Bytes()
}

// TestStreamingErrorScenarios tests various error conditions in streaming
func TestStreamingErrorScenarios(t *testing.T) {
	t.Run("FileNotFound", func(t *testing.T) {
		_, err := OpenWordDocumentStreaming("/nonexistent/file.docx", nil)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %v", err)
		}
	})

	t.Run("InvalidFileExtension", func(t *testing.T) {
		// Create a temporary non-docx file
		tmpFile, err := os.CreateTemp("", "test*.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		_, err = OpenWordDocumentStreaming(tmpFile.Name(), nil)
		if err == nil {
			t.Error("Expected error for non-docx file")
		}
		if !strings.Contains(err.Error(), "not a .docx file") {
			t.Errorf("Expected 'not a .docx file' error, got: %v", err)
		}
	})

	t.Run("InvalidZipFormat", func(t *testing.T) {
		// Create a file with .docx extension but invalid content
		tmpFile, err := os.CreateTemp("", "test*.docx")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		
		// Write some invalid data
		tmpFile.WriteString("This is not a valid zip file")
		tmpFile.Close()

		_, err = OpenWordDocumentStreaming(tmpFile.Name(), nil)
		if err == nil {
			t.Error("Expected error for invalid docx format")
		}
		if !strings.Contains(err.Error(), "invalid docx format") {
			t.Errorf("Expected 'invalid docx format' error, got: %v", err)
		}
	})

	t.Run("MemoryLimitExceeded", func(t *testing.T) {
		// Create options with very small memory limit
		opts := &StreamingOptions{
			ChunkSize:        1024,
			MaxMemory:        1, // 1 byte - will always exceed
			EnableMemoryPool: false,
		}

		// This test would require a real docx file
		// For now, we just verify the options are set correctly
		if opts.MaxMemory != 1 {
			t.Error("Memory limit not set correctly")
		}
	})

	t.Run("ClosedDocument", func(t *testing.T) {
		// Create a valid test docx file (simplified)
		tmpFile, err := os.CreateTemp("", "test*.docx")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		
		// Create a minimal valid zip structure
		zipWriter := zip.NewWriter(tmpFile)
		
		// Add document.xml
		docFile, err := zipWriter.Create("word/document.xml")
		if err != nil {
			t.Fatal(err)
		}
		docFile.Write(createTestXML("Test content"))
		
		zipWriter.Close()
		tmpFile.Close()

		// Open and immediately close the document
		doc, err := OpenWordDocumentStreaming(tmpFile.Name(), nil)
		if err != nil {
			t.Fatal(err)
		}
		
		err = doc.Close()
		if err != nil {
			t.Fatal(err)
		}

		// Try to use closed document
		err = doc.ReplaceTextStreaming("test", "replacement")
		if err == nil {
			t.Error("Expected error when using closed document")
		}
		if !strings.Contains(err.Error(), "document is closed") {
			t.Errorf("Expected 'document is closed' error, got: %v", err)
		}
	})

	t.Run("InvalidXMLContent", func(t *testing.T) {
		// Create a docx with malformed XML
		tmpFile, err := os.CreateTemp("", "test*.docx")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		
		zipWriter := zip.NewWriter(tmpFile)
		
		// Add document.xml with invalid XML
		docFile, err := zipWriter.Create("word/document.xml")
		if err != nil {
			t.Fatal(err)
		}
		docFile.Write([]byte("This is not valid XML <unclosed tag"))
		
		zipWriter.Close()
		tmpFile.Close()

		doc, err := OpenWordDocumentStreaming(tmpFile.Name(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer doc.Close()

		// This should handle XML parsing errors gracefully
		err = doc.ReplaceTextStreaming("test", "replacement")
		if err == nil {
			t.Error("Expected error for invalid XML")
		}
		if !strings.Contains(err.Error(), "XML") {
			t.Errorf("Expected XML-related error, got: %v", err)
		}
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		// This test is platform-specific and may not work in all environments
		t.Skip("Permission test is platform-specific")
	})
}

// TestMemoryUsageReduction verifies that streaming actually reduces memory usage
func TestMemoryUsageReduction(t *testing.T) {
	t.Run("StreamingMemoryEfficiency", func(t *testing.T) {
		// Create a test docx file with significant content
		tmpFile, err := os.CreateTemp("", "test*.docx")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		
		// Create a zip with large content
		zipWriter := zip.NewWriter(tmpFile)
		
		// Create a large document.xml (simulate 5MB of text)
		docFile, err := zipWriter.Create("word/document.xml")
		if err != nil {
			t.Fatal(err)
		}
		
		// Generate large content
		largeContent := strings.Repeat("This is test content that will be replaced. ", 100000)
		docFile.Write(createTestXML(largeContent))
		
		zipWriter.Close()
		tmpFile.Close()

		// Open with streaming
		doc, err := OpenWordDocumentStreaming(tmpFile.Name(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer doc.Close()

		// Track memory before operation
		initialMemory := doc.GetMemoryUsage()

		// Perform replacement
		err = doc.ReplaceTextStreaming("test", "verified")
		if err != nil {
			t.Fatal(err)
		}

		// Check memory usage after operation
		finalMemory := doc.GetMemoryUsage()
		
		// Memory usage should be minimal (only current chunk)
		// Should be much less than the full document size
		if finalMemory > 1024*1024 { // 1MB threshold
			t.Errorf("Memory usage too high for streaming: %d bytes", finalMemory)
		}
		
		t.Logf("Initial memory: %d bytes, Final memory: %d bytes", initialMemory, finalMemory)
	})

	t.Run("ChunkSizeRespected", func(t *testing.T) {
		// Create options with specific chunk size
		opts := &StreamingOptions{
			ChunkSize:        4096, // 4KB chunks
			MaxMemory:        10 * 1024 * 1024, // 10MB max
			EnableMemoryPool: true,
		}

		// Create a test docx file
		tmpFile, err := os.CreateTemp("", "test*.docx")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		
		zipWriter := zip.NewWriter(tmpFile)
		docFile, err := zipWriter.Create("word/document.xml")
		if err != nil {
			t.Fatal(err)
		}
		
		// Create content larger than chunk size
		content := strings.Repeat("Test content. ", 1000)
		docFile.Write(createTestXML(content))
		
		zipWriter.Close()
		tmpFile.Close()

		// Open with specific options
		doc, err := OpenWordDocumentStreaming(tmpFile.Name(), opts)
		if err != nil {
			t.Fatal(err)
		}
		defer doc.Close()

		// Perform replacement
		err = doc.ReplaceTextStreaming("Test", "Verified")
		if err != nil {
			t.Fatal(err)
		}

		// Memory usage should be around chunk size, not full document
		memUsage := doc.GetMemoryUsage()
		if memUsage > int64(opts.ChunkSize*2) { // Allow some overhead
			t.Errorf("Memory usage exceeds expected chunk size: %d > %d", memUsage, opts.ChunkSize*2)
		}
	})
}