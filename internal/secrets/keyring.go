package secrets

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"
	
	"github.com/zalando/go-keyring"
)

const (
	// Service name for keyring storage
	ServiceName = "pyhub-docs"
	
	// Key prefixes for different providers
	OpenAIKeyPrefix = "openai_api_key"
	ClaudeKeyPrefix = "claude_api_key"
	AnthropicKeyPrefix = "anthropic_api_key"
)

var (
	// ErrKeyringNotSupported indicates the system doesn't support secure keyring
	ErrKeyringNotSupported = errors.New("secure keyring not supported on this system")
	
	// ErrKeyNotFound indicates the key was not found in the keyring
	ErrKeyNotFound = errors.New("key not found in keyring")
)

// SecureStorage provides secure storage for sensitive data like API keys
type SecureStorage struct {
	serviceName string
	fallback    bool // Use fallback storage if keyring is not available
}

// NewSecureStorage creates a new secure storage instance
func NewSecureStorage() *SecureStorage {
	return &SecureStorage{
		serviceName: ServiceName,
		fallback:    false,
	}
}

// IsSupported checks if secure keyring is supported on the current system
func (s *SecureStorage) IsSupported() bool {
	// Keyring is generally supported on macOS, Windows, and Linux with secret service
	switch runtime.GOOS {
	case "darwin", "windows":
		return true
	case "linux":
		// Linux support depends on secret service availability
		// We'll try to use it and handle errors gracefully
		return true
	default:
		return false
	}
}

// StoreAPIKey securely stores an API key in the system keyring
func (s *SecureStorage) StoreAPIKey(provider, apiKey string) error {
	if apiKey == "" {
		return errors.New("API key cannot be empty")
	}
	
	keyName := s.getKeyName(provider)
	
	// Encode the API key for additional obfuscation
	encoded := base64.StdEncoding.EncodeToString([]byte(apiKey))
	
	// Try to store in keyring
	err := keyring.Set(s.serviceName, keyName, encoded)
	if err != nil {
		// Check if it's a temporary error or system doesn't support keyring
		if strings.Contains(err.Error(), "not supported") || 
		   strings.Contains(err.Error(), "not available") {
			return ErrKeyringNotSupported
		}
		return fmt.Errorf("failed to store API key: %w", err)
	}
	
	return nil
}

// RetrieveAPIKey retrieves an API key from the system keyring
func (s *SecureStorage) RetrieveAPIKey(provider string) (string, error) {
	keyName := s.getKeyName(provider)
	
	// Try to get from keyring
	encoded, err := keyring.Get(s.serviceName, keyName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return "", ErrKeyNotFound
		}
		if strings.Contains(err.Error(), "not supported") || 
		   strings.Contains(err.Error(), "not available") {
			return "", ErrKeyringNotSupported
		}
		return "", fmt.Errorf("failed to retrieve API key: %w", err)
	}
	
	// Decode the API key
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// If decoding fails, assume it's stored in plain text (backward compatibility)
		return encoded, nil
	}
	
	return string(decoded), nil
}

// DeleteAPIKey removes an API key from the system keyring
func (s *SecureStorage) DeleteAPIKey(provider string) error {
	keyName := s.getKeyName(provider)
	
	err := keyring.Delete(s.serviceName, keyName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrKeyNotFound
		}
		if strings.Contains(err.Error(), "not supported") || 
		   strings.Contains(err.Error(), "not available") {
			return ErrKeyringNotSupported
		}
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	
	return nil
}

// ListProviders returns a list of providers that have stored API keys
func (s *SecureStorage) ListProviders() ([]string, error) {
	// This is a simplified implementation
	// The go-keyring library doesn't provide a list function
	// We'll check for known providers
	providers := []string{}
	knownProviders := []string{"openai", "claude", "anthropic"}
	
	for _, provider := range knownProviders {
		if _, err := s.RetrieveAPIKey(provider); err == nil {
			providers = append(providers, provider)
		}
	}
	
	return providers, nil
}

// getKeyName returns the keyring key name for a provider
func (s *SecureStorage) getKeyName(provider string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return OpenAIKeyPrefix
	case "claude", "anthropic":
		return ClaudeKeyPrefix
	default:
		return fmt.Sprintf("%s_api_key", strings.ToLower(provider))
	}
}

// MaskAPIKey masks an API key for safe display in logs or UI
func MaskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	
	length := len(apiKey)
	if length <= 2 {
		// Very short key, don't mask
		return apiKey
	}
	
	if length <= 8 {
		// Short key, mask all but first 2 chars
		return apiKey[:2] + strings.Repeat("*", length-2)
	}
	
	// Show first 4 and last 4 characters
	if length <= 12 {
		// Medium key, show first 3 and last 3
		return apiKey[:3] + strings.Repeat("*", length-6) + apiKey[length-3:]
	}
	
	// Standard masking: show first 4 and last 4
	return apiKey[:4] + strings.Repeat("*", length-8) + apiKey[length-4:]
}

// ValidateAPIKey performs basic validation on an API key
func ValidateAPIKey(provider, apiKey string) error {
	if apiKey == "" {
		return errors.New("API key cannot be empty")
	}
	
	// Check for common mistakes first (before trimming)
	if strings.Contains(apiKey, "\n") || strings.Contains(apiKey, "\r") {
		return errors.New("API key should not contain newlines")
	}
	
	// Remove common prefixes that users might accidentally include
	apiKey = strings.TrimSpace(apiKey)
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	apiKey = strings.TrimPrefix(apiKey, "bearer ")
	
	// Provider-specific validation
	switch strings.ToLower(provider) {
	case "openai":
		// OpenAI keys typically start with "sk-"
		if !strings.HasPrefix(apiKey, "sk-") {
			return errors.New("OpenAI API key should start with 'sk-'")
		}
		if len(apiKey) < 20 {
			return errors.New("OpenAI API key appears to be too short")
		}
		
	case "claude", "anthropic":
		// Anthropic keys typically start with "sk-ant-"
		if !strings.HasPrefix(apiKey, "sk-ant-") {
			return errors.New("Anthropic/Claude API key should start with 'sk-ant-'")
		}
		if len(apiKey) < 20 {
			return errors.New("Anthropic/Claude API key appears to be too short")
		}
		
	default:
		// Generic validation
		if len(apiKey) < 10 {
			return fmt.Errorf("%s API key appears to be too short", provider)
		}
	}
	
	// Check for common mistakes (spaces after trimming)
	if strings.Contains(apiKey, " ") {
		return errors.New("API key should not contain spaces")
	}
	
	return nil
}