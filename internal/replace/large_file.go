package replace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/document"
	"github.com/pyhub/pyhub-docs/internal/ui"
)

// LargeFileOptions contains options for processing large files
type LargeFileOptions struct {
	// EnableStreaming enables streaming mode for large files
	EnableStreaming bool
	// FileSizeThreshold is the file size threshold for using streaming (default: 10MB)
	FileSizeThreshold int64
	// ShowMemoryUsage shows memory usage during processing
	ShowMemoryUsage bool
	// EnableMemoryMonitor enables memory monitoring
	EnableMemoryMonitor bool
}

// DefaultLargeFileOptions returns default options for large file processing
func DefaultLargeFileOptions() *LargeFileOptions {
	return &LargeFileOptions{
		EnableStreaming:     true,
		FileSizeThreshold:   10 * 1024 * 1024, // 10MB
		ShowMemoryUsage:     true,
		EnableMemoryMonitor: true,
	}
}

// ProcessLargeFile processes a potentially large file with optimizations
func ProcessLargeFile(filePath string, rules []Rule, opts *LargeFileOptions) (*ReplaceResult, error) {
	if opts == nil {
		opts = DefaultLargeFileOptions()
	}
	
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	fileSize := fileInfo.Size()
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Determine if we should use streaming
	useStreaming := opts.EnableStreaming && fileSize > opts.FileSizeThreshold
	
	if opts.ShowMemoryUsage {
		ui.PrintInfo("Processing %s (size: %s)", filePath, document.FormatBytes(uint64(fileSize)))
		if useStreaming {
			ui.PrintInfo("Using streaming mode for large file processing")
		}
	}
	
	// Start memory monitor if enabled
	var monitor *document.MemoryMonitor
	if opts.EnableMemoryMonitor {
		monitor = document.NewMemoryMonitor()
		monitor.SetThresholds(
			uint64(opts.FileSizeThreshold*5),  // Warning at 5x file size
			uint64(opts.FileSizeThreshold*10), // Critical at 10x file size
		)
		monitor.SetAlertHandler(func(level string, usage uint64, limit uint64) {
			ui.PrintWarning("[%s] Memory usage: %s / %s (%.1f%%)",
				level,
				document.FormatBytes(usage),
				document.FormatBytes(limit),
				float64(usage)/float64(limit)*100)
		})
		monitor.Start()
		defer monitor.Stop()
	}
	
	result := &ReplaceResult{
		FilePath:     filePath,
		Success:      false,
		Replacements: 0,
	}
	
	// Process based on file type and size
	switch ext {
	case ".docx":
		if useStreaming {
			result, err = processWordDocumentStreaming(filePath, rules, fileSize)
		} else {
			result, err = processWordDocumentStandard(filePath, rules)
		}
		
	case ".pptx":
		if useStreaming {
			result, err = processPowerPointDocumentStreaming(filePath, rules, fileSize)
		} else {
			result, err = processPowerPointDocumentStandard(filePath, rules)
		}
		
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
	
	// Show final memory stats if monitoring
	if monitor != nil && opts.ShowMemoryUsage {
		stats := monitor.GetStats()
		ui.PrintInfo("Memory stats - Peak: %s, Avg: %s",
			document.FormatBytes(stats.PeakUsage),
			document.FormatBytes(stats.AvgUsage))
	}
	
	return result, err
}

// processWordDocumentStreaming processes a Word document using streaming
func processWordDocumentStreaming(filePath string, rules []Rule, fileSize int64) (*ReplaceResult, error) {
	// Get adaptive options based on file size
	streamOpts := document.AdaptiveStreamingOptions(fileSize)
	
	// Open document in streaming mode
	doc, err := document.OpenWordDocumentStreaming(filePath, streamOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open document for streaming: %w", err)
	}
	defer doc.Close()
	
	result := &ReplaceResult{
		FilePath:     filePath,
		Success:      true,
		Replacements: 0,
	}
	
	// Apply each rule using streaming
	for _, rule := range rules {
		count, err := doc.ReplaceTextStreaming(rule.Old, rule.New)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
		result.Replacements += count
	}
	
	return result, nil
}

// processWordDocumentStandard processes a Word document using standard method
func processWordDocumentStandard(filePath string, rules []Rule) (*ReplaceResult, error) {
	// Use the existing standard processing
	doc, err := document.OpenWordDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	defer doc.Close()
	
	result := &ReplaceResult{
		FilePath:     filePath,
		Success:      true,
		Replacements: 0,
	}
	
	// Apply each rule
	for _, rule := range rules {
		err := doc.ReplaceText(rule.Old, rule.New)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
		// Note: The standard ReplaceText doesn't return count
		// We increment by 1 for each successful rule application
		result.Replacements++
	}
	
	// Save document
	if result.Replacements > 0 {
		if err := doc.Save(); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
	}
	
	return result, nil
}

// processPowerPointDocumentStreaming processes a PowerPoint document using streaming
func processPowerPointDocumentStreaming(filePath string, rules []Rule, fileSize int64) (*ReplaceResult, error) {
	// Get adaptive options based on file size
	streamOpts := document.AdaptiveStreamingOptions(fileSize)
	
	// Open document in streaming mode
	doc, err := document.OpenPowerPointDocumentStreaming(filePath, streamOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open presentation for streaming: %w", err)
	}
	defer doc.Close()
	
	result := &ReplaceResult{
		FilePath:     filePath,
		Success:      true,
		Replacements: 0,
	}
	
	// Apply each rule using streaming
	for _, rule := range rules {
		count, err := doc.ReplaceTextInSlidesStreaming(rule.Old, rule.New)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
		result.Replacements += count
	}
	
	return result, nil
}

// processPowerPointDocumentStandard processes a PowerPoint document using standard method
func processPowerPointDocumentStandard(filePath string, rules []Rule) (*ReplaceResult, error) {
	// Use the existing standard processing
	doc, err := document.OpenPowerPointDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open presentation: %w", err)
	}
	defer doc.Close()
	
	result := &ReplaceResult{
		FilePath:     filePath,
		Success:      true,
		Replacements: 0,
	}
	
	// Apply each rule
	for _, rule := range rules {
		err := doc.ReplaceText(rule.Old, rule.New)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
		// Note: The standard ReplaceText doesn't return count
		// We increment by 1 for each successful rule application
		result.Replacements++
	}
	
	// Save document
	if result.Replacements > 0 {
		if err := doc.Save(); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}
	}
	
	return result, nil
}

// EstimateMemoryUsage estimates memory usage for processing a file
func EstimateMemoryUsage(filePath string) (uint64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	
	// Rough estimate: XML expansion can be 5-10x the compressed size
	// Add buffer for processing overhead
	estimated := uint64(fileInfo.Size()) * 10
	
	return estimated, nil
}

// GetRecommendedOptions returns recommended options based on file characteristics
func GetRecommendedOptions(filePath string) (*LargeFileOptions, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	
	opts := DefaultLargeFileOptions()
	fileSize := fileInfo.Size()
	
	// Adjust thresholds based on file size
	switch {
	case fileSize < 1*1024*1024: // < 1MB
		opts.EnableStreaming = false
		opts.EnableMemoryMonitor = false
		
	case fileSize < 10*1024*1024: // < 10MB
		opts.EnableStreaming = false
		opts.EnableMemoryMonitor = true
		
	case fileSize < 50*1024*1024: // < 50MB
		opts.EnableStreaming = true
		opts.EnableMemoryMonitor = true
		opts.FileSizeThreshold = 5 * 1024 * 1024
		
	default: // >= 50MB
		opts.EnableStreaming = true
		opts.EnableMemoryMonitor = true
		opts.FileSizeThreshold = 1 * 1024 * 1024
	}
	
	return opts, nil
}