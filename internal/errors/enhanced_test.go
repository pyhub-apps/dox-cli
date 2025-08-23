package errors

import (
	"strings"
	"testing"
)

func TestEnhancedError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedCode    ErrorCode
		expectedInError []string
		notInError      []string
	}{
		{
			name: "FileNotFoundError",
			err: FileNotFoundError("/path/to/file.txt"),
			expectedCode: ErrCodeFileNotFound,
			expectedInError: []string{
				"File not found",
				"/path/to/file.txt",
				"Check if the file exists",
				"absolute path",
			},
			notInError: []string{},
		},
		{
			name: "PermissionDeniedError",
			err: PermissionDeniedError("/etc/passwd", "write"),
			expectedCode: ErrCodePermissionDenied,
			expectedInError: []string{
				"Permission denied",
				"Cannot write",
				"/etc/passwd",
				"sudo",
				"permissions",
			},
			notInError: []string{},
		},
		{
			name: "InvalidYAMLError",
			err: InvalidYAMLError("config.yml", 42, nil),
			expectedCode: ErrCodeInvalidYAML,
			expectedInError: []string{
				"Invalid YAML",
				"config.yml",
				"line 42",
				"spaces for indentation",
				"yamllint.com",
			},
			notInError: []string{},
		},
		{
			name: "MissingAPIKeyError",
			err: MissingAPIKeyError("OpenAI"),
			expectedCode: ErrCodeMissingAPIKey,
			expectedInError: []string{
				"OpenAI API key",
				"not configured",
				"OPENAI_API_KEY",
				"export",
				"--api-key",
			},
			notInError: []string{},
		},
		{
			name: "OutOfMemoryError",
			err: OutOfMemoryError(100 * 1024 * 1024), // 100MB
			expectedCode: ErrCodeOutOfMemory,
			expectedInError: []string{
				"Not enough memory",
				"100.0 MB", // Now using float formatting
				"--streaming",
				"Close other applications",
			},
			notInError: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check error code
			if code := GetErrorCode(tt.err); code != tt.expectedCode {
				t.Errorf("GetErrorCode() = %v, want %v", code, tt.expectedCode)
			}

			// Check if error is enhanced
			if !IsEnhancedError(tt.err) {
				t.Error("Expected error to be an EnhancedError")
			}

			// Check error message content
			errStr := tt.err.Error()
			for _, expected := range tt.expectedInError {
				if !strings.Contains(errStr, expected) {
					t.Errorf("Error message should contain %q, got: %s", expected, errStr)
				}
			}

			for _, notExpected := range tt.notInError {
				if strings.Contains(errStr, notExpected) {
					t.Errorf("Error message should not contain %q, got: %s", notExpected, errStr)
				}
			}
		})
	}
}

func TestErrorBuilder(t *testing.T) {
	err := NewError(ErrCodeFileNotFound, "Test error").
		WithDetails("Additional details").
		WithSuggestion("Try this").
		WithSuggestion("Or try that").
		WithContext("file", "test.txt").
		WithContext("line", 42).
		Build()

	errStr := err.Error()

	expectedParts := []string{
		"Test error",
		"[DOX100]", // Using the actual error code format
		"Additional details",
		"Try this",
		"Or try that",
		"file: test.txt",
		"line: 42",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errStr, part) {
			t.Errorf("Error should contain %q, got: %s", part, errStr)
		}
	}
}

func TestEnhancedErrorUnwrap(t *testing.T) {
	originalErr := FileNotFoundError("test.txt")
	wrappedErr := NewError(ErrCodePermissionDenied, "Access denied").
		WithWrapped(originalErr).
		Build()

	// Check that we can unwrap to get the original error
	var enhancedErr *EnhancedError
	if !As(wrappedErr, &enhancedErr) {
		t.Fatal("Expected to extract EnhancedError")
	}

	if enhancedErr.Wrapped == nil {
		t.Error("Expected wrapped error to be present")
	}
}

func TestGetSuggestions(t *testing.T) {
	// Test with suggestions
	err := FileNotFoundError("/missing/file.txt")
	
	var enhanced *EnhancedError
	if !As(err, &enhanced) {
		t.Fatal("Expected EnhancedError")
	}

	suggestions := enhanced.GetSuggestions()
	if len(suggestions) < 2 {
		t.Errorf("Expected at least 2 suggestions, got %d", len(suggestions))
	}
	
	// Test empty suggestions
	emptyErr := &EnhancedError{
		Code:        ErrCodeFileNotFound,
		Message:     "test error",
		Suggestions: nil,
	}
	
	emptySuggestions := emptyErr.GetSuggestions()
	if len(emptySuggestions) != 0 {
		t.Errorf("Expected empty suggestions, got %d", len(emptySuggestions))
	}
}

func TestDuplicateSuggestions(t *testing.T) {
	err := NewError(ErrCodeFileNotFound, "Test").
		WithSuggestion("Try this").
		WithSuggestion("Try this"). // Duplicate
		WithSuggestion("Try that").
		Build()
	
	var enhanced *EnhancedError
	if !As(err, &enhanced) {
		t.Fatal("Expected EnhancedError")
	}
	
	if len(enhanced.Suggestions) != 2 {
		t.Errorf("Expected 2 unique suggestions, got %d", len(enhanced.Suggestions))
	}
}

func TestNilContextValues(t *testing.T) {
	err := NewError(ErrCodeFileNotFound, "Test").
		WithContext("key1", "value1").
		WithContext("key2", nil). // Should be skipped
		WithContext("key3", "").   // Should be skipped
		WithContext("key4", 0).    // Should be included
		Build()
	
	var enhanced *EnhancedError
	if !As(err, &enhanced) {
		t.Fatal("Expected EnhancedError")
	}
	
	if len(enhanced.Context) != 2 {
		t.Errorf("Expected 2 context values, got %d", len(enhanced.Context))
	}
	
	if _, exists := enhanced.Context["key2"]; exists {
		t.Error("Nil value should not be in context")
	}
	
	if _, exists := enhanced.Context["key3"]; exists {
		t.Error("Empty string should not be in context")
	}
	
	if _, exists := enhanced.Context["key4"]; !exists {
		t.Error("Zero value should be in context")
	}
}