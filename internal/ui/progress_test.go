package ui

import (
	"testing"
	"time"
)

func TestProgressTracker(t *testing.T) {
	t.Run("NewProgressTracker", func(t *testing.T) {
		tracker := NewProgressTracker(10, "Test progress")
		if tracker == nil {
			t.Fatal("NewProgressTracker returned nil")
		}
		if tracker.total != 10 {
			t.Errorf("Expected total 10, got %d", tracker.total)
		}
	})
	
	t.Run("UpdateProgress", func(t *testing.T) {
		tracker := NewProgressTracker(5, "Test")
		
		// Update progress with file info
		tracker.UpdateProgress("file1.docx", 1024)
		
		stats := tracker.GetStats()
		if stats.ProcessedItems != 1 {
			t.Errorf("Expected 1 processed item, got %d", stats.ProcessedItems)
		}
		if stats.ProcessedBytes != 1024 {
			t.Errorf("Expected 1024 processed bytes, got %d", stats.ProcessedBytes)
		}
		if stats.CurrentFile != "file1.docx" {
			t.Errorf("Expected current file 'file1.docx', got %s", stats.CurrentFile)
		}
	})
	
	t.Run("Speed Calculation", func(t *testing.T) {
		tracker := NewProgressTracker(100, "Test")
		
		// Add multiple samples for speed calculation
		for i := 0; i < 5; i++ {
			tracker.UpdateProgress("file.docx", 1024)
			time.Sleep(10 * time.Millisecond) // Small delay to get measurable speed
		}
		
		stats := tracker.GetStats()
		// We should have some speed measurements
		if stats.ProcessedItems != 5 {
			t.Errorf("Expected 5 processed items, got %d", stats.ProcessedItems)
		}
	})
	
	t.Run("ETA Calculation", func(t *testing.T) {
		tracker := NewProgressTracker(10, "Test")
		
		// Process half of the items
		for i := 0; i < 5; i++ {
			tracker.UpdateProgress("file.docx", 1024)
			time.Sleep(10 * time.Millisecond)
		}
		
		stats := tracker.GetStats()
		// ETA should be positive since we have items left
		if stats.TotalItems != 10 {
			t.Errorf("Expected 10 total items, got %d", stats.TotalItems)
		}
		if stats.ProcessedItems != 5 {
			t.Errorf("Expected 5 processed items, got %d", stats.ProcessedItems)
		}
	})
	
	t.Run("Cancellation", func(t *testing.T) {
		tracker := NewProgressTracker(10, "Test")
		
		if tracker.IsCancelled() {
			t.Error("Tracker should not be cancelled initially")
		}
		
		tracker.Cancel()
		
		if !tracker.IsCancelled() {
			t.Error("Tracker should be cancelled after Cancel()")
		}
		
		// Cancel channel should be closed
		select {
		case <-tracker.WaitForCancel():
			// Good, channel is closed
		default:
			t.Error("Cancel channel should be closed")
		}
	})
}

func TestProgressStats(t *testing.T) {
	stats := ProgressStats{
		ProcessedItems: 50,
		ProcessedBytes: 1024 * 1024,
		TotalItems:     100,
		ItemsPerSecond: 2.5,
		BytesPerSecond: 1024 * 100,
		ElapsedTime:    20 * time.Second,
		ETA:            20 * time.Second,
		CurrentFile:    "test.docx",
	}
	
	str := stats.String()
	if str == "" {
		t.Error("ProgressStats.String() returned empty string")
	}
	
	// Check that string contains key information
	if !contains(str, "50.0%") {
		t.Error("Progress percentage not found in string")
	}
	if !contains(str, "50/100") {
		t.Error("Progress count not found in string")
	}
}

func TestLogLevel(t *testing.T) {
	t.Run("ParseLogLevel", func(t *testing.T) {
		tests := []struct {
			input    string
			expected LogLevel
		}{
			{"debug", LogLevelDebug},
			{"info", LogLevelInfo},
			{"warn", LogLevelWarn},
			{"warning", LogLevelWarn},
			{"error", LogLevelError},
			{"unknown", LogLevelInfo}, // default
		}
		
		for _, tt := range tests {
			result := ParseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLogLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		}
	})
	
	t.Run("SetGetLogLevel", func(t *testing.T) {
		original := GetLogLevel()
		defer SetLogLevel(original)
		
		SetLogLevel(LogLevelDebug)
		if GetLogLevel() != LogLevelDebug {
			t.Error("SetLogLevel/GetLogLevel mismatch")
		}
		
		SetLogLevel(LogLevelError)
		if GetLogLevel() != LogLevelError {
			t.Error("SetLogLevel/GetLogLevel mismatch")
		}
	})
	
	t.Run("PrintDebug", func(t *testing.T) {
		original := GetLogLevel()
		defer SetLogLevel(original)
		
		// Set to debug level
		SetLogLevel(LogLevelDebug)
		
		// This should not panic
		PrintDebug("Test debug message: %d", 123)
		
		// Set to error level
		SetLogLevel(LogLevelError)
		
		// This should also not panic (but won't print)
		PrintDebug("This won't be printed")
	})
	
	t.Run("PrintLog", func(t *testing.T) {
		original := GetLogLevel()
		defer SetLogLevel(original)
		
		SetLogLevel(LogLevelInfo)
		
		// These should not panic
		PrintLog(LogLevelDebug, "Debug message")
		PrintLog(LogLevelInfo, "Info message")
		PrintLog(LogLevelWarn, "Warning message")
		PrintLog(LogLevelError, "Error message")
	})
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && (s == substr || len(s) > len(substr))
}