package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenPowerPointDocument(t *testing.T) {
	// Create a sample PowerPoint file for testing
	testFile := filepath.Join(t.TempDir(), "test.pptx")
	if err := createTestPowerPoint(testFile); err != nil {
		t.Fatalf("Failed to create test PowerPoint: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid PowerPoint file",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "non-existent.pptx",
			wantErr: true,
		},
		{
			name:    "invalid file",
			path:    "testdata/invalid.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := OpenPowerPointDocument(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenPowerPointDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if doc != nil {
				defer doc.Close()
			}
		})
	}
}

func TestPowerPointDocument_GetText(t *testing.T) {
	// Create a sample PowerPoint file
	testFile := filepath.Join(t.TempDir(), "test.pptx")
	if err := createTestPowerPoint(testFile); err != nil {
		t.Fatalf("Failed to create test PowerPoint: %v", err)
	}

	doc, err := OpenPowerPointDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open PowerPoint: %v", err)
	}
	defer doc.Close()

	text, err := doc.GetText()
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}

	// Check for expected content
	expectedTexts := []string{
		"Presentation Title - Version 1.0",
		"Status: Draft",
		"Year: 2023",
		"Copyright 2023 - All rights reserved",
	}

	for _, expected := range expectedTexts {
		if !strings.Contains(text, expected) {
			t.Errorf("GetText() missing expected text: %s", expected)
		}
	}
}

func TestPowerPointDocument_ReplaceText(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		new         string
		wantErr     bool
		checkResult func(t *testing.T, doc *PowerPointDocument)
	}{
		{
			name:    "simple replacement",
			old:     "Version 1.0",
			new:     "Version 2.0",
			wantErr: false,
			checkResult: func(t *testing.T, doc *PowerPointDocument) {
				text, _ := doc.GetText()
				if !strings.Contains(text, "Version 2.0") {
					t.Error("Expected 'Version 2.0' in text")
				}
				if strings.Contains(text, "Version 1.0") {
					t.Error("'Version 1.0' should have been replaced")
				}
			},
		},
		{
			name:    "replace year",
			old:     "2023",
			new:     "2024",
			wantErr: false,
			checkResult: func(t *testing.T, doc *PowerPointDocument) {
				text, _ := doc.GetText()
				if !strings.Contains(text, "2024") {
					t.Error("Expected '2024' in text")
				}
				if strings.Contains(text, "2023") {
					t.Error("'2023' should have been replaced")
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
			name:    "XML special characters",
			old:     "Draft",
			new:     "<strong>Final</strong>",
			wantErr: false,
			checkResult: func(t *testing.T, doc *PowerPointDocument) {
				text, _ := doc.GetText()
				// The XML should be escaped, so we should see the escaped version
				if !strings.Contains(text, "Final") {
					t.Error("Expected 'Final' in text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh test file for each test
			testFile := filepath.Join(t.TempDir(), "test.pptx")
			if err := createTestPowerPoint(testFile); err != nil {
				t.Fatalf("Failed to create test PowerPoint: %v", err)
			}

			doc, err := OpenPowerPointDocument(testFile)
			if err != nil {
				t.Fatalf("Failed to open PowerPoint: %v", err)
			}
			defer doc.Close()

			err = doc.ReplaceText(tt.old, tt.new)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, doc)
			}
		})
	}
}

func TestPowerPointDocument_Save(t *testing.T) {
	// Create a test PowerPoint file
	testFile := filepath.Join(t.TempDir(), "test.pptx")
	if err := createTestPowerPoint(testFile); err != nil {
		t.Fatalf("Failed to create test PowerPoint: %v", err)
	}

	doc, err := OpenPowerPointDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open PowerPoint: %v", err)
	}
	defer doc.Close()

	// Make a change
	if err := doc.ReplaceText("Version 1.0", "Version 2.0"); err != nil {
		t.Fatalf("Failed to replace text: %v", err)
	}

	// Save the document
	if err := doc.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Open the saved document and verify changes
	doc2, err := OpenPowerPointDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved PowerPoint: %v", err)
	}
	defer doc2.Close()

	text, err := doc2.GetText()
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}

	if !strings.Contains(text, "Version 2.0") {
		t.Error("Expected 'Version 2.0' in saved document")
	}
	if strings.Contains(text, "Version 1.0") {
		t.Error("'Version 1.0' should have been replaced in saved document")
	}
}

func TestPowerPointDocument_SaveAs(t *testing.T) {
	// Create a test PowerPoint file
	testFile := filepath.Join(t.TempDir(), "test.pptx")
	if err := createTestPowerPoint(testFile); err != nil {
		t.Fatalf("Failed to create test PowerPoint: %v", err)
	}

	doc, err := OpenPowerPointDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open PowerPoint: %v", err)
	}
	defer doc.Close()

	// Make a change
	if err := doc.ReplaceText("Draft", "Final"); err != nil {
		t.Fatalf("Failed to replace text: %v", err)
	}

	// Save to a new file
	newFile := filepath.Join(t.TempDir(), "subfolder", "new.pptx")
	if err := doc.SaveAs(newFile); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	// Verify the new file exists and has the changes
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("SaveAs() did not create the new file")
	}

	// Open the new file and verify changes
	doc2, err := OpenPowerPointDocument(newFile)
	if err != nil {
		t.Fatalf("Failed to open new PowerPoint: %v", err)
	}
	defer doc2.Close()

	text, err := doc2.GetText()
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}

	if !strings.Contains(text, "Final") {
		t.Error("Expected 'Final' in new document")
	}
	if strings.Contains(text, "Draft") {
		t.Error("'Draft' should have been replaced in new document")
	}

	// Verify original file is unchanged
	doc3, err := OpenPowerPointDocument(testFile)
	if err != nil {
		t.Fatalf("Failed to open original PowerPoint: %v", err)
	}
	defer doc3.Close()

	originalText, err := doc3.GetText()
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}

	// Original should still have "Draft" since we didn't save to it
	if !strings.Contains(originalText, "Draft") {
		t.Error("Original file should still contain 'Draft'")
	}
}