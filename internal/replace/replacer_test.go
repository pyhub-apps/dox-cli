package replace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/document"
)

func TestReplaceInDocument(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupDoc       func() string // Returns path to test document
		rules          []Rule
		wantErr        bool
		validateResult func(t *testing.T, docPath string)
	}{
		{
			name: "single rule replacement",
			setupDoc: func() string {
				// Copy sample document to temp dir
				src := "testdata/sample_document.docx"
				dst := filepath.Join(tempDir, "single_rule.docx")
				copyFile(t, src, dst)
				return dst
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			wantErr: false,
			validateResult: func(t *testing.T, docPath string) {
				doc, err := document.OpenWordDocument(docPath)
				if err != nil {
					t.Fatalf("Failed to open result document: %v", err)
				}
				defer doc.Close()

				text, _ := doc.GetText()
				found := false
				for _, para := range strings.Split(text, "\n") {
					if contains(para, "Version 2.0") {
						found = true
						break
					}
					if contains(para, "Version 1.0") {
						t.Error("Old text still present after replacement")
					}
				}
				if !found {
					t.Error("New text not found after replacement")
				}
			},
		},
		{
			name: "multiple rules replacement",
			setupDoc: func() string {
				src := "testdata/sample_document.docx"
				dst := filepath.Join(tempDir, "multiple_rules.docx")
				copyFile(t, src, dst)
				return dst
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
				{Old: "2023", New: "2024"},
				{Old: "Draft", New: "Final"},
			},
			wantErr: false,
			validateResult: func(t *testing.T, docPath string) {
				doc, err := document.OpenWordDocument(docPath)
				if err != nil {
					t.Fatalf("Failed to open result document: %v", err)
				}
				defer doc.Close()

				text, _ := doc.GetText()
				allText := text
				
				// Check all replacements were made
				if !contains(allText, "Version 2.0") {
					t.Error("Version replacement failed")
				}
				if !contains(allText, "2024") {
					t.Error("Year replacement failed")
				}
				if !contains(allText, "Final") {
					t.Error("Status replacement failed")
				}
				
				// Check old text is gone
				if contains(allText, "Version 1.0") || contains(allText, "2023") || contains(allText, "Draft") {
					t.Error("Old text still present after replacement")
				}
			},
		},
		{
			name: "invalid document path",
			setupDoc: func() string {
				return filepath.Join(tempDir, "nonexistent.docx")
			},
			rules: []Rule{
				{Old: "test", New: "replacement"},
			},
			wantErr: true,
			validateResult: func(t *testing.T, docPath string) {
				// No validation needed for error case
			},
		},
		{
			name: "empty rules",
			setupDoc: func() string {
				src := "testdata/sample_document.docx"
				dst := filepath.Join(tempDir, "empty_rules.docx")
				copyFile(t, src, dst)
				return dst
			},
			rules:   []Rule{},
			wantErr: false, // Should succeed but make no changes
			validateResult: func(t *testing.T, docPath string) {
				// Document should remain unchanged
				// We'll verify by checking that original content is still there
				doc, err := document.OpenWordDocument(docPath)
				if err != nil {
					t.Fatalf("Failed to open result document: %v", err)
				}
				defer doc.Close()

				text, _ := doc.GetText()
				if text == "" {
					t.Error("Document content was lost")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docPath := tt.setupDoc()
			
			err := ReplaceInDocument(docPath, tt.rules)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceInDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				tt.validateResult(t, docPath)
			}
		})
	}
}

func TestReplaceInDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupDir       func() string // Returns directory with test documents
		rules          []Rule
		recursive      bool
		wantErr        bool
		validateResult func(t *testing.T, dirPath string)
	}{
		{
			name: "replace in all documents in directory",
			setupDir: func() string {
				// Create multiple test documents
				for i := 1; i <= 3; i++ {
					src := "testdata/sample_document.docx"
					dst := filepath.Join(tempDir, fmt.Sprintf("doc%d.docx", i))
					copyFile(t, src, dst)
				}
				return tempDir
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			recursive: false,
			wantErr:   false,
			validateResult: func(t *testing.T, dirPath string) {
				// Check all documents were processed
				for i := 1; i <= 3; i++ {
					docPath := filepath.Join(dirPath, fmt.Sprintf("doc%d.docx", i))
					doc, err := document.OpenWordDocument(docPath)
					if err != nil {
						t.Errorf("Failed to open doc%d: %v", i, err)
						continue
					}
					defer doc.Close()

					text, _ := doc.GetText()
					allText := text
					if !contains(allText, "Version 2.0") {
						t.Errorf("Replacement failed in doc%d", i)
					}
				}
			},
		},
		{
			name: "recursive replacement in subdirectories",
			setupDir: func() string {
				// Create documents in subdirectories
				subDir := filepath.Join(tempDir, "subdir")
				os.MkdirAll(subDir, 0755)
				
				copyFile(t, "testdata/sample_document.docx", filepath.Join(tempDir, "root.docx"))
				copyFile(t, "testdata/sample_document.docx", filepath.Join(subDir, "sub.docx"))
				
				return tempDir
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			recursive: true,
			wantErr:   false,
			validateResult: func(t *testing.T, dirPath string) {
				// Check root document
				checkDocument(t, filepath.Join(dirPath, "root.docx"), "Version 2.0")
				
				// Check subdirectory document
				checkDocument(t, filepath.Join(dirPath, "subdir", "sub.docx"), "Version 2.0")
			},
		},
		{
			name: "skip non-docx files",
			setupDir: func() string {
				// Create mixed file types
				copyFile(t, "testdata/sample_document.docx", filepath.Join(tempDir, "doc.docx"))
				os.WriteFile(filepath.Join(tempDir, "text.txt"), []byte("Version 1.0"), 0644)
				os.WriteFile(filepath.Join(tempDir, "data.xml"), []byte("Version 1.0"), 0644)
				
				return tempDir
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			recursive: false,
			wantErr:   false,
			validateResult: func(t *testing.T, dirPath string) {
				// Check docx was processed
				checkDocument(t, filepath.Join(dirPath, "doc.docx"), "Version 2.0")
				
				// Check other files were not modified
				txtContent, _ := os.ReadFile(filepath.Join(dirPath, "text.txt"))
				if !contains(string(txtContent), "Version 1.0") {
					t.Error("Non-docx file was modified")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath := tt.setupDir()
			
			err := ReplaceInDirectory(dirPath, tt.rules, tt.recursive)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceInDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				tt.validateResult(t, dirPath)
			}
		})
	}
}

// Helper functions

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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func checkDocument(t *testing.T, path string, expectedText string) {
	t.Helper()
	
	doc, err := document.OpenWordDocument(path)
	if err != nil {
		t.Errorf("Failed to open %s: %v", path, err)
		return
	}
	defer doc.Close()
	
	text, _ := doc.GetText()
	allText := text
	
	if !contains(allText, expectedText) {
		t.Errorf("Expected text '%s' not found in %s", expectedText, path)
	}
}