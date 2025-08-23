package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected string
	}{
		{
			name:     "API key not found",
			code:     ErrCodeAPIKeyNotFound,
			expected: "DOX001",
		},
		{
			name:     "File not found",
			code:     ErrCodeFileNotFound,
			expected: "DOX100",
		},
		{
			name:     "Document corrupted",
			code:     ErrCodeDocumentCorrupted,
			expected: "DOX200",
		},
		{
			name:     "AI request failed",
			code:     ErrCodeAIRequestFailed,
			expected: "DOX300",
		},
		{
			name:     "Invalid input",
			code:     ErrCodeInvalidInput,
			expected: "DOX400",
		},
		{
			name:     "Network timeout",
			code:     ErrCodeNetworkTimeout,
			expected: "DOX500",
		},
		{
			name:     "Internal error",
			code:     ErrCodeInternalError,
			expected: "DOX900",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code) != tt.expected {
				t.Errorf("ErrorCode = %v, want %v", tt.code, tt.expected)
			}
		})
	}
}

func TestCodedError(t *testing.T) {
	tests := []struct {
		name           string
		err            *CodedError
		expectedCode   ErrorCode
		expectedLevel  ErrorLevel
		containsInMsg  []string
		hasContext     bool
		hasSolution    bool
	}{
		{
			name: "Basic error",
			err: NewCodedError(
				ErrCodeFileNotFound,
				LevelError,
				"File not found",
				"Check the file path",
				nil,
			),
			expectedCode:  ErrCodeFileNotFound,
			expectedLevel: LevelError,
			// Check for either localized key or actual error code/message
			containsInMsg: []string{"file_not_found", "Check the file path"},
			hasSolution:   true,
		},
		{
			name: "Error with cause",
			err: NewCodedError(
				ErrCodeInternalError,
				LevelError,
				"Operation failed",
				"",
				errors.New("underlying error"),
			),
			expectedCode:  ErrCodeInternalError,
			expectedLevel: LevelError,
			// Check for either localized key or actual error code/message
			containsInMsg: []string{"internal_error", "underlying error"},
			hasSolution:   false,
		},
		{
			name: "Warning with context",
			err: NewCodedError(
				ErrCodeAIRateLimited,
				LevelWarning,
				"Rate limit exceeded",
				"Wait and retry",
				nil,
			).WithContext("provider", "OpenAI").WithContext("retry_after", "60s"),
			expectedCode:  ErrCodeAIRateLimited,
			expectedLevel: LevelWarning,
			// Check for either localized key or actual error code/message
			containsInMsg: []string{"rate_limited", "OpenAI"},
			hasContext:    true,
			hasSolution:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test GetCode
			if got := tt.err.GetCode(); got != tt.expectedCode {
				t.Errorf("GetCode() = %v, want %v", got, tt.expectedCode)
			}

			// Test GetLevel
			if got := tt.err.GetLevel(); got != tt.expectedLevel {
				t.Errorf("GetLevel() = %v, want %v", got, tt.expectedLevel)
			}

			// Test Error message contains expected strings
			errMsg := tt.err.Error()
			for _, expected := range tt.containsInMsg {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("Error() message does not contain %q, got: %v", expected, errMsg)
				}
			}

			// Test context presence
			if tt.hasContext && len(tt.err.Context) == 0 {
				t.Error("Expected context to be present")
			}

			// Test solution presence
			if tt.hasSolution && tt.err.Solution == "" {
				t.Error("Expected solution to be present")
			}

			// Test Unwrap
			if tt.err.Cause != nil {
				if unwrapped := tt.err.Unwrap(); unwrapped != tt.err.Cause {
					t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err.Cause)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("NewAPIKeyNotFoundError", func(t *testing.T) {
		err := NewAPIKeyNotFoundError("OpenAI")
		
		if err.Code != ErrCodeAPIKeyNotFound {
			t.Errorf("Code = %v, want %v", err.Code, ErrCodeAPIKeyNotFound)
		}
		
		if err.Level != LevelError {
			t.Errorf("Level = %v, want %v", err.Level, LevelError)
		}
		
		if !strings.Contains(err.Message, "OpenAI") {
			t.Errorf("Message does not contain provider name")
		}
		
		if err.Solution == "" {
			t.Error("Solution should be present")
		}
		
		// Check context
		if provider, ok := err.Context["Provider"]; !ok || provider != "OpenAI" {
			t.Error("Context should contain Provider")
		}
	})

	t.Run("NewFileNotFoundError", func(t *testing.T) {
		err := NewFileNotFoundError("/path/to/file.txt")
		
		if err.Code != ErrCodeFileNotFound {
			t.Errorf("Code = %v, want %v", err.Code, ErrCodeFileNotFound)
		}
		
		if !strings.Contains(err.Message, "/path/to/file.txt") {
			t.Errorf("Message does not contain file path")
		}
		
		// Check context
		if path, ok := err.Context["Path"]; !ok || path != "/path/to/file.txt" {
			t.Error("Context should contain Path")
		}
	})

	t.Run("NewInvalidFormatError", func(t *testing.T) {
		err := NewInvalidFormatError("txt", "docx|pptx")
		
		if err.Code != ErrCodeInvalidFormat {
			t.Errorf("Code = %v, want %v", err.Code, ErrCodeInvalidFormat)
		}
		
		// Check context
		if format, ok := err.Context["Format"]; !ok || format != "txt" {
			t.Error("Context should contain Format")
		}
		
		if expected, ok := err.Context["Expected"]; !ok || expected != "docx|pptx" {
			t.Error("Context should contain Expected")
		}
	})

	t.Run("NewPermissionDeniedError", func(t *testing.T) {
		err := NewPermissionDeniedError("/protected/file")
		
		if err.Code != ErrCodePermissionDenied {
			t.Errorf("Code = %v, want %v", err.Code, ErrCodePermissionDenied)
		}
		
		if !strings.Contains(err.Message, "/protected/file") {
			t.Errorf("Message does not contain file path")
		}
	})

	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError("Claude", "60s")
		
		if err.Code != ErrCodeAIRateLimited {
			t.Errorf("Code = %v, want %v", err.Code, ErrCodeAIRateLimited)
		}
		
		if err.Level != LevelWarning {
			t.Errorf("Level = %v, want %v", err.Level, LevelWarning)
		}
		
		if !strings.Contains(err.Message, "Claude") {
			t.Errorf("Message does not contain provider name")
		}
	})
}

func TestErrorChecking(t *testing.T) {
	t.Run("IsCodedError", func(t *testing.T) {
		codedErr := NewCodedError(ErrCodeFileNotFound, LevelError, "test", "", nil)
		regularErr := errors.New("regular error")
		
		if !IsCodedError(codedErr) {
			t.Error("IsCodedError should return true for CodedError")
		}
		
		if IsCodedError(regularErr) {
			t.Error("IsCodedError should return false for regular error")
		}
	})

	t.Run("GetErrorCode", func(t *testing.T) {
		codedErr := NewCodedError(ErrCodeFileNotFound, LevelError, "test", "", nil)
		regularErr := errors.New("regular error")
		
		if code := GetErrorCode(codedErr); code != ErrCodeFileNotFound {
			t.Errorf("GetErrorCode = %v, want %v", code, ErrCodeFileNotFound)
		}
		
		if code := GetErrorCode(regularErr); code != "" {
			t.Errorf("GetErrorCode for regular error = %v, want empty string", code)
		}
	})
}

func TestWithContext(t *testing.T) {
	err := NewCodedError(ErrCodeInternalError, LevelError, "test", "", nil)
	
	// Add context
	err.WithContext("key1", "value1").
		WithContext("key2", 123).
		WithContext("key3", true)
	
	// Check context values
	if val, ok := err.Context["key1"]; !ok || val != "value1" {
		t.Error("Context should contain key1 with value1")
	}
	
	if val, ok := err.Context["key2"]; !ok || val != 123 {
		t.Error("Context should contain key2 with 123")
	}
	
	if val, ok := err.Context["key3"]; !ok || val != true {
		t.Error("Context should contain key3 with true")
	}
}