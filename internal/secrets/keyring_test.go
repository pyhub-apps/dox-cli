package secrets

import (
	"strings"
	"testing"
)

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "Empty key",
			apiKey:   "",
			expected: "",
		},
		{
			name:     "Very short key",
			apiKey:   "ab",
			expected: "ab",
		},
		{
			name:     "Short key",
			apiKey:   "abcdef",
			expected: "ab****",
		},
		{
			name:     "Medium key",
			apiKey:   "sk-abc123def456",
			expected: "sk-a*******f456",
		},
		{
			name:     "Long OpenAI key",
			apiKey:   "sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			expected: "sk-p********************************3456",
		},
		{
			name:     "Anthropic key",
			apiKey:   "sk-ant-api03-abcdefghijklmnop",
			expected: "sk-a*********************mnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAPIKey(tt.apiKey)
			if result != tt.expected {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.apiKey, result, tt.expected)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		apiKey    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid OpenAI key",
			provider:  "openai",
			apiKey:    "sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			wantError: false,
		},
		{
			name:      "Valid Anthropic key",
			provider:  "claude",
			apiKey:    "sk-ant-api03-abcdefghijklmnopqrstuvwxyz123456",
			wantError: false,
		},
		{
			name:      "Empty key",
			provider:  "openai",
			apiKey:    "",
			wantError: true,
			errorMsg:  "API key cannot be empty",
		},
		{
			name:      "OpenAI key wrong prefix",
			provider:  "openai",
			apiKey:    "pk-abcdefghijklmnopqrstuvwxyz",
			wantError: true,
			errorMsg:  "should start with 'sk-'",
		},
		{
			name:      "OpenAI key too short",
			provider:  "openai",
			apiKey:    "sk-abc",
			wantError: true,
			errorMsg:  "too short",
		},
		{
			name:      "Anthropic key wrong prefix",
			provider:  "anthropic",
			apiKey:    "sk-abcdefghijklmnopqrstuvwxyz",
			wantError: true,
			errorMsg:  "should start with 'sk-ant-'",
		},
		{
			name:      "Key with spaces",
			provider:  "openai",
			apiKey:    "sk-proj abcdefghijklmnopqrstuvwxyz",
			wantError: true,
			errorMsg:  "should not contain spaces",
		},
		{
			name:      "Key with newline",
			provider:  "openai",
			apiKey:    "sk-proj-abcdefghijklmnopqrstuvwxyz\n",
			wantError: true,
			errorMsg:  "should not contain newlines",
		},
		{
			name:      "Key with Bearer prefix",
			provider:  "openai",
			apiKey:    "Bearer sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			wantError: false, // Should be cleaned and validated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.provider, tt.apiKey)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAPIKey() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateAPIKey() error = %v, should contain %q", err, tt.errorMsg)
				}
			}
		})
	}
}

func TestSecureStorage_IsSupported(t *testing.T) {
	storage := NewSecureStorage()
	
	// Should return true on macOS, Windows, and Linux
	// We can't test the actual value as it depends on the OS
	supported := storage.IsSupported()
	t.Logf("Keyring supported on this system: %v", supported)
}

func TestGetKeyName(t *testing.T) {
	storage := NewSecureStorage()
	
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "openai_api_key"},
		{"OpenAI", "openai_api_key"},
		{"claude", "claude_api_key"},
		{"Claude", "claude_api_key"},
		{"anthropic", "claude_api_key"},
		{"Anthropic", "claude_api_key"},
		{"custom", "custom_api_key"},
		{"CUSTOM", "custom_api_key"},
	}
	
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := storage.getKeyName(tt.provider)
			if result != tt.expected {
				t.Errorf("getKeyName(%q) = %q, want %q", tt.provider, result, tt.expected)
			}
		})
	}
}