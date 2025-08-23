package ui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ProgressTracker provides enhanced progress tracking with ETA and speed metrics
type ProgressTracker struct {
	*ProgressBar
	
	// Statistics
	startTime     time.Time
	processedItems int64
	processedBytes int64
	currentFile    string
	
	// Speed tracking
	recentItems    []int64 // Recent item counts for speed calculation
	recentBytes    []int64 // Recent byte counts for speed calculation
	recentTimes    []time.Time
	speedWindow    int // Number of samples to keep for speed calculation
	
	// Control
	cancelChan     chan struct{}
	cancelled      atomic.Bool
	mu             sync.RWMutex
}

// NewProgressTracker creates a new enhanced progress tracker
func NewProgressTracker(total int, description string) *ProgressTracker {
	return &ProgressTracker{
		ProgressBar:  NewProgressBar(total, description),
		startTime:    time.Now(),
		speedWindow:  10, // Keep last 10 samples for speed calculation
		recentItems:  make([]int64, 0, 10),
		recentBytes:  make([]int64, 0, 10),
		recentTimes:  make([]time.Time, 0, 10),
		cancelChan:   make(chan struct{}),
	}
}

// UpdateProgress updates progress with current file information
func (pt *ProgressTracker) UpdateProgress(currentFile string, bytesProcessed int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.currentFile = currentFile
	atomic.AddInt64(&pt.processedItems, 1)
	atomic.AddInt64(&pt.processedBytes, bytesProcessed)
	
	// Update recent samples for speed calculation
	now := time.Now()
	pt.recentItems = append(pt.recentItems, atomic.LoadInt64(&pt.processedItems))
	pt.recentBytes = append(pt.recentBytes, atomic.LoadInt64(&pt.processedBytes))
	pt.recentTimes = append(pt.recentTimes, now)
	
	// Keep only the last speedWindow samples
	if len(pt.recentItems) > pt.speedWindow {
		pt.recentItems = pt.recentItems[1:]
		pt.recentBytes = pt.recentBytes[1:]
		pt.recentTimes = pt.recentTimes[1:]
	}
	
	// Update description with current file and metrics
	description := pt.formatDescription()
	pt.SetDescription(description)
	pt.Increment()
}

// formatDescription formats the progress bar description with metrics
func (pt *ProgressTracker) formatDescription() string {
	itemsPerSec := pt.calculateItemsPerSecond()
	bytesPerSec := pt.calculateBytesPerSecond()
	eta := pt.calculateETA()
	
	desc := fmt.Sprintf("Processing: %s", pt.currentFile)
	
	if itemsPerSec > 0 {
		desc += fmt.Sprintf(" | %.1f files/s", itemsPerSec)
	}
	
	if bytesPerSec > 0 {
		desc += fmt.Sprintf(" | %s/s", FormatFileSize(int64(bytesPerSec)))
	}
	
	if eta > 0 {
		desc += fmt.Sprintf(" | ETA: %s", FormatDuration(eta))
	}
	
	return desc
}

// calculateItemsPerSecond calculates the current processing speed in items/second
func (pt *ProgressTracker) calculateItemsPerSecond() float64 {
	if len(pt.recentItems) < 2 {
		return 0
	}
	
	firstIdx := 0
	lastIdx := len(pt.recentItems) - 1
	
	itemsDiff := pt.recentItems[lastIdx] - pt.recentItems[firstIdx]
	timeDiff := pt.recentTimes[lastIdx].Sub(pt.recentTimes[firstIdx]).Seconds()
	
	if timeDiff == 0 {
		return 0
	}
	
	return float64(itemsDiff) / timeDiff
}

// calculateBytesPerSecond calculates the current processing speed in bytes/second
func (pt *ProgressTracker) calculateBytesPerSecond() float64 {
	if len(pt.recentBytes) < 2 {
		return 0
	}
	
	firstIdx := 0
	lastIdx := len(pt.recentBytes) - 1
	
	bytesDiff := pt.recentBytes[lastIdx] - pt.recentBytes[firstIdx]
	timeDiff := pt.recentTimes[lastIdx].Sub(pt.recentTimes[firstIdx]).Seconds()
	
	if timeDiff == 0 {
		return 0
	}
	
	return float64(bytesDiff) / timeDiff
}

// calculateETA calculates the estimated time of arrival
func (pt *ProgressTracker) calculateETA() time.Duration {
	processed := atomic.LoadInt64(&pt.processedItems)
	if processed == 0 || pt.total <= 0 {
		return 0
	}
	
	elapsed := time.Since(pt.startTime)
	avgTimePerItem := elapsed / time.Duration(processed)
	remaining := int64(pt.total) - processed
	
	if remaining <= 0 {
		return 0
	}
	
	return avgTimePerItem * time.Duration(remaining)
}

// GetStats returns current progress statistics
func (pt *ProgressTracker) GetStats() ProgressStats {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	return ProgressStats{
		ProcessedItems: atomic.LoadInt64(&pt.processedItems),
		ProcessedBytes: atomic.LoadInt64(&pt.processedBytes),
		TotalItems:     int64(pt.total),
		ItemsPerSecond: pt.calculateItemsPerSecond(),
		BytesPerSecond: pt.calculateBytesPerSecond(),
		ElapsedTime:    time.Since(pt.startTime),
		ETA:            pt.calculateETA(),
		CurrentFile:    pt.currentFile,
	}
}

// IsCancelled returns whether the operation was cancelled
func (pt *ProgressTracker) IsCancelled() bool {
	return pt.cancelled.Load()
}

// Cancel cancels the operation
func (pt *ProgressTracker) Cancel() {
	if pt.cancelled.CompareAndSwap(false, true) {
		close(pt.cancelChan)
		pt.SetDescription("Cancelling... Please wait for current operation to complete")
	}
}

// WaitForCancel returns a channel that's closed when cancelled
func (pt *ProgressTracker) WaitForCancel() <-chan struct{} {
	return pt.cancelChan
}

// SetupGracefulShutdown sets up signal handlers for graceful shutdown
func (pt *ProgressTracker) SetupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		pt.Cancel()
	}()
}

// ProgressStats contains progress statistics
type ProgressStats struct {
	ProcessedItems int64
	ProcessedBytes int64
	TotalItems     int64
	ItemsPerSecond float64
	BytesPerSecond float64
	ElapsedTime    time.Duration
	ETA            time.Duration
	CurrentFile    string
}

// String formats progress stats as a string
func (ps ProgressStats) String() string {
	progress := float64(ps.ProcessedItems) / float64(ps.TotalItems) * 100
	
	return fmt.Sprintf(
		"Progress: %.1f%% (%d/%d files) | Speed: %.1f files/s, %s/s | Elapsed: %s | ETA: %s",
		progress,
		ps.ProcessedItems,
		ps.TotalItems,
		ps.ItemsPerSecond,
		FormatFileSize(int64(ps.BytesPerSecond)),
		FormatDuration(ps.ElapsedTime),
		FormatDuration(ps.ETA),
	)
}

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var (
	currentLogLevel = LogLevelInfo
	logLevelMu     sync.RWMutex
)

// SetLogLevel sets the global log level
func SetLogLevel(level LogLevel) {
	logLevelMu.Lock()
	defer logLevelMu.Unlock()
	currentLogLevel = level
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	logLevelMu.RLock()
	defer logLevelMu.RUnlock()
	return currentLogLevel
}

// ParseLogLevel parses a string log level
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// PrintDebug prints a debug message if log level allows
func PrintDebug(format string, args ...interface{}) {
	if GetLogLevel() <= LogLevelDebug {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "%s [DEBUG] %s\n", time.Now().Format("15:04:05"), msg)
	}
}

// PrintLog prints a log message based on level
func PrintLog(level LogLevel, format string, args ...interface{}) {
	if GetLogLevel() <= level {
		levelStr := ""
		switch level {
		case LogLevelDebug:
			levelStr = "DEBUG"
		case LogLevelInfo:
			levelStr = "INFO"
		case LogLevelWarn:
			levelStr = "WARN"
		case LogLevelError:
			levelStr = "ERROR"
		}
		
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "%s [%s] %s\n", time.Now().Format("15:04:05"), levelStr, msg)
	}
}