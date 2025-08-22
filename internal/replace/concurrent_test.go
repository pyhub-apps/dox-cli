package replace

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConcurrentProcessing(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "concurrent_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tempDir, "test.txt")
		if i > 0 {
			filename = filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		}
		err := os.WriteFile(filename, []byte("test content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create rules
	rules := []Rule{
		{Old: "test", New: "replaced"},
		{Old: "content", New: "text"},
	}

	t.Run("ConcurrentProcessing", func(t *testing.T) {
		opts := ConcurrentOptions{
			MaxWorkers:   2,
			ShowProgress: false,
		}

		results, err := ReplaceInDirectoryConcurrent(tempDir, rules, false, "", opts)
		if err != nil {
			t.Errorf("Concurrent processing failed: %v", err)
		}

		// We created 5 text files, but the function only processes .docx and .pptx
		// So we expect 0 results
		if len(results) != 0 {
			t.Errorf("Expected 0 results for .txt files, got %d", len(results))
		}
	})

	// Create some .docx files (mock)
	docxFile := filepath.Join(tempDir, "test.docx")
	err = os.WriteFile(docxFile, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ConcurrentWithDocx", func(t *testing.T) {
		opts := ConcurrentOptions{
			MaxWorkers:   1,
			ShowProgress: false,
		}

		results, err := ReplaceInDirectoryConcurrent(tempDir, rules, false, "", opts)
		// This will fail because it's not a real docx, but we should get a result
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}

		// The processing should fail since it's not a real docx
		if results[0].Success {
			t.Error("Expected failure for invalid docx file")
		}
	})
}

func TestDefaultConcurrentOptions(t *testing.T) {
	opts := DefaultConcurrentOptions()
	
	if opts.MaxWorkers <= 0 {
		t.Errorf("MaxWorkers should be positive, got %d", opts.MaxWorkers)
	}
	
	if opts.ShowProgress {
		t.Error("ShowProgress should be false by default")
	}
}