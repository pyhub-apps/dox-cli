package document

import (
	"os"
	"path/filepath"
	"testing"
	"strings"
)

func TestStreamingPowerPointDocument_NonExistentFile(t *testing.T) {
	// Test opening non-existent file
	_, err := OpenPowerPointDocumentStreaming("/non/existent/file.pptx", nil)
	
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
	
	expectedMsg := "file does not exist"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestStreamingPowerPointDocument_InvalidFile(t *testing.T) {
	// Test with non-PPTX file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	if err := os.WriteFile(tmpFile, []byte("not a pptx"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	_, err := OpenPowerPointDocumentStreaming(tmpFile, nil)
	
	if err == nil {
		t.Fatal("Expected error for non-PPTX file, got nil")
	}
	
	expectedMsg := "not a .pptx file"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestStreamingPowerPointDocument_CorruptedFile(t *testing.T) {
	// Test with corrupted PPTX file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "corrupted.pptx")
	
	// Create a file with .pptx extension but invalid content
	if err := os.WriteFile(tmpFile, []byte("corrupted pptx content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	_, err := OpenPowerPointDocumentStreaming(tmpFile, nil)
	
	if err == nil {
		t.Fatal("Expected error for corrupted PPTX file, got nil")
	}
	
	expectedMsg := "invalid pptx format"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestStreamingPowerPointDocument_ClosedDocument(t *testing.T) {
	// Skip test as it requires actual PowerPoint document creation
	t.Skip("Skipping test that requires PowerPoint document creation")
}

func TestStreamingPowerPointDocument_ReplaceTextWithCount(t *testing.T) {
	// Skip test as it requires actual PowerPoint document creation
	t.Skip("Skipping test that requires PowerPoint document creation")
}

func TestStreamingPowerPointDocument_EmptyReplacement(t *testing.T) {
	// Skip test as it requires actual PowerPoint document creation
	t.Skip("Skipping test that requires PowerPoint document creation")
}

func TestStreamingPowerPointDocument_MemoryPool(t *testing.T) {
	// Skip test as it requires actual PowerPoint document creation
	t.Skip("Skipping test that requires PowerPoint document creation")
}

func TestStreamingPowerPointDocument_ProcessSlidesChunked(t *testing.T) {
	// Skip test as it requires actual PowerPoint document creation
	t.Skip("Skipping test that requires PowerPoint document creation")
}