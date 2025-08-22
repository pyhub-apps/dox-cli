package errors

import (
	"errors"
	"fmt"
)

// Common sentinel errors
var (
	// File operation errors
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidFormat     = errors.New("invalid file format")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrFileAlreadyExists = errors.New("file already exists")
	
	// Document processing errors
	ErrDocumentCorrupted = errors.New("document is corrupted or invalid")
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrEmptyDocument     = errors.New("document is empty")
	
	// Configuration errors
	ErrConfigNotFound  = errors.New("configuration file not found")
	ErrInvalidConfig   = errors.New("invalid configuration")
	ErrMissingAPIKey   = errors.New("API key is missing")
	
	// Validation errors
	ErrInvalidInput    = errors.New("invalid input")
	ErrMissingRequired = errors.New("required parameter missing")
)

// FileError represents a file operation error with context
type FileError struct {
	Path      string
	Operation string
	Err       error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("%s failed for '%s': %v", e.Operation, e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

// NewFileError creates a new FileError
func NewFileError(path, operation string, err error) error {
	return &FileError{
		Path:      path,
		Operation: operation,
		Err:       err,
	}
}

// DocumentError represents a document processing error
type DocumentError struct {
	Path   string
	Type   string
	Reason string
	Err    error
}

func (e *DocumentError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("document error (%s) in '%s': %s", e.Type, e.Path, e.Reason)
	}
	return fmt.Sprintf("document error (%s): %s", e.Type, e.Reason)
}

func (e *DocumentError) Unwrap() error {
	return e.Err
}

// NewDocumentError creates a new DocumentError
func NewDocumentError(path, docType, reason string, err error) error {
	return &DocumentError{
		Path:   path,
		Type:   docType,
		Reason: reason,
		Err:    err,
	}
}

// ValidationError represents input validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation failed for %s with value '%v': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, message string) error {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	File   string
	Reason string
	Err    error
}

func (e *ConfigError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("config error in '%s': %s", e.File, e.Reason)
	}
	return fmt.Sprintf("config error: %s", e.Reason)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(file, reason string, err error) error {
	return &ConfigError{
		File:   file,
		Reason: reason,
		Err:    err,
	}
}

// Helper functions for error checking

// IsFileNotFound checks if an error is a file not found error
func IsFileNotFound(err error) bool {
	return errors.Is(err, ErrFileNotFound)
}

// IsPermissionDenied checks if an error is a permission denied error
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsInvalidFormat checks if an error is an invalid format error
func IsInvalidFormat(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}