package document

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
)

// StreamingOptions configures streaming behavior
type StreamingOptions struct {
	// ChunkSize is the size of each chunk to process (default: 64KB)
	ChunkSize int
	// MaxMemory is the maximum memory to use (default: 100MB)
	MaxMemory int64
	// EnableMemoryPool enables memory pool for better performance
	EnableMemoryPool bool
}

// DefaultStreamingOptions returns default streaming options
func DefaultStreamingOptions() *StreamingOptions {
	return &StreamingOptions{
		ChunkSize:        64 * 1024, // 64KB chunks
		MaxMemory:        100 * 1024 * 1024, // 100MB max memory
		EnableMemoryPool: true,
	}
}

// StreamingWordDocument handles large Word documents efficiently
type StreamingWordDocument struct {
	path     string
	file     *os.File
	zipFile  *zip.Reader
	options  *StreamingOptions
	modified bool
	closed   bool
	
	// Memory management
	memPool  *sync.Pool
	memUsage int64
	mu       sync.RWMutex
}

// OpenWordDocumentStreaming opens a Word document for streaming processing
func OpenWordDocumentStreaming(path string, opts *StreamingOptions) (*StreamingWordDocument, error) {
	// Use default options if not provided
	if opts == nil {
		opts = DefaultStreamingOptions()
	}
	
	// Check if file exists and get its info
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".docx") {
		return nil, fmt.Errorf("not a .docx file: %s", path)
	}
	
	// Open file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	// Create zip reader
	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("invalid docx format: %w", err)
	}
	
	doc := &StreamingWordDocument{
		path:    path,
		file:    file,
		zipFile: zipReader,
		options: opts,
	}
	
	// Initialize memory pool if enabled
	if opts.EnableMemoryPool {
		doc.memPool = &sync.Pool{
			New: func() interface{} {
				return make([]byte, opts.ChunkSize)
			},
		}
	}
	
	return doc, nil
}

// ProcessTextChunked processes document text in chunks to save memory
func (d *StreamingWordDocument) ProcessTextChunked(processor func(chunk string) error) error {
	if d.closed {
		return fmt.Errorf("document is closed")
	}
	
	// Find document.xml
	var docXML *zip.File
	for _, file := range d.zipFile.File {
		if file.Name == "word/document.xml" {
			docXML = file
			break
		}
	}
	
	if docXML == nil {
		return fmt.Errorf("document.xml not found in docx")
	}
	
	// Open document.xml for reading
	rc, err := docXML.Open()
	if err != nil {
		return fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()
	
	// Use buffered reader for efficient reading
	reader := bufio.NewReaderSize(rc, d.options.ChunkSize)
	
	// Process XML in streaming mode
	decoder := xml.NewDecoder(reader)
	var currentText strings.Builder
	inText := false
	
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("XML parsing error: %w", err)
		}
		
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if element.Name.Local == "t" {
				inText = false
				// Process accumulated text
				if currentText.Len() > 0 {
					if err := processor(currentText.String()); err != nil {
						return err
					}
					currentText.Reset()
				}
			}
		case xml.CharData:
			if inText {
				currentText.Write(element)
			}
		}
		
		// Check memory usage periodically
		if currentText.Len() > d.options.ChunkSize {
			if err := processor(currentText.String()); err != nil {
				return err
			}
			currentText.Reset()
		}
	}
	
	// Process any remaining text
	if currentText.Len() > 0 {
		if err := processor(currentText.String()); err != nil {
			return err
		}
	}
	
	return nil
}

// ReplaceTextStreaming replaces text in the document using streaming
func (d *StreamingWordDocument) ReplaceTextStreaming(oldText, newText string) error {
	if d.closed {
		return fmt.Errorf("document is closed")
	}
	
	// Find document.xml
	var docXML *zip.File
	for _, file := range d.zipFile.File {
		if file.Name == "word/document.xml" {
			docXML = file
			break
		}
	}
	
	if docXML == nil {
		return fmt.Errorf("document.xml not found in docx")
	}
	
	// Open document.xml
	rc, err := docXML.Open()
	if err != nil {
		return fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()
	
	// Read in chunks and replace
	var buffer []byte
	if d.options.EnableMemoryPool && d.memPool != nil {
		buffer = d.memPool.Get().([]byte)
		defer d.memPool.Put(buffer)
	} else {
		buffer = make([]byte, d.options.ChunkSize)
	}
	
	var result bytes.Buffer
	reader := bufio.NewReaderSize(rc, d.options.ChunkSize)
	
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			chunk := string(buffer[:n])
			// Replace text in chunk
			chunk = strings.ReplaceAll(chunk, oldText, newText)
			result.WriteString(chunk)
			
			// Update memory usage
			d.mu.Lock()
			d.memUsage = int64(result.Len())
			d.mu.Unlock()
			
			// Check if exceeding memory limit
			if d.memUsage > d.options.MaxMemory {
				return fmt.Errorf("memory limit exceeded: %d > %d", d.memUsage, d.options.MaxMemory)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading document: %w", err)
		}
	}
	
	d.modified = true
	// Note: In real implementation, we'd need to update the zip file with new content
	// This is a simplified version
	
	return nil
}

// GetMemoryUsage returns current memory usage
func (d *StreamingWordDocument) GetMemoryUsage() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.memUsage
}

// Close closes the document and releases resources
func (d *StreamingWordDocument) Close() error {
	if d.closed {
		return nil
	}
	
	d.closed = true
	
	if d.file != nil {
		if err := d.file.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}
	
	// Force garbage collection to free memory
	runtime.GC()
	
	return nil
}

// GetEstimatedMemoryForFile estimates memory usage for a file
func GetEstimatedMemoryForFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	
	// Estimate: compressed docx is usually 10-20% of uncompressed size
	// We estimate uncompressed XML could be 5-10x the file size
	estimatedMemory := info.Size() * 10
	
	return estimatedMemory, nil
}

// AdaptiveStreamingOptions returns options based on file size
func AdaptiveStreamingOptions(fileSize int64) *StreamingOptions {
	opts := DefaultStreamingOptions()
	
	switch {
	case fileSize < 1*1024*1024: // < 1MB - small file
		opts.ChunkSize = 16 * 1024 // 16KB chunks
		opts.MaxMemory = 10 * 1024 * 1024 // 10MB max
		opts.EnableMemoryPool = false // Not needed for small files
		
	case fileSize < 10*1024*1024: // < 10MB - medium file
		opts.ChunkSize = 64 * 1024 // 64KB chunks
		opts.MaxMemory = 50 * 1024 * 1024 // 50MB max
		opts.EnableMemoryPool = true
		
	case fileSize < 100*1024*1024: // < 100MB - large file
		opts.ChunkSize = 256 * 1024 // 256KB chunks
		opts.MaxMemory = 100 * 1024 * 1024 // 100MB max
		opts.EnableMemoryPool = true
		
	default: // >= 100MB - very large file
		opts.ChunkSize = 1024 * 1024 // 1MB chunks
		opts.MaxMemory = 200 * 1024 * 1024 // 200MB max
		opts.EnableMemoryPool = true
	}
	
	return opts
}