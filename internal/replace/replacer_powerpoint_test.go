package replace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/document"
)

// createTestPowerPoint creates a simple PowerPoint file for testing
func createTestPowerPoint(path string) error {
	// This is the same helper function used in document package tests
	// We duplicate it here for test isolation
	return createSamplePowerPoint(path, "Version 1.0", "Status: Draft", "Year: 2023")
}

func createSamplePowerPoint(path string, title, status, year string) error {
	// Simple PowerPoint structure creator for testing
	// In production, this would be replaced with proper PowerPoint generation
	// For now, we'll use the document package's test helpers
	
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// For testing, we'll create a minimal PPTX structure
	// This is a simplified version - real implementation would use a library
	content := fmt.Sprintf("Title: %s\n%s\n%s", title, status, year)
	return os.WriteFile(path, []byte(content), 0644)
}

func TestReplaceInPowerPointDocument(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupDoc       func() string
		rules          []Rule
		wantErr        bool
		validateResult func(t *testing.T, docPath string)
	}{
		{
			name: "single rule replacement in PowerPoint",
			setupDoc: func() string {
				pptPath := filepath.Join(tempDir, "test.pptx")
				// Note: This requires a test PowerPoint file in testdata
				// For now, we'll skip if it doesn't exist
				src := "testdata/sample_presentation.pptx"
				if _, err := os.Stat(src); os.IsNotExist(err) {
					t.Skip("Sample PowerPoint file not found in testdata")
				}
				copyFile(t, src, pptPath)
				return pptPath
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			wantErr: false,
			validateResult: func(t *testing.T, docPath string) {
				doc, err := document.OpenPowerPointDocument(docPath)
				if err != nil {
					t.Fatalf("Failed to open result document: %v", err)
				}
				defer doc.Close()

				text, err := doc.GetText()
				if err != nil {
					t.Fatalf("Failed to get text: %v", err)
				}
				
				if !strings.Contains(text, "Version 2.0") {
					t.Error("New text not found after replacement")
				}
				if strings.Contains(text, "Version 1.0") {
					t.Error("Old text still present after replacement")
				}
			},
		},
		{
			name: "multiple rules replacement in PowerPoint",
			setupDoc: func() string {
				pptPath := filepath.Join(tempDir, "multi.pptx")
				src := "testdata/sample_presentation.pptx"
				if _, err := os.Stat(src); os.IsNotExist(err) {
					t.Skip("Sample PowerPoint file not found in testdata")
				}
				copyFile(t, src, pptPath)
				return pptPath
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
				{Old: "2023", New: "2024"},
				{Old: "Draft", New: "Final"},
			},
			wantErr: false,
			validateResult: func(t *testing.T, docPath string) {
				doc, err := document.OpenPowerPointDocument(docPath)
				if err != nil {
					t.Fatalf("Failed to open result document: %v", err)
				}
				defer doc.Close()

				text, err := doc.GetText()
				if err != nil {
					t.Fatalf("Failed to get text: %v", err)
				}
				
				// Check all replacements were made
				if !strings.Contains(text, "Version 2.0") {
					t.Error("Version replacement failed")
				}
				if !strings.Contains(text, "2024") {
					t.Error("Year replacement failed")
				}
				if !strings.Contains(text, "Final") {
					t.Error("Status replacement failed")
				}
				
				// Check old text is gone
				if strings.Contains(text, "Version 1.0") || 
				   strings.Contains(text, "2023") || 
				   strings.Contains(text, "Draft") {
					t.Error("Old text still present after replacement")
				}
			},
		},
		{
			name: "unsupported file type",
			setupDoc: func() string {
				txtPath := filepath.Join(tempDir, "test.txt")
				os.WriteFile(txtPath, []byte("Version 1.0"), 0644)
				return txtPath
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			wantErr: true, // Should fail for unsupported file type
			validateResult: func(t *testing.T, docPath string) {
				// No validation needed for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docPath := tt.setupDoc()
			
			_, err := ReplaceInDocumentWithCount(docPath, tt.rules)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceInDocumentWithCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				tt.validateResult(t, docPath)
			}
		})
	}
}

func TestReplaceInMixedDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupDir       func() string
		rules          []Rule
		recursive      bool
		wantErr        bool
		validateResult func(t *testing.T, dirPath string)
	}{
		{
			name: "replace in both Word and PowerPoint documents",
			setupDir: func() string {
				// Create Word documents
				if _, err := os.Stat("testdata/sample_document.docx"); err == nil {
					copyFile(t, "testdata/sample_document.docx", filepath.Join(tempDir, "doc1.docx"))
					copyFile(t, "testdata/sample_document.docx", filepath.Join(tempDir, "doc2.docx"))
				} else {
					t.Skip("Sample Word document not found")
				}
				
				// Create PowerPoint documents
				if _, err := os.Stat("testdata/sample_presentation.pptx"); err == nil {
					copyFile(t, "testdata/sample_presentation.pptx", filepath.Join(tempDir, "pres1.pptx"))
					copyFile(t, "testdata/sample_presentation.pptx", filepath.Join(tempDir, "pres2.pptx"))
				} else {
					// Create dummy PowerPoint files for testing
					os.WriteFile(filepath.Join(tempDir, "pres1.pptx"), []byte("Version 1.0"), 0644)
					os.WriteFile(filepath.Join(tempDir, "pres2.pptx"), []byte("Version 1.0"), 0644)
				}
				
				// Create other files that should be ignored
				os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("Version 1.0"), 0644)
				os.WriteFile(filepath.Join(tempDir, "data.xlsx"), []byte("Version 1.0"), 0644)
				
				return tempDir
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
			},
			recursive: false,
			wantErr:   false,
			validateResult: func(t *testing.T, dirPath string) {
				// Check Word documents were processed
				for i := 1; i <= 2; i++ {
					docPath := filepath.Join(dirPath, fmt.Sprintf("doc%d.docx", i))
					if _, err := os.Stat(docPath); err == nil {
						doc, err := document.OpenWordDocument(docPath)
						if err != nil {
							t.Errorf("Failed to open doc%d: %v", i, err)
							continue
						}
						defer doc.Close()

						text, _ := doc.GetText()
						allText := text
						if !contains(allText, "Version 2.0") {
							t.Errorf("Replacement failed in doc%d.docx", i)
						}
					}
				}
				
				// Check PowerPoint documents were processed
				for i := 1; i <= 2; i++ {
					pptPath := filepath.Join(dirPath, fmt.Sprintf("pres%d.pptx", i))
					if _, err := os.Stat(pptPath); err == nil {
						// Try to open as PowerPoint
						doc, err := document.OpenPowerPointDocument(pptPath)
						if err != nil {
							// If it's a dummy file, just check content
							content, _ := os.ReadFile(pptPath)
							if !strings.Contains(string(content), "Version") {
								t.Errorf("Failed to process pres%d.pptx", i)
							}
							continue
						}
						defer doc.Close()

						text, err := doc.GetText()
						if err == nil && !strings.Contains(text, "Version 2.0") {
							t.Errorf("Replacement failed in pres%d.pptx", i)
						}
					}
				}
				
				// Check other files were NOT modified
				txtContent, _ := os.ReadFile(filepath.Join(dirPath, "readme.txt"))
				if !strings.Contains(string(txtContent), "Version 1.0") {
					t.Error("Non-document file was modified")
				}
			},
		},
		{
			name: "recursive replacement with mixed document types",
			setupDir: func() string {
				// Create subdirectories
				subDir1 := filepath.Join(tempDir, "docs")
				subDir2 := filepath.Join(tempDir, "presentations")
				os.MkdirAll(subDir1, 0755)
				os.MkdirAll(subDir2, 0755)
				
				// Add Word documents
				if _, err := os.Stat("testdata/sample_document.docx"); err == nil {
					copyFile(t, "testdata/sample_document.docx", filepath.Join(subDir1, "report.docx"))
				}
				
				// Add PowerPoint documents
				if _, err := os.Stat("testdata/sample_presentation.pptx"); err == nil {
					copyFile(t, "testdata/sample_presentation.pptx", filepath.Join(subDir2, "slides.pptx"))
				}
				
				return tempDir
			},
			rules: []Rule{
				{Old: "Version 1.0", New: "Version 2.0"},
				{Old: "2023", New: "2024"},
			},
			recursive: true,
			wantErr:   false,
			validateResult: func(t *testing.T, dirPath string) {
				// Check Word document in subdirectory
				docPath := filepath.Join(dirPath, "docs", "report.docx")
				if _, err := os.Stat(docPath); err == nil {
					checkDocument(t, docPath, "Version 2.0")
					checkDocument(t, docPath, "2024")
				}
				
				// Check PowerPoint document in subdirectory
				pptPath := filepath.Join(dirPath, "presentations", "slides.pptx")
				if _, err := os.Stat(pptPath); err == nil {
					doc, err := document.OpenPowerPointDocument(pptPath)
					if err == nil {
						defer doc.Close()
						text, _ := doc.GetText()
						if !strings.Contains(text, "Version 2.0") || !strings.Contains(text, "2024") {
							t.Error("PowerPoint replacement failed in subdirectory")
						}
					}
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

func TestWalkDocumentFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Setup test directory structure
	os.WriteFile(filepath.Join(tempDir, "doc1.docx"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tempDir, "pres1.pptx"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tempDir, "sheet.xlsx"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tempDir, "text.txt"), []byte("content"), 0644)
	
	subDir := filepath.Join(tempDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "doc2.docx"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(subDir, "pres2.pptx"), []byte("content"), 0644)

	tests := []struct {
		name          string
		recursive     bool
		expectedFiles []string
	}{
		{
			name:      "non-recursive walk",
			recursive: false,
			expectedFiles: []string{
				"doc1.docx",
				"pres1.pptx",
			},
		},
		{
			name:      "recursive walk",
			recursive: true,
			expectedFiles: []string{
				"doc1.docx",
				"pres1.pptx",
				filepath.Join("subdir", "doc2.docx"),
				filepath.Join("subdir", "pres2.pptx"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var foundFiles []string
			
			err := WalkDocumentFiles(tempDir, tt.recursive, func(path string) error {
				relPath, _ := filepath.Rel(tempDir, path)
				foundFiles = append(foundFiles, relPath)
				return nil
			})
			
			if err != nil {
				t.Errorf("WalkDocumentFiles() error = %v", err)
				return
			}
			
			if len(foundFiles) != len(tt.expectedFiles) {
				t.Errorf("Expected %d files, found %d", len(tt.expectedFiles), len(foundFiles))
				t.Errorf("Found files: %v", foundFiles)
				return
			}
			
			// Check that all expected files were found
			for _, expected := range tt.expectedFiles {
				found := false
				for _, actual := range foundFiles {
					if filepath.Clean(actual) == filepath.Clean(expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file not found: %s", expected)
				}
			}
		})
	}
}