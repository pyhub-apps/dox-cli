package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// Color definitions for consistent UI
var (
	// Status colors
	Success = color.New(color.FgGreen, color.Bold)
	Error   = color.New(color.FgRed, color.Bold)
	Warning = color.New(color.FgYellow, color.Bold)
	Info    = color.New(color.FgCyan)
	
	// Element colors
	Header  = color.New(color.FgWhite, color.Bold, color.Underline)
	Accent  = color.New(color.FgMagenta, color.Bold)
	Muted   = color.New(color.FgHiBlack)
	
	// File type colors
	DocxColor = color.New(color.FgBlue)
	PptxColor = color.New(color.FgMagenta)
	MarkdownColor = color.New(color.FgGreen)
)

// Icons for different statuses (supports both Unicode and ASCII)
var (
	iconSuccess = "✓"
	iconError   = "✗"
	iconWarning = "⚠"
	iconInfo    = "ℹ"
	iconProcess = "▶"
	iconDot     = "•"
)

func init() {
	// Fallback to ASCII icons if terminal doesn't support Unicode
	if os.Getenv("TERM") == "dumb" || os.Getenv("NO_UNICODE") != "" {
		iconSuccess = "[OK]"
		iconError   = "[ERROR]"
		iconWarning = "[WARN]"
		iconInfo    = "[INFO]"
		iconProcess = ">"
		iconDot     = "*"
	}
}

// PrintSuccess prints a success message with green color
func PrintSuccess(format string, args ...interface{}) {
	Success.Printf("%s ", iconSuccess)
	fmt.Printf(format+"\n", args...)
}

// PrintError prints an error message with red color
func PrintError(format string, args ...interface{}) {
	Error.Printf("%s ", iconError)
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintWarning prints a warning message with yellow color
func PrintWarning(format string, args ...interface{}) {
	Warning.Printf("%s ", iconWarning)
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintInfo prints an info message with cyan color
func PrintInfo(format string, args ...interface{}) {
	Info.Printf("%s ", iconInfo)
	fmt.Printf(format+"\n", args...)
}

// PrintHeader prints a header with underline
func PrintHeader(text string) {
	fmt.Println()
	Header.Println(text)
	fmt.Println()
}

// PrintStep prints a step in a process
func PrintStep(step int, total int, description string) {
	Accent.Printf("[%d/%d] ", step, total)
	fmt.Printf("%s %s\n", iconProcess, description)
}

// PrintFileOperation prints a file operation with appropriate color
func PrintFileOperation(operation, filePath, fileType string) {
	fmt.Printf("%s ", iconDot)
	
	switch fileType {
	case ".docx":
		DocxColor.Printf("[DOCX] ")
	case ".pptx":
		PptxColor.Printf("[PPTX] ")
	case ".md":
		MarkdownColor.Printf("[MD] ")
	default:
		Muted.Printf("[FILE] ")
	}
	
	fmt.Printf("%s: %s\n", operation, filePath)
}

// ProgressBar represents a progress bar for operations
type ProgressBar struct {
	bar   *progressbar.ProgressBar
	total int
	mu    sync.Mutex
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, description string) *ProgressBar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[yellow]▶[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)
	
	return &ProgressBar{
		bar:   bar,
		total: total,
	}
}

// NewSpinner creates a spinner for indefinite operations
func NewSpinner(description string) *ProgressBar {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[cyan]•[reset]",
			SaucerHead:    "[cyan]•[reset]",
			SaucerPadding: " ",
			BarStart:      "",
			BarEnd:        "",
		}),
	)
	
	return &ProgressBar{
		bar:   bar,
		total: -1,
	}
}

// Increment increments the progress bar by one
func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Add(1)
}

// IncrementBy increments the progress bar by n
func (p *ProgressBar) IncrementBy(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Add(n)
}

// SetDescription updates the progress bar description
func (p *ProgressBar) SetDescription(description string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Describe(description)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Finish()
}

// Clear clears the progress bar from the terminal
func (p *ProgressBar) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Clear()
}

// MultiProgressManager manages multiple progress bars
type MultiProgressManager struct {
	bars   []*ProgressBar
	writer io.Writer
	mu     sync.Mutex
}

// NewMultiProgressManager creates a new multi-progress manager
func NewMultiProgressManager() *MultiProgressManager {
	return &MultiProgressManager{
		bars:   make([]*ProgressBar, 0),
		writer: os.Stdout,
	}
}

// AddBar adds a progress bar to the manager
func (m *MultiProgressManager) AddBar(total int, description string) *ProgressBar {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	bar := NewProgressBar(total, description)
	m.bars = append(m.bars, bar)
	return bar
}

// AddSpinner adds a spinner to the manager
func (m *MultiProgressManager) AddSpinner(description string) *ProgressBar {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	spinner := NewSpinner(description)
	m.bars = append(m.bars, spinner)
	return spinner
}

// Wait waits for all progress bars to complete
func (m *MultiProgressManager) Wait() {
	// Progress bars auto-complete when reaching their total
	time.Sleep(100 * time.Millisecond) // Small delay for visual clarity
}

// Confirmation asks for user confirmation with colored prompt
func Confirmation(prompt string) bool {
	Warning.Printf("⚠ %s [y/N]: ", prompt)
	
	var response string
	fmt.Scanln(&response)
	
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}

// PrintSummary prints a summary with statistics
func PrintSummary(title string, stats map[string]interface{}) {
	PrintHeader(title)
	
	for key, value := range stats {
		switch v := value.(type) {
		case int:
			if v > 0 {
				Success.Printf("  %s: ", key)
				fmt.Printf("%d\n", v)
			} else {
				Muted.Printf("  %s: ", key)
				fmt.Printf("%d\n", v)
			}
		case string:
			Info.Printf("  %s: ", key)
			fmt.Printf("%s\n", v)
		case bool:
			if v {
				Success.Printf("  %s: ", key)
				fmt.Println("Yes")
			} else {
				Muted.Printf("  %s: ", key)
				fmt.Println("No")
			}
		default:
			fmt.Printf("  %s: %v\n", key, v)
		}
	}
}

// FormatFileSize formats bytes into human-readable size
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats a duration into human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// EnableColor forces color output even when not in a TTY
func EnableColor() {
	color.NoColor = false
}

// DisableColor disables all color output
func DisableColor() {
	color.NoColor = true
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return !color.NoColor
}