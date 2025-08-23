package secrets

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string    // Strings that should be present
		notExpected []string // Strings that should NOT be present
	}{
		{
			name:     "Config with API key",
			input:    "Loading config with api_key=sk-proj-1234567890abcdef",
			expected: []string{"Loading config with api_key="},
			notExpected: []string{"1234567890abcdef"},
		},
		{
			name:     "JSON with password",
			input:    `{"username":"admin","password":"secretpass123"}`,
			expected: []string{`"username":"admin"`},
			notExpected: []string{"secretpass123"},
		},
		{
			name:     "Environment variables",
			input:    "OPENAI_API_KEY=sk-proj-abc123 ANTHROPIC_API_KEY=sk-ant-def456",
			expected: []string{"OPENAI_API_KEY=", "ANTHROPIC_API_KEY="},
			notExpected: []string{"abc123", "def456"},
		},
		{
			name:     "Custom sensitive field",
			input:    "custom_secret=mysecretvalue",
			expected: []string{"custom_secret="},
			notExpected: []string{"mysecretvalue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input, "custom_secret")
			
			// Check expected strings
			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Result should contain %q, got: %s", exp, result)
				}
			}
			
			// Check not expected strings
			for _, notExp := range tt.notExpected {
				if strings.Contains(result, notExp) {
					t.Errorf("Result should NOT contain %q, got: %s", notExp, result)
				}
			}
		})
	}
}

func TestGetSecureConfigPath(t *testing.T) {
	path, err := GetSecureConfigPath()
	if err != nil {
		t.Fatalf("GetSecureConfigPath() error = %v", err)
	}
	
	// Check path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Config path should be created: %s", path)
	}
	
	// Check path contains expected directory name
	if !strings.Contains(path, "pyhub-docs") {
		t.Errorf("Config path should contain 'pyhub-docs': %s", path)
	}
	
	// OS-specific checks
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(path, "Library/Application Support") {
			t.Errorf("macOS path should contain 'Library/Application Support': %s", path)
		}
	case "windows":
		if !strings.Contains(path, "AppData") && !strings.Contains(path, "pyhub-docs") {
			t.Errorf("Windows path should contain 'AppData' or 'pyhub-docs': %s", path)
		}
	default:
		if !strings.Contains(path, ".config") {
			t.Errorf("Linux path should contain '.config': %s", path)
		}
	}
}

func TestCheckConfigFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission tests on Windows")
	}
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pyhub-docs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test with secure permissions
	securePath := filepath.Join(tempDir, "secure.yml")
	if err := os.WriteFile(securePath, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	
	if err := CheckConfigFilePermissions(securePath); err != nil {
		t.Errorf("CheckConfigFilePermissions() with 0600 should not error: %v", err)
	}
	
	// Test with insecure permissions (world-writable)
	insecurePath := filepath.Join(tempDir, "insecure.yml")
	if err := os.WriteFile(insecurePath, []byte("test"), 0666); err != nil {
		t.Fatal(err)
	}
	
	// Verify the file actually has world-writable permissions
	info, err := os.Stat(insecurePath)
	if err != nil {
		t.Fatal(err)
	}
	
	// Only test if we were able to set world-writable permissions
	// (some systems or filesystems may not allow this)
	if info.Mode().Perm()&0002 != 0 {
		if err := CheckConfigFilePermissions(insecurePath); err == nil {
			t.Error("CheckConfigFilePermissions() with world-writable permissions should error")
		}
	} else {
		t.Skip("System doesn't allow world-writable permissions, skipping test")
	}
	
	// Test with non-existent file (should not error)
	nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")
	if err := CheckConfigFilePermissions(nonExistentPath); err != nil {
		t.Errorf("CheckConfigFilePermissions() with non-existent file should not error: %v", err)
	}
}

func TestSetSecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix permission tests on Windows")
	}
	
	// Create temp file
	tempFile, err := os.CreateTemp("", "pyhub-docs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Set insecure permissions first
	if err := os.Chmod(tempFile.Name(), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Apply secure permissions
	if err := SetSecurePermissions(tempFile.Name()); err != nil {
		t.Fatalf("SetSecurePermissions() error = %v", err)
	}
	
	// Check permissions
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("File permissions = %v, want 0600", mode)
	}
}

func TestEnsureSecureDirectory(t *testing.T) {
	// Create temp base directory
	tempDir, err := os.MkdirTemp("", "pyhub-docs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test creating new directory
	newDir := filepath.Join(tempDir, "config")
	if err := EnsureSecureDirectory(newDir); err != nil {
		t.Fatalf("EnsureSecureDirectory() error = %v", err)
	}
	
	// Check directory exists
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if !info.IsDir() {
		t.Error("Should create a directory")
	}
	
	// Check permissions on Unix-like systems
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode != 0700 {
			t.Errorf("Directory permissions = %v, want 0700", mode)
		}
	}
	
	// Test with existing directory
	if err := EnsureSecureDirectory(newDir); err != nil {
		t.Errorf("EnsureSecureDirectory() on existing directory error = %v", err)
	}
}