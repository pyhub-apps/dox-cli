package ui

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
	
	"github.com/fatih/color"
)

func TestColorDefinitions(t *testing.T) {
	// Test that color definitions are not nil
	tests := []struct {
		name  string
		color *color.Color
	}{
		{"Success", Success},
		{"Error", Error},
		{"Warning", Warning},
		{"Info", Info},
		{"Header", Header},
		{"Accent", Accent},
		{"Muted", Muted},
		{"DocxColor", DocxColor},
		{"PptxColor", PptxColor},
		{"MarkdownColor", MarkdownColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color == nil {
				t.Errorf("%s color is nil", tt.name)
			}
		})
	}
}

func TestIcons(t *testing.T) {
	// Test that icon-based functions work without panicking
	// We can't directly test the icon variables as they're private
	
	// Test ASCII fallback behavior
	t.Run("ASCII fallback", func(t *testing.T) {
		// Save original env
		origTerm := os.Getenv("TERM")
		origNoUnicode := os.Getenv("NO_UNICODE")
		defer func() {
			os.Setenv("TERM", origTerm)
			os.Setenv("NO_UNICODE", origNoUnicode)
		}()
		
		// Set dumb terminal to trigger ASCII mode
		os.Setenv("TERM", "dumb")
		
		// The icons are set during init(), so we can't test the change directly
		// But we can verify the print functions still work
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		PrintSuccess("Test")
		
		w.Close()
		os.Stdout = old
		
		out, _ := io.ReadAll(r)
		output := string(out)
		
		// Should contain the message regardless of icon type
		if !strings.Contains(output, "Test") {
			t.Error("Output should contain the test message")
		}
	})
}

func TestPrintFunctions(t *testing.T) {
	// Test that print functions don't panic
	tests := []struct {
		name string
		fn   func(string, ...interface{})
		msg  string
	}{
		{"PrintSuccess", PrintSuccess, "Success message"},
		{"PrintError", PrintError, "Error message"},
		{"PrintWarning", PrintWarning, "Warning message"},
		{"PrintInfo", PrintInfo, "Info message"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// PrintError and PrintWarning write to stderr, others to stdout
			var old *os.File
			var r, w *os.File
			
			if tt.name == "PrintError" || tt.name == "PrintWarning" {
				old = os.Stderr
				r, w, _ = os.Pipe()
				os.Stderr = w
			} else {
				old = os.Stdout
				r, w, _ = os.Pipe()
				os.Stdout = w
			}
			
			// Call function
			tt.fn(tt.msg)
			
			// Restore output
			w.Close()
			if tt.name == "PrintError" || tt.name == "PrintWarning" {
				os.Stderr = old
			} else {
				os.Stdout = old
			}
			
			// Read output
			out, _ := io.ReadAll(r)
			output := string(out)
			
			// Verify output contains message
			if !strings.Contains(output, tt.msg) {
				t.Errorf("Output should contain %q, got %q", tt.msg, output)
			}
		})
	}
}

func TestSpinner(t *testing.T) {
	t.Run("NewSpinner", func(t *testing.T) {
		spinner := NewSpinner("Testing")
		if spinner == nil {
			t.Fatal("NewSpinner returned nil")
		}
		
		// Check that it was created with -1 total (spinner mode)
		if spinner.total != -1 {
			t.Errorf("Expected spinner to have total -1, got %d", spinner.total)
		}
		
		// Finish the spinner
		spinner.Finish()
		
		// Give time for completion
		time.Sleep(100 * time.Millisecond)
	})
	
	t.Run("SetDescription", func(t *testing.T) {
		spinner := NewSpinner("Initial")
		spinner.SetDescription("Updated")
		
		// We can't directly test the description as the bar is private
		// but we can verify it doesn't panic
		
		spinner.Finish()
		time.Sleep(100 * time.Millisecond)
	})
	
	t.Run("Clear", func(t *testing.T) {
		spinner := NewSpinner("Test")
		time.Sleep(100 * time.Millisecond)
		
		// Clear should work without error
		spinner.Clear()
	})
}

func TestProgressBar(t *testing.T) {
	t.Run("NewProgressBar", func(t *testing.T) {
		pb := NewProgressBar(100, "Test Progress")
		if pb == nil {
			t.Fatal("NewProgressBar returned nil")
		}
		
		// Check total is set
		if pb.total != 100 {
			t.Errorf("Expected total 100, got %d", pb.total)
		}
		
		// Clean up
		pb.Finish()
	})
	
	t.Run("Increment", func(t *testing.T) {
		pb := NewProgressBar(10, "Test Increment")
		
		// Increment should not panic
		pb.Increment()
		pb.IncrementBy(3)
		
		pb.Finish()
	})
	
	t.Run("SetDescription", func(t *testing.T) {
		pb := NewProgressBar(50, "Initial")
		
		// SetDescription should not panic
		pb.SetDescription("Updated Description")
		
		pb.Finish()
	})
	
	t.Run("Clear", func(t *testing.T) {
		pb := NewProgressBar(20, "Test Clear")
		
		// Clear should not panic
		pb.Clear()
	})
}

func TestMultiProgressManager(t *testing.T) {
	t.Run("NewMultiProgressManager", func(t *testing.T) {
		manager := NewMultiProgressManager()
		
		if manager == nil {
			t.Fatal("NewMultiProgressManager returned nil")
		}
		
		// Check initialization
		if len(manager.bars) != 0 {
			t.Errorf("Expected empty bars slice, got %d bars", len(manager.bars))
		}
	})
	
	t.Run("AddBar", func(t *testing.T) {
		manager := NewMultiProgressManager()
		bar := manager.AddBar(100, "Test Bar")
		
		if bar == nil {
			t.Error("AddBar returned nil")
		}
		
		if len(manager.bars) != 1 {
			t.Errorf("Expected 1 bar, got %d", len(manager.bars))
		}
	})
	
	t.Run("AddSpinner", func(t *testing.T) {
		manager := NewMultiProgressManager()
		spinner := manager.AddSpinner("Test Spinner")
		
		if spinner == nil {
			t.Error("AddSpinner returned nil")
		}
		
		if len(manager.bars) != 1 {
			t.Errorf("Expected 1 spinner, got %d", len(manager.bars))
		}
	})
	
	t.Run("Wait", func(t *testing.T) {
		manager := NewMultiProgressManager()
		// Wait should not panic even with no bars
		manager.Wait()
	})
}

func TestPrintFunctionsExtended(t *testing.T) {
	// These tests verify that functions don't panic
	// Note: Color output goes directly to terminal and is hard to capture
	
	t.Run("PrintHeader", func(t *testing.T) {
		// Verify PrintHeader doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintHeader panicked: %v", r)
			}
		}()
		
		PrintHeader("Test Header")
	})
	
	t.Run("PrintStep", func(t *testing.T) {
		// Verify PrintStep doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintStep panicked: %v", r)
			}
		}()
		
		PrintStep(1, 5, "Processing item")
	})
	
	t.Run("PrintFileOperation", func(t *testing.T) {
		// Verify PrintFileOperation doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintFileOperation panicked: %v", r)
			}
		}()
		
		PrintFileOperation("Processing", "test.docx", ".docx")
		PrintFileOperation("Converting", "test.pptx", ".pptx")
		PrintFileOperation("Reading", "test.md", ".md")
		PrintFileOperation("Writing", "test.txt", ".txt") // Unknown type
	})
}

func TestPrintSummary(t *testing.T) {
	// Verify PrintSummary doesn't panic with various data types
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSummary panicked: %v", r)
		}
	}()
	
	stats := map[string]interface{}{
		"Total Files": 10,
		"Processed": 8,
		"Failed": 2,
		"Status": "Complete",
		"Success": true,
		"Zero Value": 0,
		"Negative": -5,
		"Float": 3.14,
		"False": false,
	}
	
	PrintSummary("Test Summary", stats)
}

func TestConfirmation(t *testing.T) {
	// Test Confirmation function with various inputs
	// Note: This test is limited as it requires user input
	// We can only test that it doesn't panic with empty input
	
	t.Run("EmptyInput", func(t *testing.T) {
		// Redirect stdin to provide empty input
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		
		// Write newline to simulate pressing Enter
		w.Write([]byte("\n"))
		w.Close()
		
		// Suppress output
		oldStderr := os.Stderr
		os.Stderr, _ = os.Open(os.DevNull)
		
		result := Confirmation("Test prompt")
		
		// Restore
		os.Stdin = oldStdin
		os.Stderr = oldStderr
		
		// Empty input should return false
		if result {
			t.Error("Confirmation should return false for empty input")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("FormatFileSize", func(t *testing.T) {
		tests := []struct {
			bytes    int64
			expected string
		}{
			{0, "0 B"},
			{100, "100 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{1048576, "1.0 MB"},
			{1073741824, "1.0 GB"},
		}
		
		for _, tt := range tests {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		}
	})
	
	t.Run("FormatDuration", func(t *testing.T) {
		tests := []struct {
			duration time.Duration
			contains string
		}{
			{500 * time.Millisecond, "500ms"},
			{2 * time.Second, "2.0s"},
			{65 * time.Second, "1m 5s"},
			{125 * time.Minute, "2h 5m"},
		}
		
		for _, tt := range tests {
			result := FormatDuration(tt.duration)
			if result != tt.contains {
				t.Errorf("FormatDuration(%v) = %s, want %s", tt.duration, result, tt.contains)
			}
		}
	})
	
	t.Run("ColorControl", func(t *testing.T) {
		// Test EnableColor
		EnableColor()
		if !IsColorEnabled() {
			t.Error("EnableColor should enable color output")
		}
		
		// Test DisableColor
		DisableColor()
		if IsColorEnabled() {
			t.Error("DisableColor should disable color output")
		}
		
		// Restore original state
		EnableColor()
	})
}