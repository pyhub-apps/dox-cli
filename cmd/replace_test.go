package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/replace"
	"gopkg.in/yaml.v3"
)

func TestReplaceCommand(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "replace_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test rules file
	rulesFile := filepath.Join(tempDir, "rules.yml")
	rules := []replace.Rule{
		{Old: "old text", New: "new text"},
		{Old: "v1.0.0", New: "v2.0.0"},
	}
	rulesData, _ := yaml.Marshal(rules)
	if err := os.WriteFile(rulesFile, rulesData, 0644); err != nil {
		t.Fatal(err)
	}

	// Test createBackup function
	t.Run("CreateBackupFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := createBackup(testFile, false); err != nil {
			t.Errorf("createBackup failed: %v", err)
		}

		// Check if backup file was created
		files, err := filepath.Glob(filepath.Join(tempDir, "test_backup_*.txt"))
		if err != nil || len(files) == 0 {
			t.Error("Backup file was not created")
		}
	})

	// Test createBackup directory function
	t.Run("CreateBackupDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		if err := os.Mkdir(testDir, 0755); err != nil {
			t.Fatal(err)
		}
		
		testFile := filepath.Join(testDir, "file.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := createBackup(testDir, true); err != nil {
			t.Errorf("createBackup for directory failed: %v", err)
		}

		// Check if backup directory was created
		dirs, err := filepath.Glob(filepath.Join(tempDir, "testdir_backup_*"))
		if err != nil || len(dirs) == 0 {
			t.Error("Backup directory was not created")
		}
	})

	// Test printResults function
	t.Run("PrintResults", func(t *testing.T) {
		results := []replace.ReplaceResult{
			{FilePath: "doc1.docx", Success: true, Replacements: 5},
			{FilePath: "doc2.docx", Success: false, Error: os.ErrNotExist},
			{FilePath: "doc3.docx", Success: true, Replacements: 3},
		}
		
		// This won't panic if printResults works correctly
		printResults(results)
	})
}

func TestPreviewDirectoryReplacements(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "preview_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test Word documents (empty files for testing)
	testFiles := []string{"doc1.docx", "doc2.docx", "test.pptx"}
	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create subdirectory with more files
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "doc3.docx"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	rules := []replace.Rule{{Old: "test", New: "test"}}

	// Test non-recursive preview
	t.Run("NonRecursivePreview", func(t *testing.T) {
		excludeGlob = "" // Reset global variable
		if err := previewDirectoryReplacements(tempDir, rules, false); err != nil {
			t.Errorf("previewDirectoryReplacements failed: %v", err)
		}
	})

	// Test recursive preview
	t.Run("RecursivePreview", func(t *testing.T) {
		excludeGlob = "" // Reset global variable
		if err := previewDirectoryReplacements(tempDir, rules, true); err != nil {
			t.Errorf("previewDirectoryReplacements failed: %v", err)
		}
	})

	// Test with exclude pattern
	t.Run("PreviewWithExclude", func(t *testing.T) {
		excludeGlob = "doc1*"
		if err := previewDirectoryReplacements(tempDir, rules, false); err != nil {
			t.Errorf("previewDirectoryReplacements with exclude failed: %v", err)
		}
		excludeGlob = "" // Reset
	})
}