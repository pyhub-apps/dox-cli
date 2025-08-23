package document

import (
	"archive/zip"
	"bufio"
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
	
	// Create temporary file for output
	tmpFile, err := os.CreateTemp("", "docx-stream-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Create new zip writer for output
	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()
	
	// Track replacement count
	replacementCount := 0
	
	// Process each file in the source zip
	for _, file := range d.zipFile.File {
		if file.Name == "word/document.xml" {
			// Stream and modify this file
			count, err := d.streamAndModifyXML(file, zipWriter, oldText, newText)
			if err != nil {
				return fmt.Errorf("failed to process document.xml: %w", err)
			}
			replacementCount += count
		} else {
			// Copy other files as-is
			if err := d.copyZipFile(file, zipWriter); err != nil {
				return fmt.Errorf("failed to copy %s: %w", file.Name, err)
			}
		}
	}
	
	// Close the zip writer to finalize the archive
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize zip: %w", err)
	}
	tmpFile.Close()
	
	if replacementCount > 0 {
		d.modified = true
		// Close the original file handle
		if err := d.file.Close(); err != nil {
			return fmt.Errorf("failed to close original file: %w", err)
		}
		
		// Replace the original file with the modified version
		if err := os.Rename(tmpFile.Name(), d.path); err != nil {
			return fmt.Errorf("failed to replace original file: %w", err)
		}
		
		// Reopen the file for potential further operations
		d.file, err = os.Open(d.path)
		if err != nil {
			return fmt.Errorf("failed to reopen file: %w", err)
		}
		
		// Recreate zip reader
		fileInfo, err := d.file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat reopened file: %w", err)
		}
		d.zipFile, err = zip.NewReader(d.file, fileInfo.Size())
		if err != nil {
			return fmt.Errorf("failed to recreate zip reader: %w", err)
		}
	}
	
	return nil
}

// streamAndModifyXML processes and modifies XML content in a streaming manner
func (d *StreamingWordDocument) streamAndModifyXML(src *zip.File, dst *zip.Writer, oldText, newText string) (int, error) {
	reader, err := src.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to open source file: %w", err)
	}
	defer reader.Close()
	
	writer, err := dst.Create(src.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to create destination file: %w", err)
	}
	
	// Use XML decoder/encoder for proper streaming
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(writer)
	
	replacementCount := 0
	var buffer []byte
	if d.options.EnableMemoryPool && d.memPool != nil {
		buffer = d.memPool.Get().([]byte)
		defer d.memPool.Put(buffer)
	}
	
	// Stream XML tokens
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return replacementCount, fmt.Errorf("XML decode error: %w", err)
		}
		
		// Modify text content
		if charData, ok := token.(xml.CharData); ok {
			original := string(charData)
			modified := strings.ReplaceAll(original, oldText, newText)
			if original != modified {
				replacementCount += strings.Count(original, oldText)
				token = xml.CharData(modified)
			}
			
			// Update memory usage tracking (only tracks current chunk size, not cumulative)
			// This represents the memory used for the current processing buffer
			d.mu.Lock()
			// Track the larger of the current chunk or configured chunk size
			currentChunkSize := len(modified)
			if currentChunkSize < d.options.ChunkSize {
				d.memUsage = int64(d.options.ChunkSize)
			} else {
				d.memUsage = int64(currentChunkSize)
			}
			d.mu.Unlock()
		}
		
		// Write token immediately (true streaming)
		if err := encoder.EncodeToken(token); err != nil {
			return replacementCount, fmt.Errorf("XML encode error: %w", err)
		}
	}
	
	// Flush encoder
	if err := encoder.Flush(); err != nil {
		return replacementCount, fmt.Errorf("failed to flush encoder: %w", err)
	}
	
	return replacementCount, nil
}

// copyZipFile copies a file from source zip to destination zip without modification
func (d *StreamingWordDocument) copyZipFile(src *zip.File, dst *zip.Writer) error {
	reader, err := src.Open()
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer reader.Close()
	
	// Create destination file with same metadata
	header := src.FileHeader
	writer, err := dst.CreateHeader(&header)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	
	// Use buffer from pool if available
	var buffer []byte
	if d.options.EnableMemoryPool && d.memPool != nil {
		buffer = d.memPool.Get().([]byte)
		defer d.memPool.Put(buffer)
	} else {
		buffer = make([]byte, d.options.ChunkSize)
	}
	
	// Stream copy with buffer
	_, err = io.CopyBuffer(writer, reader, buffer)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	
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