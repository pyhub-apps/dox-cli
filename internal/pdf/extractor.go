package pdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExtractResult represents the extraction result from PDF
type ExtractResult struct {
	Success  bool     `json:"success"`
	Filename string   `json:"filename"`
	Pages    []Page   `json:"pages"`
	Metadata Metadata `json:"metadata"`
	Error    string   `json:"error,omitempty"`
}

// Page represents a single page from PDF
type Page struct {
	Number   int       `json:"number"`
	Text     string    `json:"text,omitempty"`    // For backward compatibility
	Elements []Element `json:"elements,omitempty"` // Structured elements with coordinates
	Tables   []Table   `json:"tables,omitempty"`
	Layout   Layout    `json:"layout"`
}

// Element represents a structured text element with coordinates
type Element struct {
	Type    string  `json:"type"`    // heading, text, list_item, table_row
	Content string  `json:"content"`
	BBox    BBox    `json:"bbox"`
	Level   int     `json:"level,omitempty"`  // For headings
	Marker  string  `json:"marker,omitempty"` // For list items
}

// BBox represents a bounding box with coordinates
type BBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Table represents extracted table data
type Table struct {
	Index    int        `json:"index,omitempty"`
	Data     [][]string `json:"data"`
	Rows     int        `json:"rows"`
	Cols     int        `json:"cols"`
	Position Position   `json:"position,omitempty"` // Deprecated, use BBox
	BBox     *BBox      `json:"bbox,omitempty"`      // New coordinate format
}

// Position represents table position on page (deprecated, kept for compatibility)
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Layout represents page layout information
type Layout struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Metadata represents PDF metadata
type Metadata struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	Subject    string `json:"subject"`
	Creator    string `json:"creator"`
	TotalPages int    `json:"total_pages"`
}

// ExtractorOptions contains options for PDF extraction
type ExtractorOptions struct {
	Debug         bool
	Strict        bool
	MinQuality    float64
	IgnoreQuality bool
}

// Extractor handles PDF extraction using Python script
type Extractor struct {
	pythonPath string
	scriptPath string
	options    ExtractorOptions
}

// NewExtractor creates a new PDF extractor
func NewExtractor(options ExtractorOptions) (*Extractor, error) {
	// Find Python executable
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("Python not found: %w", err)
	}

	// Get script path
	scriptPath, err := getScriptPath()
	if err != nil {
		return nil, fmt.Errorf("extraction script not found: %w", err)
	}

	return &Extractor{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
		options:    options,
	}, nil
}

// Extract extracts content from PDF file
func (e *Extractor) Extract(pdfPath string) (*ExtractResult, error) {
	// Verify PDF file exists
	if _, err := os.Stat(pdfPath); err != nil {
		return nil, fmt.Errorf("PDF file not found: %s", pdfPath)
	}

	// Prepare command
	args := []string{e.scriptPath, pdfPath}
	if e.options.Debug {
		args = append(args, "--debug")
	}
	if e.options.Strict {
		args = append(args, "--strict")
	}
	if e.options.IgnoreQuality {
		args = append(args, "--ignore-quality")
	}
	if e.options.MinQuality > 0 && e.options.MinQuality != 0.2 {
		args = append(args, "--min-quality", fmt.Sprintf("%.2f", e.options.MinQuality))
	}

	cmd := exec.Command(e.pythonPath, args...)
	
	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	if err := cmd.Run(); err != nil {
		if e.options.Debug && stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "Python stderr: %s\n", stderr.String())
		}
		// Check if it's a quality error based on exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			switch exitCode {
			case 2:
				return nil, fmt.Errorf("extraction quality warning (exit code %d)", exitCode)
			case 3:
				return nil, fmt.Errorf("extraction quality too low (exit code %d)", exitCode)
			default:
				return nil, fmt.Errorf("extraction failed with exit code %d", exitCode)
			}
		}
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Parse JSON output
	var result ExtractResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	// Check for extraction errors
	if !result.Success {
		return nil, fmt.Errorf("extraction failed: %s", result.Error)
	}

	return &result, nil
}

// CheckDependencies checks if Python and required libraries are installed
func (e *Extractor) CheckDependencies() error {
	// Check Python
	if _, err := findPython(); err != nil {
		return fmt.Errorf("Python not found: %w", err)
	}

	// Check pdfplumber
	cmd := exec.Command(e.pythonPath, "-c", "import pdfplumber")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pdfplumber not installed. Please run: pip install pdfplumber")
	}

	return nil
}

// findPython finds Python executable in the system
func findPython() (string, error) {
	// Try common Python commands
	candidates := []string{"python3", "python"}
	
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			// Verify it's Python 3
			cmd := exec.Command(path, "--version")
			output, err := cmd.Output()
			if err == nil && strings.Contains(string(output), "Python 3") {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("Python 3 not found in PATH")
}

// getScriptPath returns the path to the extraction script
func getScriptPath() (string, error) {
	// Try coordinate-based extractor first
	scriptPath := filepath.Join("scripts", "pdf_extract_coords.py")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}
	
	// Fallback to original extractor
	scriptPath = filepath.Join("scripts", "pdf_extract.py")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}

	// Try path relative to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		scriptPath = filepath.Join(exeDir, "scripts", "pdf_extract.py")
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
	}

	// Check if we're in development (go run)
	if runtime.GOOS != "windows" {
		// Try GOPATH/src path
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			scriptPath = filepath.Join(gopath, "src", "github.com", "pyhub", "pyhub-docs", "scripts", "pdf_extract.py")
			if _, err := os.Stat(scriptPath); err == nil {
				return scriptPath, nil
			}
		}
	}

	return "", fmt.Errorf("pdf_extract.py script not found")
}