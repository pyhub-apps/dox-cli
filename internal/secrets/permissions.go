package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FilePermissions represents file permission requirements
type FilePermissions struct {
	Path string
	Mode os.FileMode
}

// CheckConfigFilePermissions checks if config files have secure permissions
func CheckConfigFilePermissions(configPath string) error {
	if configPath == "" {
		return nil
	}
	
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's okay
			return nil
		}
		return fmt.Errorf("failed to check config file: %w", err)
	}
	
	// Get file permissions
	mode := info.Mode().Perm()
	
	// On Unix-like systems, check for secure permissions (600 or 644)
	if runtime.GOOS != "windows" {
		// Check if others have write permission (insecure)
		if mode&0002 != 0 {
			return fmt.Errorf("config file %s has insecure permissions %v (world-writable), should be 600 or 644", configPath, mode)
		}
		
		// Check if group has write permission (potentially insecure)
		if mode&0020 != 0 {
			// This is a warning, not an error
			fmt.Fprintf(os.Stderr, "Warning: config file %s has group-writable permissions %v, consider using 600\n", configPath, mode)
		}
		
		// Ideal permission is 0600 (owner read/write only)
		if mode != 0600 && mode != 0644 {
			fmt.Fprintf(os.Stderr, "Note: config file %s has permissions %v, recommended is 600 for maximum security\n", configPath, mode)
		}
	}
	
	return nil
}

// SetSecurePermissions sets secure permissions on a file (600 on Unix-like systems)
func SetSecurePermissions(path string) error {
	if runtime.GOOS == "windows" {
		// Windows handles permissions differently
		// File permissions are managed through ACLs
		return nil
	}
	
	// Set permission to 0600 (owner read/write only)
	err := os.Chmod(path, 0600)
	if err != nil {
		return fmt.Errorf("failed to set secure permissions on %s: %w", path, err)
	}
	
	return nil
}

// EnsureSecureDirectory ensures a directory exists with secure permissions
func EnsureSecureDirectory(dir string) error {
	// Create directory if it doesn't exist
	err := os.MkdirAll(dir, 0700) // Owner read/write/execute only
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// On Unix-like systems, verify permissions
	if runtime.GOOS != "windows" {
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("failed to check directory %s: %w", dir, err)
		}
		
		mode := info.Mode().Perm()
		
		// Check if others have any permission (insecure for config directory)
		if mode&0007 != 0 {
			// Try to fix permissions
			if err := os.Chmod(dir, 0700); err != nil {
				return fmt.Errorf("directory %s has insecure permissions %v and cannot be fixed: %w", dir, mode, err)
			}
			fmt.Fprintf(os.Stderr, "Fixed insecure permissions on directory %s\n", dir)
		}
	}
	
	return nil
}

// GetSecureConfigPath returns the secure configuration directory path
func GetSecureConfigPath() (string, error) {
	var configDir string
	
	// Determine config directory based on OS
	switch runtime.GOOS {
	case "windows":
		// Use APPDATA on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "pyhub-docs")
		
	case "darwin":
		// Use ~/Library/Application Support on macOS
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "Library", "Application Support", "pyhub-docs")
		
	default:
		// Use XDG_CONFIG_HOME or ~/.config on Linux and others
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(xdgConfig, "pyhub-docs")
	}
	
	// Ensure directory exists with secure permissions
	if err := EnsureSecureDirectory(configDir); err != nil {
		return "", err
	}
	
	return configDir, nil
}

// SanitizeForLogging removes sensitive information from a string for safe logging
func SanitizeForLogging(input string, sensitiveFields ...string) string {
	result := input
	
	// Default sensitive field names
	defaultSensitive := []string{
		"api_key", "apikey", "api-key",
		"password", "passwd", "pwd",
		"secret", "token", "auth",
		"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "CLAUDE_API_KEY",
	}
	
	// Combine with provided sensitive fields
	allSensitive := append(defaultSensitive, sensitiveFields...)
	
	// Replace sensitive values with masked versions
	for _, field := range allSensitive {
		// Look for patterns like: field=value, field:value, "field":"value"
		// This is a simple implementation; a more robust one would use regex
		if idx := findFieldValue(result, field); idx >= 0 {
			// Mask the value after the field
			result = maskFieldValue(result, idx)
		}
	}
	
	return result
}

// findFieldValue finds the position of a field value in a string
func findFieldValue(s, field string) int {
	// Simple implementation - can be enhanced with regex
	patterns := []string{
		field + "=",
		field + ":",
		`"` + field + `":"`,
		`'` + field + `':'`,
	}
	
	for _, pattern := range patterns {
		if idx := stringIndex(s, pattern); idx >= 0 {
			return idx + len(pattern)
		}
	}
	
	return -1
}

// maskFieldValue masks the value starting at the given index
func maskFieldValue(s string, startIdx int) string {
	if startIdx >= len(s) {
		return s
	}
	
	// Find the end of the value (space, comma, newline, or end of string)
	endIdx := startIdx
	inQuotes := false
	quoteChar := byte(0)
	
	// Check if value starts with quotes
	if s[startIdx] == '"' || s[startIdx] == '\'' {
		inQuotes = true
		quoteChar = s[startIdx]
		startIdx++
		endIdx = startIdx
	}
	
	for endIdx < len(s) {
		if inQuotes {
			if s[endIdx] == quoteChar && (endIdx == 0 || s[endIdx-1] != '\\') {
				break
			}
		} else {
			if s[endIdx] == ' ' || s[endIdx] == ',' || s[endIdx] == '\n' || s[endIdx] == '\r' {
				break
			}
		}
		endIdx++
	}
	
	// Extract the value
	value := s[startIdx:endIdx]
	
	// Mask the value
	masked := MaskAPIKey(value)
	
	// Reconstruct the string
	result := s[:startIdx] + masked
	if endIdx < len(s) {
		result += s[endIdx:]
	}
	
	return result
}

// stringIndex is a simple case-insensitive string search
func stringIndex(s, substr string) int {
	// Convert both to lowercase for case-insensitive search
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	
	idx := strings.Index(sLower, substrLower)
	return idx
}