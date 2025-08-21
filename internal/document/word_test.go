package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenWordDocument(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid docx file",
			path:    "testdata/sample.docx",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "testdata/non_existent.docx",
			wantErr: true,
			errMsg:  "file does not exist",
		},
		{
			name:    "invalid file extension",
			path:    "testdata/sample.txt",
			wantErr: true,
			errMsg:  "not a .docx file",
		},
		{
			name:    "corrupted docx file",
			path:    "testdata/corrupted.docx",
			wantErr: true,
			errMsg:  "invalid docx format",
		},
		{
			name:    "valid empty docx file",
			path:    "testdata/empty.docx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := OpenWordDocument(tt.path)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenWordDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("OpenWordDocument() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
			
			if !tt.wantErr && doc == nil {
				t.Error("OpenWordDocument() returned nil document for valid file")
			}
			
			// Clean up
			if doc != nil {
				doc.Close()
			}
		})
	}
}

func TestWordDocument_GetText(t *testing.T) {
	// Test with sample document
	t.Run("sample document", func(t *testing.T) {
		doc, err := OpenWordDocument("testdata/sample.docx")
		if err != nil {
			t.Skipf("Skipping test: could not open test document: %v", err)
		}
		defer doc.Close()

		wantText := []string{
			"This is a sample document",
			"Second paragraph with some text",
			"Third paragraph",
		}

		paragraphs := doc.GetText()
		
		if len(paragraphs) != len(wantText) {
			t.Errorf("GetText() returned %d paragraphs, want %d", len(paragraphs), len(wantText))
			return
		}
		
		for i, para := range paragraphs {
			if para != wantText[i] {
				t.Errorf("GetText() paragraph[%d] = %q, want %q", i, para, wantText[i])
			}
		}
	})

	// Test with empty document
	t.Run("empty document", func(t *testing.T) {
		doc, err := OpenWordDocument("testdata/empty.docx")
		if err != nil {
			t.Skipf("Skipping test: could not open empty document: %v", err)
		}
		defer doc.Close()

		paragraphs := doc.GetText()
		
		if len(paragraphs) != 0 {
			t.Errorf("GetText() returned %d paragraphs for empty doc, want 0", len(paragraphs))
		}
	})

	// Test with unicode document
	t.Run("unicode document", func(t *testing.T) {
		doc, err := OpenWordDocument("testdata/unicode.docx")
		if err != nil {
			t.Skipf("Skipping test: could not open unicode document: %v", err)
		}
		defer doc.Close()

		wantText := []string{
			"Hello, 世界! 👋",
			"한글 텍스트 테스트",
			"Café naïve fiancé",
			"Emoji: 😃🚀🌟",
		}

		paragraphs := doc.GetText()
		
		if len(paragraphs) != len(wantText) {
			t.Errorf("GetText() returned %d paragraphs, want %d", len(paragraphs), len(wantText))
			return
		}
		
		for i, para := range paragraphs {
			if para != wantText[i] {
				t.Errorf("GetText() paragraph[%d] = %q, want %q", i, para, wantText[i])
			}
		}
	})
}

func TestWordDocument_ReplaceText(t *testing.T) {
	tests := []struct {
		name    string
		old     string
		new     string
		wantErr bool
		verify  func(t *testing.T, doc *WordDocument)
	}{
		{
			name:    "simple replacement",
			old:     "sample",
			new:     "example",
			wantErr: false,
			verify: func(t *testing.T, doc *WordDocument) {
				text := doc.GetText()
				for _, para := range text {
					if contains(para, "sample") {
						t.Errorf("ReplaceText() failed: still contains old text 'sample'")
					}
					if !contains(para, "example") && contains(para, "This is a") {
						t.Errorf("ReplaceText() failed: expected text not replaced")
					}
				}
			},
		},
		{
			name:    "replace with empty string",
			old:     "Second",
			new:     "",
			wantErr: false,
			verify: func(t *testing.T, doc *WordDocument) {
				text := doc.GetText()
				for _, para := range text {
					if contains(para, "Second") {
						t.Errorf("ReplaceText() failed: still contains old text 'Second'")
					}
				}
			},
		},
		{
			name:    "empty old text",
			old:     "",
			new:     "new",
			wantErr: true,
		},
		{
			name:    "text not found",
			old:     "nonexistent",
			new:     "replacement",
			wantErr: false, // Should succeed but make no changes
			verify: func(t *testing.T, doc *WordDocument) {
				// Original text should remain unchanged
			},
		},
		{
			name:    "XML special characters - prevent injection",
			old:     "sample",
			new:     "<w:t>INJECTED</w:t>",
			wantErr: false,
			verify: func(t *testing.T, doc *WordDocument) {
				text := doc.GetText()
				for _, para := range text {
					// The XML tags should be escaped, not interpreted
					if contains(para, "INJECTED") && !contains(para, "<w:t>") {
						t.Errorf("ReplaceText() failed: XML injection not prevented")
					}
					if contains(para, "&lt;w:t&gt;INJECTED&lt;/w:t&gt;") {
						// This is expected - XML should be escaped
						return
					}
				}
			},
		},
		{
			name:    "replace with ampersand",
			old:     "sample",
			new:     "R&D Department",
			wantErr: false,
			verify: func(t *testing.T, doc *WordDocument) {
				text := doc.GetText()
				found := false
				for _, para := range text {
					if contains(para, "R&D Department") || contains(para, "R&amp;D Department") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ReplaceText() failed: ampersand not properly handled")
				}
			},
		},
		{
			name:    "replace with quotes",
			old:     "sample",
			new:     `"quoted" text`,
			wantErr: false,
			verify: func(t *testing.T, doc *WordDocument) {
				text := doc.GetText()
				found := false
				for _, para := range text {
					if contains(para, `"quoted" text`) || contains(para, `&quot;quoted&quot; text`) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ReplaceText() failed: quotes not properly handled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh document for each test
			doc, err := OpenWordDocument("testdata/sample.docx")
			if err != nil {
				t.Skipf("Skipping test: could not open test document: %v", err)
			}
			defer doc.Close()
			
			err = doc.ReplaceText(tt.old, tt.new)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, doc)
			}
		})
	}
}

func TestWordDocument_SaveAs(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "save to valid path",
			path:    filepath.Join(tempDir, "output.docx"),
			wantErr: false,
		},
		{
			name:    "save with non-docx extension",
			path:    filepath.Join(tempDir, "output.txt"),
			wantErr: true,
		},
		{
			name:    "save to non-existent directory (auto-creates)",
			path:    filepath.Join(tempDir, "nonexistent", "dir", "output.docx"),
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Open test document
			doc, err := OpenWordDocument("testdata/sample.docx")
			if err != nil {
				t.Skipf("Skipping test: could not open test document: %v", err)
			}
			defer doc.Close()
			
			// Make a change to ensure save includes modifications
			doc.ReplaceText("sample", "modified")
			
			err = doc.SaveAs(tt.path)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveAs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Verify file was created
			if !tt.wantErr {
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("SaveAs() did not create file at %s", tt.path)
				}
				
				// Try to open the saved file to verify it's valid
				savedDoc, err := OpenWordDocument(tt.path)
				if err != nil {
					t.Errorf("SaveAs() created invalid document: %v", err)
				} else {
					savedDoc.Close()
				}
			}
		})
	}
}

func TestWordDocument_Save(t *testing.T) {
	tempDir := t.TempDir()
	
	// Copy test file to temp directory
	testFile := filepath.Join(tempDir, "test.docx")
	copyFile(t, "testdata/sample.docx", testFile)
	
	// Open the copied file
	doc, err := OpenWordDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open test document: %v", err)
	}
	defer doc.Close()
	
	// Make changes
	err = doc.ReplaceText("sample", "modified")
	if err != nil {
		t.Fatalf("Failed to replace text: %v", err)
	}
	
	// Save changes
	err = doc.Save()
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}
	
	// Re-open to verify changes were saved
	doc2, err := OpenWordDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to re-open document: %v", err)
	}
	defer doc2.Close()
	
	text := doc2.GetText()
	found := false
	for _, para := range text {
		if contains(para, "modified") {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Save() did not persist changes")
	}
}

func TestWordDocument_Close(t *testing.T) {
	doc, err := OpenWordDocument("testdata/sample.docx")
	if err != nil {
		t.Skipf("Skipping test: could not open test document: %v", err)
	}
	
	// Close should not return error
	err = doc.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	
	// Operations after close should fail gracefully
	err = doc.ReplaceText("test", "test2")
	if err == nil {
		t.Error("ReplaceText() should fail after Close()")
	}
	
	err = doc.Save()
	if err == nil {
		t.Error("Save() should fail after Close()")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	
	input, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
}