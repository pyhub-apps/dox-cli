package errors

import (
	"errors"
	"fmt"
	"github.com/pyhub/pyhub-docs/internal/i18n"
)

// Error message keys for i18n
const (
	// Error messages
	MsgErrFileNotFound       = "error.file_not_found"
	MsgErrPermissionDenied   = "error.permission_denied"
	MsgErrInvalidFormat      = "error.invalid_format"
	MsgErrDocumentCorrupted  = "error.document_corrupted"
	MsgErrConfigNotFound     = "error.config_not_found"
	MsgErrInvalidYAML        = "error.invalid_yaml"
	MsgErrMissingAPIKey      = "error.missing_api_key"
	MsgErrOutOfMemory        = "error.out_of_memory"
	MsgErrNetworkTimeout     = "error.network_timeout"
	
	// Suggestion messages
	MsgSugCheckFileExists    = "suggestion.check_file_exists"
	MsgSugUseAbsolutePath    = "suggestion.use_absolute_path"
	MsgSugCheckPermissions   = "suggestion.check_permissions"
	MsgSugRunAsAdmin         = "suggestion.run_as_admin"
	MsgSugUseSpaces          = "suggestion.use_spaces_not_tabs"
	MsgSugSetAPIKey          = "suggestion.set_api_key"
	MsgSugUseStreaming       = "suggestion.use_streaming"
	MsgSugCloseApps          = "suggestion.close_apps"
	MsgSugCheckFormat        = "suggestion.check_format"
	MsgSugRetryRequest       = "suggestion.retry_request"
	
	// Detail messages
	MsgDetailSearchedIn      = "detail.searched_in"
	MsgDetailErrorAtLine     = "detail.error_at_line"
	MsgDetailExpectedFormat  = "detail.expected_format"
	MsgDetailFileSize        = "detail.file_size"
	MsgDetailOperation       = "detail.operation"
)

// LocalizedError creates an enhanced error with i18n support
type LocalizedError struct {
	*EnhancedError
	Locale string
}

// LocalizedErrorBuilder helps construct localized errors
type LocalizedErrorBuilder struct {
	*ErrorBuilder
	locale string
}

// NewLocalizedError creates a new localized error builder
func NewLocalizedError(code ErrorCode, messageKey string, args ...interface{}) *LocalizedErrorBuilder {
	data := make(map[string]interface{})
	if len(args) > 0 {
		// Map arguments to template data based on message key
		switch messageKey {
		case MsgErrFileNotFound:
			data["Path"] = args[0]
		case MsgErrPermissionDenied:
			if len(args) >= 2 {
				data["Operation"] = args[0]
				data["Path"] = args[1]
			}
		case MsgErrInvalidYAML, MsgErrConfigNotFound:
			data["File"] = args[0]
		case MsgErrMissingAPIKey:
			data["Provider"] = args[0]
		case MsgErrInvalidFormat:
			data["Path"] = args[0]
		}
	}
	
	message := i18n.T(messageKey, data)
	
	return &LocalizedErrorBuilder{
		ErrorBuilder: NewError(code, message),
		locale:      i18n.GetCurrentLanguage(),
	}
}

// WithLocalizedDetails adds localized details
func (b *LocalizedErrorBuilder) WithLocalizedDetails(detailKey string, args ...interface{}) *LocalizedErrorBuilder {
	data := make(map[string]interface{})
	if len(args) > 0 {
		// Map arguments based on detail key
		switch detailKey {
		case MsgDetailErrorAtLine:
			if len(args) >= 2 {
				data["Line"] = args[0]
				data["Error"] = args[1]
			}
		case MsgDetailExpectedFormat:
			if len(args) >= 2 {
				data["Expected"] = args[0]
				data["Found"] = args[1]
			}
		case MsgDetailFileSize:
			data["Size"] = args[0]
		case MsgDetailSearchedIn:
			data["Path"] = args[0]
		case MsgDetailOperation:
			data["Operation"] = args[0]
		}
	}
	
	details := i18n.T(detailKey, data)
	b.ErrorBuilder.WithDetails(details)
	return b
}

// WithLocalizedSuggestion adds a localized suggestion
func (b *LocalizedErrorBuilder) WithLocalizedSuggestion(suggestionKey string, args ...interface{}) *LocalizedErrorBuilder {
	data := make(map[string]interface{})
	if len(args) > 0 {
		// Map arguments based on suggestion key
		switch suggestionKey {
		case MsgSugCheckPermissions:
			data["Path"] = args[0]
		case MsgSugSetAPIKey:
			data["Provider"] = args[0]
		case MsgSugCheckFormat:
			data["Format"] = args[0]
		}
	}
	
	suggestion := i18n.T(suggestionKey, data)
	b.ErrorBuilder.WithSuggestion(suggestion)
	return b
}

// Localized error constructors

// LocalizedFileNotFoundError creates a localized file not found error
func LocalizedFileNotFoundError(path string) error {
	b := NewLocalizedError(ErrCodeFileNotFound, MsgErrFileNotFound, path)
	b.WithContext("path", path)
	b.WithLocalizedSuggestion(MsgSugCheckFileExists)
	b.WithLocalizedSuggestion(MsgSugUseAbsolutePath)
	return b.Build()
}

// LocalizedPermissionDeniedError creates a localized permission denied error
func LocalizedPermissionDeniedError(path string, operation string) error {
	b := NewLocalizedError(ErrCodePermissionDenied, MsgErrPermissionDenied, operation, path)
	b.WithContext("path", path)
	b.WithContext("operation", operation)
	b.WithLocalizedSuggestion(MsgSugCheckPermissions, path)
	b.WithLocalizedSuggestion(MsgSugRunAsAdmin)
	return b.Build()
}

// LocalizedInvalidYAMLError creates a localized invalid YAML error
func LocalizedInvalidYAMLError(file string, line int, err error) error {
	b := NewLocalizedError(ErrCodeInvalidYAML, MsgErrInvalidYAML, file)
	b.WithLocalizedDetails(MsgDetailErrorAtLine, line, err)
	b.WithContext("file", file)
	b.WithContext("line", line)
	b.WithLocalizedSuggestion(MsgSugUseSpaces)
	b.WithWrapped(err)
	return b.Build()
}

// LocalizedMissingAPIKeyError creates a localized API key missing error
func LocalizedMissingAPIKeyError(provider string) error {
	b := NewLocalizedError(ErrCodeMissingAPIKey, MsgErrMissingAPIKey, provider)
	b.WithContext("provider", provider)
	b.WithLocalizedSuggestion(MsgSugSetAPIKey, provider)
	return b.Build()
}

// LocalizedOutOfMemoryError creates a localized out of memory error
func LocalizedOutOfMemoryError(fileSize int64) error {
	sizeMB := fileSize / 1024 / 1024
	b := NewLocalizedError(ErrCodeOutOfMemory, MsgErrOutOfMemory)
	b.WithLocalizedDetails(MsgDetailFileSize, sizeMB)
	b.WithContext("file_size_mb", sizeMB)
	b.WithLocalizedSuggestion(MsgSugUseStreaming)
	b.WithLocalizedSuggestion(MsgSugCloseApps)
	return b.Build()
}

// LocalizedInvalidFormatError creates a localized invalid format error
func LocalizedInvalidFormatError(path string, expected string, found string) error {
	b := NewLocalizedError(ErrCodeInvalidFormat, MsgErrInvalidFormat, path)
	b.WithLocalizedDetails(MsgDetailExpectedFormat, expected, found)
	b.WithContext("path", path)
	b.WithContext("expected", expected)
	b.WithContext("found", found)
	b.WithLocalizedSuggestion(MsgSugCheckFormat, expected)
	return b.Build()
}

// FormatError formats an error for display with proper localization
func FormatError(err error, verbose bool) string {
	// Check if it's an enhanced error
	var enhanced *EnhancedError
	if As(err, &enhanced) {
		if verbose {
			return enhanced.Error()
		}
		// In non-verbose mode, just show message and first suggestion
		if len(enhanced.Suggestions) > 0 {
			return fmt.Sprintf("%s\n  💡 %s", enhanced.Message, enhanced.Suggestions[0])
		}
		return enhanced.Message
	}
	
	// Fallback to standard error
	return err.Error()
}

// As is a wrapper around errors.As for convenience
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is is a wrapper around errors.Is for convenience  
func Is(err error, target error) bool {
	return errors.Is(err, target)
}