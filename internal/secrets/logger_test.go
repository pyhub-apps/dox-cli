package secrets

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestSecureLogger_Sanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Strings that should be in the output
		notExpected []string // Strings that should NOT be in the output
	}{
		{
			name:     "OpenAI API key in log",
			input:    "Calling OpenAI with key sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			expected: []string{"Calling OpenAI with key"},
			notExpected: []string{"abcdefghijklmnopqrstuvwxyz"},
		},
		{
			name:     "API key in environment variable format",
			input:    "Setting OPENAI_API_KEY=sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			expected: []string{"OPENAI_API_KEY="},
			notExpected: []string{"abcdefghijklmnopqrstuvwxyz"},
		},
		{
			name:     "API key in JSON",
			input:    `{"api_key":"sk-proj-abcdefghijklmnopqrstuvwxyz123456","model":"gpt-4"}`,
			expected: []string{`"api_key"`, `"model":"gpt-4"`},
			notExpected: []string{"abcdefghijklmnopqrstuvwxyz"},
		},
		{
			name:     "Authorization header",
			input:    "Authorization: Bearer sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			expected: []string{"Authorization:"},
			notExpected: []string{"abcdefghijklmnopqrstuvwxyz"},
		},
		{
			name:     "Password in config",
			input:    "password=SuperSecret123!",
			expected: []string{"password="},
			notExpected: []string{"SuperSecret123"},
		},
		{
			name:     "Multiple sensitive fields",
			input:    "api_key=sk-12345678 and token=abc123def456 in the same line",
			expected: []string{"api_key=", "token=", "in the same line"},
			notExpected: []string{"12345678", "abc123def456"},
		},
		{
			name:     "No sensitive data",
			input:    "This is a normal log message without any sensitive information",
			expected: []string{"This is a normal log message"},
			notExpected: []string{},
		},
		{
			name:     "Long random token",
			input:    "Found token: abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOP",
			expected: []string{"Found token:"},
			notExpected: []string{"ghijklmnopqrstuvwxyz0123456789ABCDEFGH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger with a buffer to capture output
			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			sl := NewSecureLoggerWithWriter(logger)
			
			// Log the message
			sl.Println(tt.input)
			
			// Get the output
			output := buf.String()
			
			// Check expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got: %s", expected, output)
				}
			}
			
			// Check sensitive strings are NOT present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(output, notExpected) {
					t.Errorf("Output should NOT contain %q, got: %s", notExpected, output)
				}
			}
		})
	}
}

func TestMaskMatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Key-value with equals",
			input:    "api_key=sk-proj-abcdefghijklmnopqrstuvwxyz",
			expected: "api_key=sk-p***wxyz",
		},
		{
			name:     "Key-value with colon",
			input:    "token:abcdefghijklmnopqrstuvwxyz",
			expected: "token:abcd***wxyz",
		},
		{
			name:     "Standalone token",
			input:    "sk-proj-abcdefghijklmnopqrstuvwxyz",
			expected: "sk-p***wxyz",
		},
		{
			name:     "Short value",
			input:    "key=abc",
			expected: "key=[REDACTED]",
		},
		{
			name:     "Empty value",
			input:    "key=",
			expected: "key=[EMPTY]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskMatch(tt.input)
			if result != tt.expected {
				t.Errorf("maskMatch(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLooksLikeSensitiveToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "OpenAI key",
			token:    "sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			expected: true,
		},
		{
			name:     "API key prefix",
			token:    "api-key-abcdefghijklmnopqrstuvwxyz123456",
			expected: true,
		},
		{
			name:     "Token prefix",
			token:    "token_abcdefghijklmnopqrstuvwxyz123456",
			expected: true,
		},
		{
			name:     "High entropy string",
			token:    "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ABCD",
			expected: true,
		},
		{
			name:     "Regular UUID",
			token:    "550e8400e29b41d4a716446655440000",
			expected: false, // No mixed case
		},
		{
			name:     "Short string",
			token:    "abcdef123456", // Too short
			expected: false,
		},
		{
			name:     "All lowercase",
			token:    "abcdefghijklmnopqrstuvwxyz0123456789",
			expected: false, // No uppercase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeSensitiveToken(tt.token)
			if result != tt.expected {
				t.Errorf("looksLikeSensitiveToken(%q) = %v, want %v", tt.token, result, tt.expected)
			}
		})
	}
}