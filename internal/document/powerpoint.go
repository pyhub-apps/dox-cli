package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PowerPointDocument represents a PowerPoint presentation
type PowerPointDocument struct {
	path     string
	zipFile  *zip.ReadCloser
	slides   map[string]*slideContent
	modified bool
}

// slideContent holds the content of a single slide
type slideContent struct {
	path    string
	xmlDoc  string
}

// OpenPowerPointDocument opens a PowerPoint file for reading and modification
func OpenPowerPointDocument(path string) (*PowerPointDocument, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Open the file as a zip archive
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PowerPoint file: %w", err)
	}

	doc := &PowerPointDocument{
		path:    path,
		zipFile: reader,
		slides:  make(map[string]*slideContent),
	}

	// Load all slides
	if err := doc.loadSlides(); err != nil {
		reader.Close()
		return nil, fmt.Errorf("failed to load slides: %w", err)
	}

	return doc, nil
}

// loadSlides loads all slide content from the PowerPoint file
func (d *PowerPointDocument) loadSlides() error {
	for _, file := range d.zipFile.File {
		// Check if this is a slide file
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			// Skip slide relationships files
			if strings.Contains(file.Name, "_rels") {
				continue
			}

			// Read slide content
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open slide %s: %w", file.Name, err)
			}
			
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to read slide %s: %w", file.Name, err)
			}

			d.slides[file.Name] = &slideContent{
				path:    file.Name,
				xmlDoc:  string(content),
			}
		}
	}

	return nil
}

// GetText extracts all text from the PowerPoint presentation
func (d *PowerPointDocument) GetText() (string, error) {
	var allText strings.Builder

	// Process slides by finding all that match the pattern
	var slideNums []int
	for path := range d.slides {
		if strings.HasPrefix(path, "ppt/slides/slide") && strings.HasSuffix(path, ".xml") {
			// Extract slide number from path like "ppt/slides/slide1.xml"
			baseName := strings.TrimPrefix(path, "ppt/slides/slide")
			baseName = strings.TrimSuffix(baseName, ".xml")
			if num, err := strconv.Atoi(baseName); err == nil {
				slideNums = append(slideNums, num)
			}
		}
	}
	
	// Sort slide numbers to process in order
	sort.Ints(slideNums)
	
	// Process each slide in order
	for _, num := range slideNums {
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", num)
		slide := d.slides[slidePath]
		
		// Extract text from the slide
		text := extractTextFromSlide(slide.xmlDoc)
		if text != "" {
			allText.WriteString(fmt.Sprintf("Slide %d:\n%s\n\n", num, text))
		}
	}

	return allText.String(), nil
}

// extractTextFromSlide extracts text from a slide's XML content
func extractTextFromSlide(xmlContent string) string {
	var texts []string
	
	// Find all text elements using regex
	// PowerPoint uses <a:t> tags for text
	re := regexp.MustCompile(`<a:t[^>]*>([^<]+)</a:t>`)
	matches := re.FindAllStringSubmatch(xmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			// Unescape HTML entities using standard library
			text := html.UnescapeString(match[1])
			texts = append(texts, text)
		}
	}
	
	return strings.Join(texts, "\n")
}

// ReplaceText replaces all occurrences of old text with new text in the presentation
func (d *PowerPointDocument) ReplaceText(old, new string) error {
	if old == "" {
		return fmt.Errorf("search text cannot be empty")
	}

	// Process each slide
	for _, slide := range d.slides {
		originalContent := slide.xmlDoc
		
		// Escape the old and new text for XML
		oldEscaped := escapeXMLStringPPT(old)
		newEscaped := escapeXMLStringPPT(new)
		
		// Replace text in <a:t> tags
		// We need to be careful to only replace within text content
		modified := replaceTextInXML(slide.xmlDoc, oldEscaped, newEscaped)
		
		// Also try replacing non-escaped version in case text is already in the document
		modified = replaceTextInXML(modified, old, newEscaped)
		
		if modified != originalContent {
			slide.xmlDoc = modified
			d.modified = true
		}
	}

	return nil
}

// replaceTextInXML replaces text within <a:t> tags in XML content
func replaceTextInXML(xmlContent, old, new string) string {
	// Create a pattern to match text within <a:t> tags
	re := regexp.MustCompile(`(<a:t[^>]*>)([^<]*)(<\/a:t>)`)
	
	result := re.ReplaceAllStringFunc(xmlContent, func(match string) string {
		// Extract the parts
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		
		openTag := parts[1]
		content := parts[2]
		closeTag := parts[3]
		
		// Replace the text
		newContent := strings.ReplaceAll(content, old, new)
		
		return openTag + newContent + closeTag
	})
	
	return result
}

// escapeXMLStringPPT escapes special XML characters in a string for PowerPoint
func escapeXMLStringPPT(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// Save saves the modified PowerPoint document
func (d *PowerPointDocument) Save() error {
	if !d.modified {
		return nil // No changes to save
	}

	// Create a new zip file in memory
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Copy all files from the original, replacing modified slides
	for _, file := range d.zipFile.File {
		// Check if this is a modified slide
		if slide, exists := d.slides[file.Name]; exists && strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			// Write modified slide content
			writer, err := w.Create(file.Name)
			if err != nil {
				return fmt.Errorf("failed to create %s in zip: %w", file.Name, err)
			}
			
			if _, err := writer.Write([]byte(slide.xmlDoc)); err != nil {
				return fmt.Errorf("failed to write %s: %w", file.Name, err)
			}
		} else {
			// Copy original file
			reader, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", file.Name, err)
			}
			defer reader.Close()
			
			writer, err := w.Create(file.Name)
			if err != nil {
				return fmt.Errorf("failed to create %s in zip: %w", file.Name, err)
			}
			
			if _, err := io.Copy(writer, reader); err != nil {
				return fmt.Errorf("failed to copy %s: %w", file.Name, err)
			}
		}
	}

	// Close the zip writer
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Write to a temporary file first for atomic save
	dir := filepath.Dir(d.path)
	tmpFile, err := os.CreateTemp(dir, "ppt_save_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	
	// Ensure temp file is cleaned up
	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()
	
	// Write content to temp file
	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Ensure data is flushed to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	
	// Atomically replace the original file
	if err := os.Rename(tmpPath, d.path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// SaveAs saves the PowerPoint document to a new file
func (d *PowerPointDocument) SaveAs(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Temporarily change the path
	originalPath := d.path
	d.path = path
	
	// Save to the new path
	err := d.Save()
	
	// Restore the original path
	d.path = originalPath
	
	return err
}

// Close closes the PowerPoint document
func (d *PowerPointDocument) Close() error {
	if d.zipFile != nil {
		return d.zipFile.Close()
	}
	return nil
}