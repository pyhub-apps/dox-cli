package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// WordDocument represents an open Word document
type WordDocument struct {
	path     string
	zipFile  *zip.Reader
	content  *documentContent
	modified bool
	closed   bool
}

// documentContent holds the parsed document.xml content
type documentContent struct {
	XMLName xml.Name `xml:"document"`
	Body    body     `xml:"body"`
	rawXML  []byte   // Store raw XML for preservation
}

type body struct {
	Paragraphs []paragraph `xml:"p"`
}

type paragraph struct {
	Runs []run `xml:"r"`
}

type run struct {
	Text []text `xml:"t"`
}

type text struct {
	Space string `xml:"space,attr,omitempty"`
	Value string `xml:",chardata"`
}

// OpenWordDocument opens a Word document for reading and editing
func OpenWordDocument(path string) (*WordDocument, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".docx") {
		return nil, fmt.Errorf("not a .docx file: %s", path)
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Open as zip
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid docx format: %w", err)
	}
	
	// Find and parse document.xml
	var docXML *zip.File
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			docXML = file
			break
		}
	}
	
	if docXML == nil {
		return nil, fmt.Errorf("invalid docx format: missing document.xml")
	}
	
	// Read document.xml
	rc, err := docXML.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()
	
	xmlData, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read document.xml: %w", err)
	}
	
	// Store raw XML for later use
	doc := &WordDocument{
		path:    path,
		zipFile: reader,
		content: &documentContent{
			rawXML: xmlData,
		},
		modified: false,
		closed:   false,
	}
	
	return doc, nil
}

// GetText extracts all text content from the document
func (w *WordDocument) GetText() []string {
	if w.closed {
		return nil
	}
	
	// Parse the raw XML to extract text
	// For simplicity, we'll use regex to extract text from <w:t> tags
	var paragraphs []string
	
	// Split by paragraph tags
	paraPattern := regexp.MustCompile(`<w:p[^>]*>.*?</w:p>`)
	textPattern := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	
	paras := paraPattern.FindAllString(string(w.content.rawXML), -1)
	
	for _, para := range paras {
		matches := textPattern.FindAllStringSubmatch(para, -1)
		var paraText strings.Builder
		for _, match := range matches {
			if len(match) > 1 {
				// Unescape HTML entities to get the original text
				unescaped := html.UnescapeString(match[1])
				paraText.WriteString(unescaped)
			}
		}
		if text := paraText.String(); text != "" {
			paragraphs = append(paragraphs, text)
		}
	}
	
	return paragraphs
}

// escapeXMLString escapes special XML characters to prevent XML injection
func escapeXMLString(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// ReplaceText replaces all occurrences of old text with new text
func (w *WordDocument) ReplaceText(old, new string) error {
	if w.closed {
		return errors.New("document is closed")
	}
	
	if old == "" {
		return errors.New("old text cannot be empty")
	}
	
	// Escape the new text to prevent XML injection
	newEscaped := escapeXMLString(new)
	
	// Replace in raw XML
	// We need to be careful to only replace text content, not XML tags
	xmlStr := string(w.content.rawXML)
	
	// Use a more sophisticated approach to replace only within text nodes
	textPattern := regexp.MustCompile(`(<w:t[^>]*>)([^<]*)(</w:t>)`)
	
	replaced := false
	xmlStr = textPattern.ReplaceAllStringFunc(xmlStr, func(match string) string {
		submatches := textPattern.FindStringSubmatch(match)
		if len(submatches) == 4 {
			textContent := submatches[2]
			if strings.Contains(textContent, old) {
				replaced = true
				// Note: old text is not escaped as we're searching for it as-is in the document
				newContent := strings.ReplaceAll(textContent, old, newEscaped)
				return submatches[1] + newContent + submatches[3]
			}
		}
		return match
	})
	
	if replaced {
		w.content.rawXML = []byte(xmlStr)
		w.modified = true
	}
	
	return nil
}

// SaveAs saves the document to a new file
func (w *WordDocument) SaveAs(path string) error {
	if w.closed {
		return errors.New("document is closed")
	}
	
	if path == "" {
		return errors.New("path cannot be empty")
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".docx") {
		return fmt.Errorf("output file must have .docx extension")
	}
	
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create new zip file
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	
	// Copy all files from original, replacing document.xml if modified
	for _, file := range w.zipFile.File {
		var data []byte
		
		if file.Name == "word/document.xml" && w.modified {
			// Use modified content
			data = w.content.rawXML
		} else {
			// Copy original file
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in zip: %w", err)
			}
			data, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to read file in zip: %w", err)
			}
		}
		
		// Write to new zip
		writer, err := zipWriter.Create(file.Name)
		if err != nil {
			return fmt.Errorf("failed to create file in zip: %w", err)
		}
		
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write file in zip: %w", err)
		}
	}
	
	// Close zip writer
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// Save saves changes to the original file
func (w *WordDocument) Save() error {
	if w.closed {
		return errors.New("document is closed")
	}
	
	return w.SaveAs(w.path)
}

// Close closes the document
func (w *WordDocument) Close() error {
	w.closed = true
	return nil
}