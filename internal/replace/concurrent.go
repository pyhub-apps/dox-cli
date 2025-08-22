package replace

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/pyhub/pyhub-docs/internal/ui"
)

// ConcurrentOptions configures concurrent processing
type ConcurrentOptions struct {
	MaxWorkers   int  // Maximum number of concurrent workers
	ShowProgress bool // Whether to show progress
}

// DefaultConcurrentOptions returns default concurrent options
func DefaultConcurrentOptions() ConcurrentOptions {
	return ConcurrentOptions{
		MaxWorkers:   runtime.NumCPU(),
		ShowProgress: false,
	}
}

// ReplaceInDirectoryConcurrent processes documents concurrently
func ReplaceInDirectoryConcurrent(dirPath string, rules []Rule, recursive bool, excludePattern string, opts ConcurrentOptions) ([]ReplaceResult, error) {
	// Collect all files to process
	var files []string
	err := WalkDocumentFilesWithExclude(dirPath, recursive, excludePattern, func(path string) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []ReplaceResult{}, nil
	}

	// Validate all rules before processing
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return nil, err
		}
		_ = i // avoid unused variable warning
	}

	// Create worker pool
	if opts.MaxWorkers <= 0 {
		opts.MaxWorkers = 1
	}
	
	// Create progress bar if needed
	var progressBar *ui.ProgressBar
	if opts.ShowProgress {
		progressBar = ui.NewProgressBar(len(files), "Processing documents")
	}
	
	// Use buffered channel as semaphore for limiting workers
	sem := make(chan struct{}, opts.MaxWorkers)
	
	// Results channel and wait group
	results := make([]ReplaceResult, len(files))
	var wg sync.WaitGroup
	var processed int32

	// Process files concurrently
	for i, file := range files {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		
		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore
			
			result := ReplaceResult{
				FilePath: path,
			}
			
			// Process the document
			count, err := ReplaceInDocumentWithCount(path, rules)
			if err != nil {
				result.Success = false
				result.Error = err
			} else {
				result.Success = true
				result.Replacements = count
			}
			
			results[idx] = result
			
			// Update progress
			if opts.ShowProgress && progressBar != nil {
				progressBar.Increment()
				current := atomic.AddInt32(&processed, 1)
				_ = current // Progress bar handles display
			}
		}(i, file)
	}
	
	// Wait for all workers to complete
	wg.Wait()
	
	// Finish progress bar
	if progressBar != nil {
		progressBar.Finish()
	}
	
	return results, nil
}

// printProgress prints a simple progress indicator
func printProgress(current, total int) {
	percent := float64(current) * 100.0 / float64(total)
	if current == total {
		// Clear line and print completion
		fmt.Print("\r\033[K") // Clear line
		fmt.Println("Processing complete!")
	} else {
		// Update progress on same line
		fmt.Print("\r\033[K") // Clear line
		fmt.Printf("Processing: %.1f%% (%d/%d)", percent, current, total)
	}
}