package errors

import (
	"errors"
	"fmt"
	"strings"
	
	"github.com/pyhub/pyhub-docs/internal/i18n"
)

// ErrorCode represents a unique error code
type ErrorCode string

// Error code constants
const (
	// Configuration errors (DOX001-DOX099)
	ErrCodeAPIKeyNotFound    ErrorCode = "DOX001"
	ErrCodeInvalidConfig     ErrorCode = "DOX002"
	ErrCodeConfigNotFound    ErrorCode = "DOX003"
	ErrCodeInvalidAPIKey     ErrorCode = "DOX004"
	ErrCodeConfigSaveFailed  ErrorCode = "DOX005"
	
	// File operation errors (DOX100-DOX199)
	ErrCodeFileNotFound      ErrorCode = "DOX100"
	ErrCodeFileReadFailed    ErrorCode = "DOX101"
	ErrCodeFileWriteFailed   ErrorCode = "DOX102"
	ErrCodePermissionDenied  ErrorCode = "DOX103"
	ErrCodeFileAlreadyExists ErrorCode = "DOX104"
	ErrCodeInvalidPath       ErrorCode = "DOX105"
	
	// Document processing errors (DOX200-DOX299)
	ErrCodeDocumentCorrupted ErrorCode = "DOX200"
	ErrCodeUnsupportedFormat ErrorCode = "DOX201"
	ErrCodeEmptyDocument     ErrorCode = "DOX202"
	ErrCodeDocumentParseFailed ErrorCode = "DOX203"
	ErrCodeTemplateParseFailed ErrorCode = "DOX204"
	
	// AI/Generation errors (DOX300-DOX399)
	ErrCodeAIRequestFailed   ErrorCode = "DOX300"
	ErrCodeAIRateLimited     ErrorCode = "DOX301"
	ErrCodeAITimeout         ErrorCode = "DOX302"
	ErrCodeAIInvalidResponse ErrorCode = "DOX303"
	ErrCodeAIServiceDown     ErrorCode = "DOX304"
	
	// Validation errors (DOX400-DOX499)
	ErrCodeInvalidInput      ErrorCode = "DOX400"
	ErrCodeMissingRequired   ErrorCode = "DOX401"
	ErrCodeInvalidFormat     ErrorCode = "DOX402"
	ErrCodeOutOfRange        ErrorCode = "DOX403"
	
	// Network errors (DOX500-DOX599)
	ErrCodeNetworkTimeout    ErrorCode = "DOX500"
	ErrCodeConnectionRefused ErrorCode = "DOX501"
	ErrCodeDNSResolutionFailed ErrorCode = "DOX502"
	
	// Internal errors (DOX900-DOX999)
	ErrCodeInternalError     ErrorCode = "DOX900"
	ErrCodeNotImplemented    ErrorCode = "DOX901"
)

// ErrorLevel represents the severity of an error
type ErrorLevel string

const (
	LevelError   ErrorLevel = "ERROR"
	LevelWarning ErrorLevel = "WARNING"
	LevelInfo    ErrorLevel = "INFO"
)

// CodedError represents an error with a code, level, and solution
type CodedError struct {
	Code     ErrorCode
	Level    ErrorLevel
	Message  string
	Solution string
	Context  map[string]interface{}
	Cause    error
}

// Error implements the error interface
func (e *CodedError) Error() string {
	return e.LocalizedError()
}

// LocalizedError returns a localized error message
func (e *CodedError) LocalizedError() string {
	var sb strings.Builder
	
	// Try to get localized error message
	errorMsgKey := fmt.Sprintf("error.code.%s", strings.ToLower(string(e.Code)))
	errorMsgKey = strings.ReplaceAll(errorMsgKey, "dox", "")
	errorMsgKey = "error.code." + strings.ToLower(string(e.Code)[3:]) // Remove DOX prefix
	
	// Map error codes to message keys
	msgKeyMap := map[ErrorCode]string{
		ErrCodeAPIKeyNotFound:    i18n.MsgErrCodeAPIKeyNotFound,
		ErrCodeInvalidConfig:     i18n.MsgErrCodeInvalidConfig,
		ErrCodeConfigNotFound:    i18n.MsgErrCodeConfigNotFound,
		ErrCodeInvalidAPIKey:     i18n.MsgErrCodeInvalidAPIKey,
		ErrCodeConfigSaveFailed:  i18n.MsgErrCodeConfigSaveFailed,
		ErrCodeFileNotFound:      i18n.MsgErrCodeFileNotFound,
		ErrCodeFileReadFailed:    i18n.MsgErrCodeFileReadFailed,
		ErrCodeFileWriteFailed:   i18n.MsgErrCodeFileWriteFailed,
		ErrCodePermissionDenied:  i18n.MsgErrCodePermissionDenied,
		ErrCodeFileAlreadyExists: i18n.MsgErrCodeFileAlreadyExists,
		ErrCodeInvalidPath:       i18n.MsgErrCodeInvalidPath,
		ErrCodeDocumentCorrupted: i18n.MsgErrCodeDocumentCorrupted,
		ErrCodeUnsupportedFormat: i18n.MsgErrCodeUnsupportedFormat,
		ErrCodeEmptyDocument:     i18n.MsgErrCodeEmptyDocument,
		ErrCodeDocumentParseFailed: i18n.MsgErrCodeDocumentParseFailed,
		ErrCodeTemplateParseFailed: i18n.MsgErrCodeTemplateParseFailed,
		ErrCodeAIRequestFailed:   i18n.MsgErrCodeAIRequestFailed,
		ErrCodeAIRateLimited:     i18n.MsgErrCodeAIRateLimited,
		ErrCodeAITimeout:         i18n.MsgErrCodeAITimeout,
		ErrCodeAIInvalidResponse: i18n.MsgErrCodeAIInvalidResponse,
		ErrCodeAIServiceDown:     i18n.MsgErrCodeAIServiceDown,
		ErrCodeInvalidInput:      i18n.MsgErrCodeInvalidInput,
		ErrCodeMissingRequired:   i18n.MsgErrCodeMissingRequired,
		ErrCodeInvalidFormat:     i18n.MsgErrCodeInvalidFormat,
		ErrCodeOutOfRange:        i18n.MsgErrCodeOutOfRange,
		ErrCodeNetworkTimeout:    i18n.MsgErrCodeNetworkTimeout,
		ErrCodeConnectionRefused: i18n.MsgErrCodeConnectionRefused,
		ErrCodeDNSResolutionFailed: i18n.MsgErrCodeDNSResolutionFailed,
		ErrCodeInternalError:     i18n.MsgErrCodeInternalError,
		ErrCodeNotImplemented:    i18n.MsgErrCodeNotImplemented,
	}
	
	// Get localized message or fallback to default
	if msgKey, ok := msgKeyMap[e.Code]; ok {
		localizedMsg := i18n.T(msgKey, e.Context)
		sb.WriteString(localizedMsg)
	} else {
		// Fallback to original format
		sb.WriteString(fmt.Sprintf("[%s] [%s]: %s", e.Level, e.Code, e.Message))
	}
	
	// Add additional context if not already in the localized message
	if len(e.Context) > 0 && !strings.Contains(sb.String(), "Context:") {
		sb.WriteString("\nContext:")
		for key, value := range e.Context {
			sb.WriteString(fmt.Sprintf("\n  %s: %v", key, value))
		}
	}
	
	// Add solution if available
	if e.Solution != "" {
		sb.WriteString("\nSolution: ")
		sb.WriteString(e.Solution)
	}
	
	// Add cause if available
	if e.Cause != nil {
		sb.WriteString("\nCause: ")
		sb.WriteString(e.Cause.Error())
	}
	
	return sb.String()
}

// Unwrap returns the underlying error
func (e *CodedError) Unwrap() error {
	return e.Cause
}

// GetCode returns the error code
func (e *CodedError) GetCode() ErrorCode {
	return e.Code
}

// GetLevel returns the error level
func (e *CodedError) GetLevel() ErrorLevel {
	return e.Level
}

// WithContext adds context to the error
func (e *CodedError) WithContext(key string, value interface{}) *CodedError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewCodedError creates a new coded error
func NewCodedError(code ErrorCode, level ErrorLevel, message, solution string, cause error) *CodedError {
	return &CodedError{
		Code:     code,
		Level:    level,
		Message:  message,
		Solution: solution,
		Context:  make(map[string]interface{}),
		Cause:    cause,
	}
}

// Helper functions for creating common errors

// NewAPIKeyNotFoundError creates an API key not found error
func NewAPIKeyNotFoundError(provider string) *CodedError {
	// Determine which solution message to use
	var solutionKey string
	switch strings.ToLower(provider) {
	case "openai":
		solutionKey = i18n.MsgSolutionAPIKeyOpenAI
	case "claude":
		solutionKey = i18n.MsgSolutionAPIKeyClaude
	default:
		solutionKey = i18n.MsgSolutionAPIKeyGeneric
	}
	
	solution := i18n.T(solutionKey, nil)
	
	return NewCodedError(
		ErrCodeAPIKeyNotFound,
		LevelError,
		fmt.Sprintf("API key not found for %s", provider),
		solution,
		ErrMissingAPIKey,
	).WithContext("Provider", provider)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string) *CodedError {
	solution := i18n.T(i18n.MsgSolutionCheckFile, nil)
	return NewCodedError(
		ErrCodeFileNotFound,
		LevelError,
		fmt.Sprintf("File not found: %s", path),
		solution,
		ErrFileNotFound,
	).WithContext("Path", path)
}

// NewInvalidFormatError creates an invalid format error
func NewInvalidFormatError(format, expected string) *CodedError {
	solution := i18n.T(i18n.MsgSolutionCheckFormat, map[string]interface{}{"Expected": expected})
	return NewCodedError(
		ErrCodeInvalidFormat,
		LevelError,
		fmt.Sprintf("Invalid format: %s", format),
		solution,
		ErrInvalidFormat,
	).WithContext("Format", format).WithContext("Expected", expected)
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(path string) *CodedError {
	solution := i18n.T(i18n.MsgSolutionCheckPermission, nil)
	return NewCodedError(
		ErrCodePermissionDenied,
		LevelError,
		fmt.Sprintf("Permission denied: %s", path),
		solution,
		ErrPermissionDenied,
	).WithContext("Path", path)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(provider string, retryAfter string) *CodedError {
	var solution string
	if retryAfter != "" {
		solution = i18n.T(i18n.MsgSolutionWaitRetry, map[string]interface{}{"RetryAfter": retryAfter})
	} else {
		solution = i18n.T(i18n.MsgSolutionUpgradeAPI, nil)
	}
	
	return NewCodedError(
		ErrCodeAIRateLimited,
		LevelWarning,
		fmt.Sprintf("%s API rate limit exceeded", provider),
		solution,
		nil,
	).WithContext("Provider", provider).WithContext("RetryAfter", retryAfter)
}

// IsCodedError checks if an error is a CodedError
func IsCodedError(err error) bool {
	var ce *CodedError
	return errors.As(err, &ce)
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	var ce *CodedError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}