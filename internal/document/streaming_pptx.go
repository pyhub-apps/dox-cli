package document

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// StreamingPowerPointDocument handles large PowerPoint documents efficiently
type StreamingPowerPointDocument struct {
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

// OpenPowerPointDocumentStreaming opens a PowerPoint document for streaming processing
func OpenPowerPointDocumentStreaming(path string, opts *StreamingOptions) (*StreamingPowerPointDocument, error) {
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
	if !strings.HasSuffix(strings.ToLower(path), ".pptx") {
		return nil, fmt.Errorf("not a .pptx file: %s", path)
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
		return nil, fmt.Errorf("invalid pptx format: %w", err)
	}
	
	doc := &StreamingPowerPointDocument{
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

// ProcessSlidesChunked processes all slides text in chunks
func (d *StreamingPowerPointDocument) ProcessSlidesChunked(processor func(slideNum int, chunk string) error) error {
	if d.closed {
		return fmt.Errorf("document is closed")
	}
	
	// Find all slide files
	slideFiles := make([]*zip.File, 0)
	for _, file := range d.zipFile.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && 
		   strings.HasSuffix(file.Name, ".xml") &&
		   !strings.Contains(file.Name, "_rels") {
			slideFiles = append(slideFiles, file)
		}
	}
	
	// Process each slide
	for i, slideFile := range slideFiles {
		if err := d.processSlideChunked(slideFile, i+1, processor); err != nil {
			return fmt.Errorf("error processing slide %d: %w", i+1, err)
		}
	}
	
	return nil
}

// processSlideChunked processes a single slide in chunks
func (d *StreamingPowerPointDocument) processSlideChunked(slideFile *zip.File, slideNum int, processor func(int, string) error) error {
	// Open slide XML
	rc, err := slideFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open slide: %w", err)
	}
	defer rc.Close()
	
	// Use buffered reader
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
			// PowerPoint uses 'a:t' for text elements
			if element.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if element.Name.Local == "t" {
				inText = false
				// Process accumulated text
				if currentText.Len() > 0 {
					if err := processor(slideNum, currentText.String()); err != nil {
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
		
		// Check chunk size
		if currentText.Len() > d.options.ChunkSize {
			if err := processor(slideNum, currentText.String()); err != nil {
				return err
			}
			currentText.Reset()
		}
	}
	
	// Process any remaining text
	if currentText.Len() > 0 {
		if err := processor(slideNum, currentText.String()); err != nil {
			return err
		}
	}
	
	return nil
}

// ReplaceTextInSlidesStreaming replaces text in all slides using streaming
// Returns the number of replacements made
func (d *StreamingPowerPointDocument) ReplaceTextInSlidesStreaming(oldText, newText string) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("document is closed")
	}
	
	// Create temporary file for output
	tmpFile, err := os.CreateTemp("", "pptx-stream-*.tmp")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	
	// Ensure temp file is cleaned up in all cases
	defer CleanupTempFile(tmpPath)
	
	// Create new zip writer for output
	zipWriter := zip.NewWriter(tmpFile)
	
	// Track replacement count
	totalReplacements := 0
	
	// Process each file in the source zip
	for _, file := range d.zipFile.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && 
		   strings.HasSuffix(file.Name, ".xml") &&
		   !strings.Contains(file.Name, "_rels") {
			// Stream and modify slide files
			count, err := d.streamAndModifySlide(file, zipWriter, oldText, newText)
			if err != nil {
				zipWriter.Close()
				tmpFile.Close()
				return 0, fmt.Errorf("failed to process %s: %w", file.Name, err)
			}
			totalReplacements += count
		} else {
			// Copy other files as-is
			if err := d.copyZipFile(file, zipWriter); err != nil {
				zipWriter.Close()
				tmpFile.Close()
				return 0, fmt.Errorf("failed to copy %s: %w", file.Name, err)
			}
		}
	}
	
	// Close the zip writer to finalize the archive
	if err := zipWriter.Close(); err != nil {
		tmpFile.Close()
		return 0, fmt.Errorf("failed to finalize zip: %w", err)
	}
	
	// Close the temp file
	if err := tmpFile.Close(); err != nil {
		return 0, fmt.Errorf("failed to close temp file: %w", err)
	}
	
	if totalReplacements > 0 {
		d.modified = true
		// Close the original file handle
		if err := d.file.Close(); err != nil {
			return totalReplacements, fmt.Errorf("failed to close original file: %w", err)
		}
		
		// Replace the original file with the modified version
		if err := os.Rename(tmpPath, d.path); err != nil {
			// Try to reopen the original file
			d.file, _ = os.Open(d.path)
			return totalReplacements, fmt.Errorf("failed to replace original file: %w", err)
		}
		
		// Reopen the file for potential further operations
		d.file, err = os.Open(d.path)
		if err != nil {
			return totalReplacements, fmt.Errorf("failed to reopen file: %w", err)
		}
		
		// Recreate zip reader
		fileInfo, err := d.file.Stat()
		if err != nil {
			d.file.Close()
			d.file = nil
			return totalReplacements, fmt.Errorf("failed to stat reopened file: %w", err)
		}
		d.zipFile, err = zip.NewReader(d.file, fileInfo.Size())
		if err != nil {
			d.file.Close()
			d.file = nil
			return totalReplacements, fmt.Errorf("failed to recreate zip reader: %w", err)
		}
	}
	
	return totalReplacements, nil
}

// streamAndModifySlide processes and modifies slide XML content in a streaming manner
func (d *StreamingPowerPointDocument) streamAndModifySlide(src *zip.File, dst *zip.Writer, oldText, newText string) (int, error) {
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
		
		// Modify text content (PowerPoint uses 'a:t' elements for text)
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
func (d *StreamingPowerPointDocument) copyZipFile(src *zip.File, dst *zip.Writer) error {
	// Use buffer from pool if available
	var buffer []byte
	if d.options.EnableMemoryPool && d.memPool != nil {
		buffer = d.memPool.Get().([]byte)
		defer d.memPool.Put(buffer)
	} else {
		buffer = make([]byte, d.options.ChunkSize)
	}
	
	err := CopyZipFileWithCompression(src, dst, buffer)
	if err != nil {
		return fmt.Errorf("failed to copy zip file: %w", err)
	}
	return nil
}

// GetMemoryUsage returns current memory usage
func (d *StreamingPowerPointDocument) GetMemoryUsage() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.memUsage
}

// Close closes the document and releases resources
func (d *StreamingPowerPointDocument) Close() error {
	if d.closed {
		return nil
	}
	
	d.closed = true
	
	if d.file != nil {
		if err := d.file.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}
	
	return nil
}

// CountSlides counts the number of slides in the presentation
func (d *StreamingPowerPointDocument) CountSlides() int {
	count := 0
	for _, file := range d.zipFile.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && 
		   strings.HasSuffix(file.Name, ".xml") &&
		   !strings.Contains(file.Name, "_rels") {
			count++
		}
	}
	return count
}

// GetSlideNumbers returns all slide numbers in the presentation
func (d *StreamingPowerPointDocument) GetSlideNumbers() []int {
	slides := make([]int, 0)
	for i := 1; i <= d.CountSlides(); i++ {
		slides = append(slides, i)
	}
	return slides
}