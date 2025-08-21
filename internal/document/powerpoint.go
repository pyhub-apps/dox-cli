package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PowerPointDocument represents a PowerPoint presentation
type PowerPointDocument struct {
	path     string
	zipFile  *zip.ReadCloser
	slides   map[string]*slideContent
	modified bool
	tempBuf  *bytes.Buffer
}

// slideContent holds the content of a single slide
type slideContent struct {
	path    string
	content []byte
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
		tempBuf: new(bytes.Buffer),
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
				content: content,
				xmlDoc:  string(content),
			}
		}
	}

	return nil
}

// GetText extracts all text from the PowerPoint presentation
func (d *PowerPointDocument) GetText() (string, error) {
	var allText strings.Builder

	// Process each slide in order
	for i := 1; i <= len(d.slides); i++ {
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", i)
		slide, exists := d.slides[slidePath]
		if !exists {
			continue
		}

		// Extract text from the slide
		text := extractTextFromSlide(slide.xmlDoc)
		if text != "" {
			allText.WriteString(fmt.Sprintf("Slide %d:\n%s\n\n", i, text))
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
			// Unescape HTML entities
			text := strings.ReplaceAll(match[1], "&lt;", "<")
			text = strings.ReplaceAll(text, "&gt;", ">")
			text = strings.ReplaceAll(text, "&amp;", "&")
			text = strings.ReplaceAll(text, "&quot;", "\"")
			text = strings.ReplaceAll(text, "&apos;", "'")
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

	replacementCount := 0

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
			replacementCount++
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
			
			writer, err := w.Create(file.Name)
			if err != nil {
				reader.Close()
				return fmt.Errorf("failed to create %s in zip: %w", file.Name, err)
			}
			
			if _, err := io.Copy(writer, reader); err != nil {
				reader.Close()
				return fmt.Errorf("failed to copy %s: %w", file.Name, err)
			}
			reader.Close()
		}
	}

	// Close the zip writer
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Write to file
	if err := os.WriteFile(d.path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
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