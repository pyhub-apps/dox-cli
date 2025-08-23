package replace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProcessLargeFile_NonExistentFile(t *testing.T) {
	// Test with non-existent file
	rules := []Rule{
		{Old: "test", New: "replacement"},
	}
	
	opts := DefaultLargeFileOptions()
	result, err := ProcessLargeFile("/non/existent/file.docx", rules, opts)
	
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
	
	if result != nil {
		t.Errorf("Expected nil result for non-existent file, got %v", result)
	}
	
	expectedMsg := "file does not exist"
	if !testContains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestProcessLargeFile_UnsupportedFileType(t *testing.T) {
	// Create a temp file with unsupported extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	rules := []Rule{
		{Old: "test", New: "replacement"},
	}
	
	opts := DefaultLargeFileOptions()
	result, err := ProcessLargeFile(tmpFile, rules, opts)
	
	if err == nil {
		t.Fatal("Expected error for unsupported file type, got nil")
	}
	
	if result != nil {
		t.Errorf("Expected nil result for unsupported file type, got %v", result)
	}
	
	expectedMsg := "unsupported file type"
	if !testContains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestProcessLargeFile_EmptyRules(t *testing.T) {
	// Skip this test as it requires actual Word document creation
	t.Skip("Skipping test that requires Word document creation")
}

func TestProcessLargeFile_StreamingMode(t *testing.T) {
	// Skip this test as it requires actual Word document creation
	t.Skip("Skipping test that requires Word document creation")
}

func TestGetRecommendedOptions(t *testing.T) {
	tests := []struct {
		name              string
		fileSize          int64
		expectStreaming   bool
		expectMonitor     bool
		expectThreshold   int64
	}{
		{
			name:            "Small file (<1MB)",
			fileSize:        500 * 1024, // 500KB
			expectStreaming: false,
			expectMonitor:   false,
			expectThreshold: 10 * 1024 * 1024, // Default
		},
		{
			name:            "Medium file (5MB)",
			fileSize:        5 * 1024 * 1024,
			expectStreaming: false,
			expectMonitor:   true,
			expectThreshold: 10 * 1024 * 1024,
		},
		{
			name:            "Large file (20MB)",
			fileSize:        20 * 1024 * 1024,
			expectStreaming: true,
			expectMonitor:   true,
			expectThreshold: 5 * 1024 * 1024,
		},
		{
			name:            "Very large file (100MB)",
			fileSize:        100 * 1024 * 1024,
			expectStreaming: true,
			expectMonitor:   true,
			expectThreshold: 1 * 1024 * 1024,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp file with the specified size
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.docx")
			
			// Create a file with the specified size
			if err := createFileWithSize(tmpFile, tt.fileSize); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			opts, err := GetRecommendedOptions(tmpFile)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if opts.EnableStreaming != tt.expectStreaming {
				t.Errorf("EnableStreaming: expected %v, got %v", tt.expectStreaming, opts.EnableStreaming)
			}
			
			if opts.EnableMemoryMonitor != tt.expectMonitor {
				t.Errorf("EnableMemoryMonitor: expected %v, got %v", tt.expectMonitor, opts.EnableMemoryMonitor)
			}
			
			if opts.FileSizeThreshold != tt.expectThreshold {
				t.Errorf("FileSizeThreshold: expected %d, got %d", tt.expectThreshold, opts.FileSizeThreshold)
			}
		})
	}
}

func TestEstimateMemoryUsage(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.docx")
	
	fileSize := int64(10 * 1024 * 1024) // 10MB
	if err := createFileWithSize(tmpFile, fileSize); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	estimated, err := EstimateMemoryUsage(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Expected: fileSize * 10 (as per implementation)
	expected := uint64(fileSize) * 10
	if estimated != expected {
		t.Errorf("Expected estimated memory %d, got %d", expected, estimated)
	}
}

// Helper functions

func testContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			 testFindSubstring(s, substr)))
}

func testFindSubstring(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func createFileWithSize(path string, size int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write zeros to create a file of the specified size
	buffer := make([]byte, 1024)
	written := int64(0)
	for written < size {
		toWrite := int64(len(buffer))
		if written+toWrite > size {
			toWrite = size - written
		}
		n, err := file.Write(buffer[:toWrite])
		if err != nil {
			return err
		}
		written += int64(n)
	}
	
	return nil
}