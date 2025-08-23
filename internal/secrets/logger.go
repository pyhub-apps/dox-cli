package secrets

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// SecureLogger wraps standard logger to prevent logging sensitive information
type SecureLogger struct {
	logger *log.Logger
	// Patterns to detect and mask sensitive information
	patterns []*regexp.Regexp
}

// NewSecureLogger creates a new secure logger
func NewSecureLogger() *SecureLogger {
	return &SecureLogger{
		logger:   log.New(os.Stderr, "", log.LstdFlags),
		patterns: initPatterns(),
	}
}

// NewSecureLoggerWithWriter creates a new secure logger with custom writer
func NewSecureLoggerWithWriter(writer *log.Logger) *SecureLogger {
	return &SecureLogger{
		logger:   writer,
		patterns: initPatterns(),
	}
}

// initPatterns initializes regex patterns for detecting sensitive data
func initPatterns() []*regexp.Regexp {
	patternStrings := []string{
		// API Keys
		`(sk-[a-zA-Z0-9]{20,})`,                    // OpenAI style keys
		`(sk-ant-[a-zA-Z0-9]{20,})`,                // Anthropic keys
		`(api[_-]?key\s*[=:]\s*)([^\s,;]+)`,        // Generic API key patterns
		`(OPENAI_API_KEY\s*[=:]\s*)([^\s,;]+)`,     // OpenAI env var
		`(ANTHROPIC_API_KEY\s*[=:]\s*)([^\s,;]+)`,  // Anthropic env var
		`(CLAUDE_API_KEY\s*[=:]\s*)([^\s,;]+)`,     // Claude env var
		
		// Authorization headers
		`(Authorization:\s*Bearer\s+)([^\s]+)`,
		`(X-API-Key:\s*)([^\s]+)`,
		
		// Passwords and tokens
		`(password\s*[=:]\s*)([^\s,;]+)`,
		`(token\s*[=:]\s*)([^\s,;]+)`,
		`(secret\s*[=:]\s*)([^\s,;]+)`,
		
		// JSON formatted secrets
		`"api_key"\s*:\s*"([^"]+)"`,
		`"apiKey"\s*:\s*"([^"]+)"`,
		`"password"\s*:\s*"([^"]+)"`,
		`"token"\s*:\s*"([^"]+)"`,
		`"secret"\s*:\s*"([^"]+)"`,
	}
	
	patterns := make([]*regexp.Regexp, 0, len(patternStrings))
	for _, pattern := range patternStrings {
		// Compile with case-insensitive flag
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			// Log compilation error but continue
			fmt.Fprintf(os.Stderr, "Warning: Failed to compile pattern %s: %v\n", pattern, err)
			continue
		}
		patterns = append(patterns, re)
	}
	
	return patterns
}

// Printf logs a formatted message after sanitizing sensitive information
func (sl *SecureLogger) Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	sanitized := sl.sanitize(message)
	sl.logger.Print(sanitized)
}

// Println logs a message after sanitizing sensitive information
func (sl *SecureLogger) Println(v ...interface{}) {
	message := fmt.Sprint(v...)
	sanitized := sl.sanitize(message)
	sl.logger.Println(sanitized)
}

// Print logs a message after sanitizing sensitive information
func (sl *SecureLogger) Print(v ...interface{}) {
	message := fmt.Sprint(v...)
	sanitized := sl.sanitize(message)
	sl.logger.Print(sanitized)
}

// Fatalf logs a formatted message and exits after sanitizing sensitive information
func (sl *SecureLogger) Fatalf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	sanitized := sl.sanitize(message)
	sl.logger.Fatal(sanitized)
}

// Fatal logs a message and exits after sanitizing sensitive information
func (sl *SecureLogger) Fatal(v ...interface{}) {
	message := fmt.Sprint(v...)
	sanitized := sl.sanitize(message)
	sl.logger.Fatal(sanitized)
}

// sanitize removes or masks sensitive information from a message
func (sl *SecureLogger) sanitize(message string) string {
	result := message
	
	// Apply all patterns to mask sensitive data
	for _, pattern := range sl.patterns {
		result = pattern.ReplaceAllStringFunc(result, maskMatch)
	}
	
	// Additional safety: mask any remaining potential API keys
	result = maskLongTokens(result)
	
	return result
}

// maskMatch masks a regex match, preserving some structure for debugging
func maskMatch(match string) string {
	// Check if this is a key=value or key:value pattern
	if strings.Contains(match, "=") || strings.Contains(match, ":") {
		parts := regexp.MustCompile(`[=:]`).Split(match, 2)
		if len(parts) == 2 {
			// Preserve the key, mask the value
			separator := "="
			if strings.Contains(match, ":") {
				separator = ":"
			}
			return parts[0] + separator + maskValue(strings.TrimSpace(parts[1]))
		}
	}
	
	// For standalone tokens, mask them entirely
	return maskValue(match)
}

// maskValue masks a sensitive value while preserving some information for debugging
func maskValue(value string) string {
	// Remove quotes if present
	value = strings.Trim(value, `"'`)
	
	if value == "" {
		return "[EMPTY]"
	}
	
	length := len(value)
	
	// For very short values, mask entirely
	if length <= 4 {
		return "[REDACTED]"
	}
	
	// For medium length, show first 2 and last 2
	if length <= 20 {
		return value[:2] + "***" + value[length-2:]
	}
	
	// For long values, show first 4 and last 4
	return value[:4] + "***" + value[length-4:]
}

// maskLongTokens masks any suspiciously long tokens that might be API keys
func maskLongTokens(message string) string {
	// Look for long alphanumeric strings that might be tokens
	tokenPattern := regexp.MustCompile(`\b[a-zA-Z0-9_-]{32,}\b`)
	
	return tokenPattern.ReplaceAllStringFunc(message, func(match string) string {
		// Check if this looks like it might be a sensitive token
		if looksLikeSensitiveToken(match) {
			return maskValue(match)
		}
		return match
	})
}

// looksLikeSensitiveToken checks if a string looks like it might be a sensitive token
func looksLikeSensitiveToken(token string) bool {
	// Check for common API key prefixes
	prefixes := []string{"sk-", "pk-", "api-", "key-", "token-", "pat-", "sk_", "pk_", "api_", "key_", "token_", "pat_"}
	
	lowerToken := strings.ToLower(token)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowerToken, prefix) {
			return true
		}
	}
	
	// Check if it has high entropy (mix of upper, lower, numbers)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(token)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(token)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(token)
	
	// If it has all three and is long, it's probably a token
	if hasUpper && hasLower && hasDigit && len(token) >= 32 {
		return true
	}
	
	return false
}

// DefaultSecureLogger is a package-level secure logger instance
var DefaultSecureLogger = NewSecureLogger()

// Debugf logs a debug message after sanitizing (only if debug mode is enabled)
func Debugf(format string, v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		DefaultSecureLogger.Printf("[DEBUG] "+format, v...)
	}
}

// Infof logs an info message after sanitizing
func Infof(format string, v ...interface{}) {
	DefaultSecureLogger.Printf("[INFO] "+format, v...)
}

// Warnf logs a warning message after sanitizing
func Warnf(format string, v ...interface{}) {
	DefaultSecureLogger.Printf("[WARN] "+format, v...)
}

// Errorf logs an error message after sanitizing
func Errorf(format string, v ...interface{}) {
	DefaultSecureLogger.Printf("[ERROR] "+format, v...)
}