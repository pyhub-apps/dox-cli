package errors

import (
	"fmt"
	"strings"
)

// Additional error codes not in codes.go
const (
	// YAML-specific error code
	ErrCodeInvalidYAML     ErrorCode = "DOX206"
	
	// System error codes
	ErrCodeOutOfMemory      ErrorCode = "DOX600"
	ErrCodeDiskFull         ErrorCode = "DOX601"
	
	// API error codes
	ErrCodeAPIUnauthorized  ErrorCode = "DOX305"
	ErrCodeMissingAPIKey    ErrorCode = "DOX306"
)

// EnhancedError provides detailed error information with suggestions
type EnhancedError struct {
	Code        ErrorCode
	Message     string
	Details     string
	Suggestions []string
	Context     map[string]interface{}
	Wrapped     error
}

// Error implements the error interface
func (e *EnhancedError) Error() string {
	var sb strings.Builder
	
	// Main error message
	sb.WriteString(e.Message)
	
	// Add error code
	if e.Code != "" {
		sb.WriteString(fmt.Sprintf(" [%s]", e.Code))
	}
	
	// Add details if available
	if e.Details != "" {
		sb.WriteString("\n  Details: ")
		sb.WriteString(e.Details)
	}
	
	// Add suggestions
	if len(e.Suggestions) > 0 {
		sb.WriteString("\n  Suggestions:")
		for _, suggestion := range e.Suggestions {
			sb.WriteString("\n    • ")
			sb.WriteString(suggestion)
		}
	}
	
	// Add context if available
	if len(e.Context) > 0 {
		sb.WriteString("\n  Context:")
		for key, value := range e.Context {
			sb.WriteString(fmt.Sprintf("\n    %s: %v", key, value))
		}
	}
	
	return sb.String()
}

// Unwrap returns the wrapped error
func (e *EnhancedError) Unwrap() error {
	return e.Wrapped
}

// GetCode returns the error code
func (e *EnhancedError) GetCode() ErrorCode {
	return e.Code
}

// GetSuggestions returns the suggestions for fixing the error
func (e *EnhancedError) GetSuggestions() []string {
	return e.Suggestions
}

// ErrorBuilder helps construct enhanced errors
type ErrorBuilder struct {
	err *EnhancedError
}

// NewError creates a new error builder
func NewError(code ErrorCode, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &EnhancedError{
			Code:        code,
			Message:     message,
			Context:     make(map[string]interface{}),
			Suggestions: []string{},
		},
	}
}

// WithDetails adds details to the error
func (b *ErrorBuilder) WithDetails(details string) *ErrorBuilder {
	b.err.Details = details
	return b
}

// WithSuggestion adds a suggestion (prevents duplicates)
func (b *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	// Check for duplicates before adding
	for _, existing := range b.err.Suggestions {
		if existing == suggestion {
			return b
		}
	}
	b.err.Suggestions = append(b.err.Suggestions, suggestion)
	return b
}

// WithContext adds context information (skips nil/empty values)
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	// Skip nil values
	if value == nil {
		return b
	}
	
	// Skip empty string values
	if str, ok := value.(string); ok && str == "" {
		return b
	}
	
	b.err.Context[key] = value
	return b
}

// WithWrapped wraps another error
func (b *ErrorBuilder) WithWrapped(err error) *ErrorBuilder {
	b.err.Wrapped = err
	return b
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() error {
	return b.err
}

// Common error constructors

// FileNotFoundError creates a file not found error with suggestions
func FileNotFoundError(path string) error {
	return NewError(ErrCodeFileNotFound, fmt.Sprintf("File not found: '%s'", path)).
		WithContext("searched_path", path).
		WithSuggestion("Check if the file exists using 'ls' or 'dir'").
		WithSuggestion("Ensure you have the correct file path").
		WithSuggestion("Use an absolute path if the relative path isn't working").
		Build()
}

// PermissionDeniedError creates a permission denied error with suggestions
func PermissionDeniedError(path string, operation string) error {
	return NewError(ErrCodePermissionDenied, fmt.Sprintf("Permission denied: Cannot %s '%s'", operation, path)).
		WithContext("path", path).
		WithContext("operation", operation).
		WithSuggestion(fmt.Sprintf("Check file permissions: ls -l %s", path)).
		WithSuggestion("Run with appropriate permissions (sudo on Unix/Admin on Windows)").
		WithSuggestion("Ensure the file is not locked by another process").
		Build()
}

// InvalidYAMLError creates an invalid YAML error with context
func InvalidYAMLError(file string, line int, err error) error {
	return NewError(ErrCodeInvalidYAML, fmt.Sprintf("Invalid YAML syntax in '%s'", file)).
		WithDetails(fmt.Sprintf("Error at line %d: %v", line, err)).
		WithContext("file", file).
		WithContext("line", line).
		WithSuggestion("Use spaces for indentation, not tabs").
		WithSuggestion("Check for missing colons after keys").
		WithSuggestion("Ensure proper indentation (usually 2 spaces)").
		WithSuggestion("Validate your YAML at https://www.yamllint.com/").
		WithWrapped(err).
		Build()
}

// MissingAPIKeyError creates an API key missing error with setup instructions
func MissingAPIKeyError(provider string) error {
	suggestions := []string{
		fmt.Sprintf("Set the environment variable: export %s_API_KEY=your-key-here", strings.ToUpper(provider)),
		fmt.Sprintf("Or add to your shell profile: echo 'export %s_API_KEY=your-key' >> ~/.bashrc", strings.ToUpper(provider)),
		fmt.Sprintf("Or use the --api-key flag: dox generate --api-key your-key"),
	}
	
	return NewError(ErrCodeMissingAPIKey, fmt.Sprintf("%s API key is not configured", provider)).
		WithContext("provider", provider).
		WithSuggestion(suggestions[0]).
		WithSuggestion(suggestions[1]).
		WithSuggestion(suggestions[2]).
		Build()
}

// OutOfMemoryError creates an out of memory error with optimization suggestions
// fileSize is reported in MB with one decimal place for accuracy
func OutOfMemoryError(fileSize int64) error {
	fileSizeMB := float64(fileSize) / (1024 * 1024)
	return NewError(ErrCodeOutOfMemory, "Not enough memory to process document").
		WithDetails(fmt.Sprintf("File size: %.1f MB", fileSizeMB)).
		WithContext("file_size_mb", fmt.Sprintf("%.1f", fileSizeMB)).
		WithSuggestion("Use --streaming mode for large files: dox replace --streaming").
		WithSuggestion("Close other applications to free up memory").
		WithSuggestion("Process the file on a machine with more RAM").
		Build()
}

// InvalidDocumentFormatError creates an invalid document format error
func InvalidDocumentFormatError(path string, expected string, found string) error {
	return NewError(ErrCodeInvalidFormat, fmt.Sprintf("Invalid document format for '%s'", path)).
		WithDetails(fmt.Sprintf("Expected %s, but found %s", expected, found)).
		WithContext("path", path).
		WithContext("expected", expected).
		WithContext("found", found).
		WithSuggestion(fmt.Sprintf("Convert the file to %s format first", expected)).
		WithSuggestion("Check if the file extension matches the actual format").
		WithSuggestion("Ensure the file is not corrupted").
		Build()
}

// IsEnhancedError checks if an error is an EnhancedError
func IsEnhancedError(err error) bool {
	var ee *EnhancedError
	return As(err, &ee)
}