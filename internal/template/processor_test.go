package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWordProcessor(t *testing.T) {
	t.Run("NewWordProcessor", func(t *testing.T) {
		processor := NewWordProcessor()
		if processor == nil {
			t.Fatal("NewWordProcessor returned nil")
		}
	})
	
	t.Run("ProcessTemplate", func(t *testing.T) {
		// Create a temporary Word document for testing
		tempDir, err := os.MkdirTemp("", "word_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		templatePath := filepath.Join(tempDir, "template.docx")
		outputPath := filepath.Join(tempDir, "output.docx")
		
		// Create a simple test Word document
		// We'll copy a test file instead of creating one
		testData := []byte{0x50, 0x4B} // Minimal ZIP header for .docx
		if err := os.WriteFile(templatePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		
		// Process the template
		processor := NewWordProcessor()
		values := map[string]interface{}{
			"name":  "Alice",
			"place": "Wonderland",
		}
		
		// Note: This will fail with real file validation, which is expected
		// We're just testing that the function doesn't panic
		err = processor.ProcessTemplate(templatePath, values, outputPath)
		// We expect an error since our test file is not a valid docx
		if err == nil {
			t.Error("Expected error for invalid docx file")
		}
	})
	
	t.Run("ValidateTemplate", func(t *testing.T) {
		// Create a temporary Word document for testing
		tempDir, err := os.MkdirTemp("", "word_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		templatePath := filepath.Join(tempDir, "template.docx")
		
		// Create a simple test Word document
		testData := []byte{0x50, 0x4B} // Minimal ZIP header for .docx
		if err := os.WriteFile(templatePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		
		processor := NewWordProcessor()
		values := map[string]interface{}{
			"name": "Test",
			"age":  25,
		}
		
		// Test validation - expect error for invalid file
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err == nil {
			t.Error("Expected error for invalid docx file")
		}
		_ = missing
	})
}

func TestPowerPointProcessor(t *testing.T) {
	t.Run("NewPowerPointProcessor", func(t *testing.T) {
		processor := NewPowerPointProcessor()
		if processor == nil {
			t.Fatal("NewPowerPointProcessor returned nil")
		}
	})
	
	t.Run("ProcessTemplate", func(t *testing.T) {
		// Create a temporary PowerPoint document for testing
		tempDir, err := os.MkdirTemp("", "ppt_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		templatePath := filepath.Join(tempDir, "template.pptx")
		outputPath := filepath.Join(tempDir, "output.pptx")
		
		// Create a simple test PowerPoint document
		testData := []byte{0x50, 0x4B} // Minimal ZIP header for .pptx
		if err := os.WriteFile(templatePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		
		// Process the template
		processor := NewPowerPointProcessor()
		values := map[string]interface{}{
			"title":   "Presentation",
			"content": "Sample content",
		}
		
		err = processor.ProcessTemplate(templatePath, values, outputPath)
		// We expect an error since our test file is not a valid pptx
		if err == nil {
			t.Error("Expected error for invalid pptx file")
		}
	})
	
	t.Run("ValidateTemplate", func(t *testing.T) {
		// Create a temporary PowerPoint document for testing
		tempDir, err := os.MkdirTemp("", "ppt_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		templatePath := filepath.Join(tempDir, "template.pptx")
		
		// Create a simple test PowerPoint document
		testData := []byte{0x50, 0x4B} // Minimal ZIP header for .pptx
		if err := os.WriteFile(templatePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		
		processor := NewPowerPointProcessor()
		values := map[string]interface{}{
			"title":   "Test",
			"subtitle": "Demo",
		}
		
		// Test validation - expect error for invalid file
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err == nil {
			t.Error("Expected error for invalid pptx file")
		}
		_ = missing
	})
}

