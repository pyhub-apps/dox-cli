package export

import (
	"fmt"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/pdf"
)

// Format represents the export format
type Format string

const (
	FormatHTML     Format = "html"
	FormatMarkdown Format = "markdown"
)

// Converter handles conversion from PDF extraction result to various formats
type Converter struct {
	result *pdf.ExtractResult
}

// NewConverter creates a new converter
func NewConverter(result *pdf.ExtractResult) *Converter {
	return &Converter{
		result: result,
	}
}

// Convert converts the extraction result to the specified format
func (c *Converter) Convert(format Format) (string, error) {
	switch format {
	case FormatHTML:
		return c.ToHTML()
	case FormatMarkdown:
		return c.ToMarkdown()
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// ToHTML converts the extraction result to HTML
func (c *Converter) ToHTML() (string, error) {
	var builder strings.Builder

	// HTML header
	builder.WriteString("<!DOCTYPE html>\n")
	builder.WriteString("<html lang=\"ko\">\n")
	builder.WriteString("<head>\n")
	builder.WriteString("  <meta charset=\"UTF-8\">\n")
	builder.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	
	// Title from metadata or filename
	title := c.result.Metadata.Title
	if title == "" {
		title = c.result.Filename
	}
	builder.WriteString(fmt.Sprintf("  <title>%s</title>\n", escapeHTML(title)))
	
	// CSS styles for better table rendering
	builder.WriteString("  <style>\n")
	builder.WriteString("    body { font-family: 'Malgun Gothic', sans-serif; line-height: 1.6; margin: 40px; }\n")
	builder.WriteString("    table { border-collapse: collapse; margin: 20px 0; width: auto; }\n")
	builder.WriteString("    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }\n")
	builder.WriteString("    th { background-color: #f2f2f2; font-weight: bold; }\n")
	builder.WriteString("    .page-break { page-break-after: always; margin: 40px 0; border-top: 2px solid #ccc; }\n")
	builder.WriteString("    .metadata { background: #f9f9f9; padding: 10px; margin-bottom: 20px; border-radius: 5px; }\n")
	builder.WriteString("  </style>\n")
	builder.WriteString("</head>\n")
	builder.WriteString("<body>\n")

	// Add metadata if available
	if c.result.Metadata.Title != "" || c.result.Metadata.Author != "" {
		builder.WriteString("  <div class=\"metadata\">\n")
		if c.result.Metadata.Title != "" {
			builder.WriteString(fmt.Sprintf("    <h1>%s</h1>\n", escapeHTML(c.result.Metadata.Title)))
		}
		if c.result.Metadata.Author != "" {
			builder.WriteString(fmt.Sprintf("    <p><strong>Author:</strong> %s</p>\n", escapeHTML(c.result.Metadata.Author)))
		}
		if c.result.Metadata.Subject != "" {
			builder.WriteString(fmt.Sprintf("    <p><strong>Subject:</strong> %s</p>\n", escapeHTML(c.result.Metadata.Subject)))
		}
		builder.WriteString("  </div>\n")
	}

	// Process each page
	for i, page := range c.result.Pages {
		if i > 0 {
			builder.WriteString("  <div class=\"page-break\"></div>\n")
		}

		builder.WriteString(fmt.Sprintf("  <!-- Page %d -->\n", page.Number))
		
		// Process structured elements if available
		if len(page.Elements) > 0 {
			for _, elem := range page.Elements {
				switch elem.Type {
				case "heading":
					level := elem.Level
					if level == 0 {
						level = 3
					}
					builder.WriteString(fmt.Sprintf("  <h%d>%s</h%d>\n", level, escapeHTML(elem.Content), level))
				case "list_item":
					builder.WriteString(fmt.Sprintf("  <li>%s</li>\n", escapeHTML(elem.Content)))
				case "table_row":
					// Skip, will be handled in tables section
					continue
				default:
					builder.WriteString(fmt.Sprintf("  <p>%s</p>\n", escapeHTML(elem.Content)))
				}
			}
		} else if page.Text != "" {
			// Fallback to simple text processing
			lines := strings.Split(page.Text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Simple heading detection (lines that are short and might be titles)
				if len(line) < 50 && !strings.HasSuffix(line, ".") && !strings.HasSuffix(line, ",") {
					builder.WriteString(fmt.Sprintf("  <h3>%s</h3>\n", escapeHTML(line)))
				} else {
					builder.WriteString(fmt.Sprintf("  <p>%s</p>\n", escapeHTML(line)))
				}
			}
		}

		// Process tables
		for _, table := range page.Tables {
			builder.WriteString("  <table>\n")
			for rowIdx, row := range table.Data {
				builder.WriteString("    <tr>\n")
				for _, cell := range row {
					// Use th for first row if it looks like headers
					if rowIdx == 0 && looksLikeHeader(row) {
						builder.WriteString(fmt.Sprintf("      <th>%s</th>\n", escapeHTML(cell)))
					} else {
						builder.WriteString(fmt.Sprintf("      <td>%s</td>\n", escapeHTML(cell)))
					}
				}
				builder.WriteString("    </tr>\n")
			}
			builder.WriteString("  </table>\n")
		}
	}

	builder.WriteString("</body>\n")
	builder.WriteString("</html>\n")

	return builder.String(), nil
}

// ToMarkdown converts the extraction result to Markdown
func (c *Converter) ToMarkdown() (string, error) {
	var builder strings.Builder

	// Add metadata as frontmatter if available
	if c.result.Metadata.Title != "" || c.result.Metadata.Author != "" {
		builder.WriteString("---\n")
		if c.result.Metadata.Title != "" {
			builder.WriteString(fmt.Sprintf("title: %s\n", c.result.Metadata.Title))
		}
		if c.result.Metadata.Author != "" {
			builder.WriteString(fmt.Sprintf("author: %s\n", c.result.Metadata.Author))
		}
		if c.result.Metadata.Subject != "" {
			builder.WriteString(fmt.Sprintf("subject: %s\n", c.result.Metadata.Subject))
		}
		builder.WriteString("---\n\n")
	}

	// Process each page
	for i, page := range c.result.Pages {
		if i > 0 {
			builder.WriteString("\n---\n\n")
		}

		// Process structured elements if available
		if len(page.Elements) > 0 {
			var inList bool
			for _, elem := range page.Elements {
				switch elem.Type {
				case "heading":
					level := elem.Level
					if level == 0 {
						level = 2
					}
					prefix := strings.Repeat("#", level)
					builder.WriteString(fmt.Sprintf("%s %s\n\n", prefix, elem.Content))
					inList = false
				case "list_item":
					if !inList {
						builder.WriteString("\n")
						inList = true
					}
					marker := elem.Marker
					if marker == "" {
						marker = "-"
					}
					builder.WriteString(fmt.Sprintf("%s %s\n", marker, strings.TrimPrefix(elem.Content, marker)))
				case "table_row":
					// Skip, will be handled in tables section
					continue
				default:
					if inList {
						builder.WriteString("\n")
						inList = false
					}
					builder.WriteString(fmt.Sprintf("%s\n\n", elem.Content))
				}
			}
			if inList {
				builder.WriteString("\n")
			}
		} else if page.Text != "" {
			// Fallback to simple text processing
			lines := strings.Split(page.Text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					builder.WriteString("\n")
					continue
				}

				// Simple heading detection
				if len(line) < 50 && !strings.HasSuffix(line, ".") && !strings.HasSuffix(line, ",") {
					builder.WriteString(fmt.Sprintf("## %s\n\n", line))
				} else {
					builder.WriteString(fmt.Sprintf("%s\n\n", line))
				}
			}
		}

		// Process tables
		for _, table := range page.Tables {
			if len(table.Data) == 0 {
				continue
			}

			// Write table in Markdown format
			for rowIdx, row := range table.Data {
				builder.WriteString("|")
				for _, cell := range row {
					builder.WriteString(fmt.Sprintf(" %s |", strings.ReplaceAll(cell, "|", "\\|")))
				}
				builder.WriteString("\n")

				// Add separator after header row
				if rowIdx == 0 {
					builder.WriteString("|")
					for range row {
						builder.WriteString(" --- |")
					}
					builder.WriteString("\n")
				}
			}
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// looksLikeHeader checks if a row looks like table headers
func looksLikeHeader(row []string) bool {
	for _, cell := range row {
		// Check if cells contain typical header keywords
		lower := strings.ToLower(cell)
		if strings.Contains(lower, "번호") || strings.Contains(lower, "이름") ||
			strings.Contains(lower, "날짜") || strings.Contains(lower, "구분") ||
			strings.Contains(lower, "no.") || strings.Contains(lower, "name") ||
			strings.Contains(lower, "date") || strings.Contains(lower, "type") {
			return true
		}
	}
	return false
}