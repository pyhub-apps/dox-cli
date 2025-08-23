package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/i18n"
)

func TestLocalizedErrors(t *testing.T) {
	// Initialize i18n with English
	if err := i18n.Init("en"); err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}

	tests := []struct {
		name            string
		errFunc         func() error
		expectedInError []string
	}{
		{
			name: "LocalizedFileNotFoundError_EN",
			errFunc: func() error {
				return LocalizedFileNotFoundError("/path/to/missing.txt")
			},
			expectedInError: []string{
				"File not found",
				"/path/to/missing.txt",
				"Check if the file exists",
			},
		},
		{
			name: "LocalizedPermissionDeniedError_EN",
			errFunc: func() error {
				return LocalizedPermissionDeniedError("/etc/passwd", "write")
			},
			expectedInError: []string{
				"Permission denied",
				"/etc/passwd",
				"write",
				"Check file permissions",
			},
		},
		{
			name: "LocalizedInvalidYAMLError_EN",
			errFunc: func() error {
				return LocalizedInvalidYAMLError("config.yml", 10, nil)
			},
			expectedInError: []string{
				"Invalid YAML",
				"config.yml",
				"line 10",
				"spaces",
			},
		},
		{
			name: "LocalizedMissingAPIKeyError_EN",
			errFunc: func() error {
				return LocalizedMissingAPIKeyError("OpenAI")
			},
			expectedInError: []string{
				"OpenAI",
				"API key",
				"environment variable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			errStr := err.Error()
			for _, expected := range tt.expectedInError {
				if !strings.Contains(errStr, expected) {
					t.Errorf("Error message should contain %q, got: %s", expected, errStr)
				}
			}
		})
	}
}

func TestLocalizedErrorsKorean(t *testing.T) {
	// Initialize i18n with Korean
	if err := i18n.Init("ko"); err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}

	tests := []struct {
		name            string
		errFunc         func() error
		expectedInError []string
	}{
		{
			name: "LocalizedFileNotFoundError_KO",
			errFunc: func() error {
				return LocalizedFileNotFoundError("/경로/파일.txt")
			},
			expectedInError: []string{
				"파일을 찾을 수 없습니다",
				"/경로/파일.txt",
				"파일이 존재하는지 확인",
			},
		},
		{
			name: "LocalizedPermissionDeniedError_KO",
			errFunc: func() error {
				return LocalizedPermissionDeniedError("/etc/passwd", "쓰기")
			},
			expectedInError: []string{
				"권한 거부",
				"/etc/passwd",
				"쓰기",
				"파일 권한 확인",
			},
		},
		{
			name: "LocalizedMissingAPIKeyError_KO",
			errFunc: func() error {
				return LocalizedMissingAPIKeyError("OpenAI")
			},
			expectedInError: []string{
				"OpenAI",
				"API 키",
				"환경 변수",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			errStr := err.Error()
			for _, expected := range tt.expectedInError {
				if !strings.Contains(errStr, expected) {
					t.Errorf("Error message should contain %q, got: %s", expected, errStr)
				}
			}
		})
	}

	// Reset to English for other tests
	i18n.Init("en")
}

func TestLocalizedErrorBuilder(t *testing.T) {
	// Initialize i18n
	if err := i18n.Init("en"); err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}

	err := NewLocalizedError(ErrCodeFileNotFound, MsgErrFileNotFound, "/test/file.txt").
		WithLocalizedDetails(MsgDetailSearchedIn, "/test").
		WithLocalizedSuggestion(MsgSugCheckFileExists).
		WithContext("attempts", 3).
		Build()

	errStr := err.Error()

	expectedParts := []string{
		"File not found",
		"/test/file.txt",
		"Searched in: /test",
		"Check if the file exists",
		"attempts: 3",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errStr, part) {
			t.Errorf("Error should contain %q, got: %s", part, errStr)
		}
	}
}

func TestFormatError(t *testing.T) {
	// Test with enhanced error
	enhancedErr := FileNotFoundError("/missing.txt")
	
	// Test verbose formatting
	verboseOutput := FormatError(enhancedErr, true)
	if !strings.Contains(verboseOutput, "DOX100") {
		t.Error("Verbose output should contain error code")
	}
	if !strings.Contains(verboseOutput, "Suggestions:") {
		t.Error("Verbose output should contain suggestions section")
	}

	// Test non-verbose formatting
	simpleOutput := FormatError(enhancedErr, false)
	if strings.Contains(simpleOutput, "Context:") {
		t.Error("Simple output should not contain context section")
	}
	if !strings.Contains(simpleOutput, "💡") {
		t.Error("Simple output should contain suggestion emoji")
	}

	// Test with regular error
	regularErr := fmt.Errorf("simple error")
	output := FormatError(regularErr, true)
	if output != "simple error" {
		t.Errorf("Regular error formatting incorrect: %s", output)
	}
}