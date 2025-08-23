package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestFileError(t *testing.T) {
	t.Run("Error message", func(t *testing.T) {
		fe := &FileError{
			Path:      "/test/file.txt",
			Operation: "reading",
			Err:       ErrFileNotFound,
		}
		
		expected := "reading failed for '/test/file.txt': file not found"
		if fe.Error() != expected {
			t.Errorf("FileError.Error() = %q, want %q", fe.Error(), expected)
		}
	})
	
	t.Run("Unwrap", func(t *testing.T) {
		baseErr := ErrPermissionDenied
		fe := &FileError{
			Path:      "/test/file.txt",
			Operation: "writing",
			Err:       baseErr,
		}
		
		if fe.Unwrap() != baseErr {
			t.Errorf("FileError.Unwrap() = %v, want %v", fe.Unwrap(), baseErr)
		}
	})
}

func TestNewFileError(t *testing.T) {
	err := NewFileError("/path/to/file", "opening", ErrFileNotFound)
	
	fe, ok := err.(*FileError)
	if !ok {
		t.Fatal("NewFileError() should return *FileError")
	}
	
	if fe.Path != "/path/to/file" {
		t.Errorf("Path = %q, want %q", fe.Path, "/path/to/file")
	}
	if fe.Operation != "opening" {
		t.Errorf("Operation = %q, want %q", fe.Operation, "opening")
	}
	if fe.Err != ErrFileNotFound {
		t.Errorf("Err = %v, want %v", fe.Err, ErrFileNotFound)
	}
}

func TestDocumentError(t *testing.T) {
	t.Run("Error message with path", func(t *testing.T) {
		de := &DocumentError{
			Path:   "/doc.docx",
			Type:   "docx",
			Reason: "corrupted file",
			Err:    ErrDocumentCorrupted,
		}
		
		msg := de.Error()
		if !strings.Contains(msg, "/doc.docx") {
			t.Errorf("Error message should contain path, got: %s", msg)
		}
		if !strings.Contains(msg, "docx") {
			t.Errorf("Error message should contain type, got: %s", msg)
		}
		if !strings.Contains(msg, "corrupted file") {
			t.Errorf("Error message should contain reason, got: %s", msg)
		}
	})
	
	t.Run("Error message without path", func(t *testing.T) {
		de := &DocumentError{
			Path:   "",
			Type:   "pptx",
			Reason: "invalid structure",
			Err:    ErrDocumentCorrupted,
		}
		
		msg := de.Error()
		if strings.Contains(msg, "''") {
			t.Errorf("Error message should not contain empty path, got: %s", msg)
		}
		if !strings.Contains(msg, "pptx") {
			t.Errorf("Error message should contain type, got: %s", msg)
		}
		if !strings.Contains(msg, "invalid structure") {
			t.Errorf("Error message should contain reason, got: %s", msg)
		}
	})
	
	t.Run("Unwrap", func(t *testing.T) {
		baseErr := ErrUnsupportedFormat
		de := &DocumentError{
			Path:   "/doc.pptx",
			Type:   "pptx",
			Reason: "unsupported",
			Err:    baseErr,
		}
		
		if de.Unwrap() != baseErr {
			t.Errorf("DocumentError.Unwrap() = %v, want %v", de.Unwrap(), baseErr)
		}
	})
}

func TestNewDocumentError(t *testing.T) {
	err := NewDocumentError("/test.docx", "docx", "invalid structure", ErrDocumentCorrupted)
	
	de, ok := err.(*DocumentError)
	if !ok {
		t.Fatal("NewDocumentError() should return *DocumentError")
	}
	
	if de.Path != "/test.docx" {
		t.Errorf("Path = %q, want %q", de.Path, "/test.docx")
	}
	if de.Type != "docx" {
		t.Errorf("Type = %q, want %q", de.Type, "docx")
	}
	if de.Reason != "invalid structure" {
		t.Errorf("Reason = %q, want %q", de.Reason, "invalid structure")
	}
	if de.Err != ErrDocumentCorrupted {
		t.Errorf("Err = %v, want %v", de.Err, ErrDocumentCorrupted)
	}
}

func TestValidationError(t *testing.T) {
	t.Run("Error message with value", func(t *testing.T) {
		ve := &ValidationError{
			Field:   "email",
			Value:   "invalid",
			Message: "invalid email format",
		}
		
		msg := ve.Error()
		if !strings.Contains(msg, "email") {
			t.Errorf("Error message should contain field, got: %s", msg)
		}
		if !strings.Contains(msg, "invalid email format") {
			t.Errorf("Error message should contain message, got: %s", msg)
		}
		if !strings.Contains(msg, "invalid") {
			t.Errorf("Error message should contain value, got: %s", msg)
		}
	})
	
	t.Run("Error message without value", func(t *testing.T) {
		ve := &ValidationError{
			Field:   "age",
			Value:   nil,
			Message: "is required",
		}
		
		msg := ve.Error()
		if !strings.Contains(msg, "age") {
			t.Errorf("Error message should contain field, got: %s", msg)
		}
		if !strings.Contains(msg, "is required") {
			t.Errorf("Error message should contain message, got: %s", msg)
		}
		if strings.Contains(msg, "<nil>") {
			t.Errorf("Error message should not contain nil value, got: %s", msg)
		}
	})
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("username", "test user", "username already exists")
	
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("NewValidationError() should return *ValidationError")
	}
	
	if ve.Field != "username" {
		t.Errorf("Field = %q, want %q", ve.Field, "username")
	}
	if ve.Value != "test user" {
		t.Errorf("Value = %q, want %q", ve.Value, "test user")
	}
	if ve.Message != "username already exists" {
		t.Errorf("Message = %q, want %q", ve.Message, "username already exists")
	}
}

func TestConfigError(t *testing.T) {
	t.Run("Error message with file", func(t *testing.T) {
		ce := &ConfigError{
			File:   "/config.yml",
			Reason: "invalid YAML syntax",
			Err:    ErrInvalidConfig,
		}
		
		msg := ce.Error()
		if !strings.Contains(msg, "/config.yml") {
			t.Errorf("Error message should contain file, got: %s", msg)
		}
		if !strings.Contains(msg, "invalid YAML syntax") {
			t.Errorf("Error message should contain reason, got: %s", msg)
		}
	})
	
	t.Run("Error message without file", func(t *testing.T) {
		ce := &ConfigError{
			File:   "",
			Reason: "missing required field",
			Err:    ErrInvalidConfig,
		}
		
		msg := ce.Error()
		if strings.Contains(msg, "''") {
			t.Errorf("Error message should not contain empty file, got: %s", msg)
		}
		if !strings.Contains(msg, "missing required field") {
			t.Errorf("Error message should contain reason, got: %s", msg)
		}
	})
	
	t.Run("Unwrap", func(t *testing.T) {
		baseErr := ErrConfigNotFound
		ce := &ConfigError{
			File:   "/config.json",
			Reason: "not found",
			Err:    baseErr,
		}
		
		if ce.Unwrap() != baseErr {
			t.Errorf("ConfigError.Unwrap() = %v, want %v", ce.Unwrap(), baseErr)
		}
	})
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("/app/config.yml", "missing required field", ErrInvalidConfig)
	
	ce, ok := err.(*ConfigError)
	if !ok {
		t.Fatal("NewConfigError() should return *ConfigError")
	}
	
	if ce.File != "/app/config.yml" {
		t.Errorf("File = %q, want %q", ce.File, "/app/config.yml")
	}
	if ce.Reason != "missing required field" {
		t.Errorf("Reason = %q, want %q", ce.Reason, "missing required field")
	}
	if ce.Err != ErrInvalidConfig {
		t.Errorf("Err = %v, want %v", ce.Err, ErrInvalidConfig)
	}
}

func TestErrorCheckers(t *testing.T) {
	t.Run("IsFileNotFound", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			want bool
		}{
			{"Direct match", ErrFileNotFound, true},
			{"Wrapped in FileError", NewFileError("/test", "read", ErrFileNotFound), true},
			{"Different error", ErrPermissionDenied, false},
			{"Nil error", nil, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IsFileNotFound(tt.err); got != tt.want {
					t.Errorf("IsFileNotFound(%v) = %v, want %v", tt.err, got, tt.want)
				}
			})
		}
	})
	
	t.Run("IsPermissionDenied", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			want bool
		}{
			{"Direct match", ErrPermissionDenied, true},
			{"Wrapped in FileError", NewFileError("/test", "write", ErrPermissionDenied), true},
			{"Different error", ErrFileNotFound, false},
			{"Nil error", nil, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IsPermissionDenied(tt.err); got != tt.want {
					t.Errorf("IsPermissionDenied(%v) = %v, want %v", tt.err, got, tt.want)
				}
			})
		}
	})
	
	t.Run("IsInvalidFormat", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			want bool
		}{
			{"Direct match", ErrInvalidFormat, true},
			{"Wrapped in DocumentError", NewDocumentError("/test.doc", "doc", "bad", ErrInvalidFormat), true},
			{"Different error", ErrFileNotFound, false},
			{"Nil error", nil, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IsInvalidFormat(tt.err); got != tt.want {
					t.Errorf("IsInvalidFormat(%v) = %v, want %v", tt.err, got, tt.want)
				}
			})
		}
	})
	
	t.Run("IsValidationError", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			want bool
		}{
			{"ValidationError type", NewValidationError("field", "value", "msg"), true},
			{"Different error type", NewFileError("/test", "read", ErrFileNotFound), false},
			{"Plain error", errors.New("test error"), false},
			{"Nil error", nil, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IsValidationError(tt.err); got != tt.want {
					t.Errorf("IsValidationError(%v) = %v, want %v", tt.err, got, tt.want)
				}
			})
		}
	})
}