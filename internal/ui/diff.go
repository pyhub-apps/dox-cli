package ui

import (
	"fmt"
	"strings"
	
	"github.com/fatih/color"
)

// DiffLine represents a line in a diff
type DiffLine struct {
	Type    string // "context", "removed", "added"
	Content string
	LineNum int
}

// DiffFormatter formats text differences in a diff-like format
type DiffFormatter struct {
	contextLines int
	colorEnabled bool
}

// NewDiffFormatter creates a new diff formatter
func NewDiffFormatter(contextLines int) *DiffFormatter {
	return &DiffFormatter{
		contextLines: contextLines,
		colorEnabled: !color.NoColor,
	}
}

// FormatTextDiff formats the difference between old and new text
func (df *DiffFormatter) FormatTextDiff(oldText, newText string, filename string) string {
	var result strings.Builder
	
	// Header
	if df.colorEnabled {
		result.WriteString(fmt.Sprintf("\n%s\n", Warning.Sprintf("=== %s ===", filename)))
	} else {
		result.WriteString(fmt.Sprintf("\n=== %s ===\n", filename))
	}
	
	// Split texts into lines
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")
	
	// Simple diff: show changed lines
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}
	
	for i := 0; i < maxLines; i++ {
		var oldLine, newLine string
		
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		
		if oldLine != newLine {
			if oldLine != "" {
				df.writeDiffLine(&result, "-", oldLine, i+1)
			}
			if newLine != "" {
				df.writeDiffLine(&result, "+", newLine, i+1)
			}
		} else if oldLine != "" {
			// Context line
			df.writeDiffLine(&result, " ", oldLine, i+1)
		}
	}
	
	return result.String()
}

// FormatReplacementDiff shows the replacements that will be made
func (df *DiffFormatter) FormatReplacementDiff(text string, replacements map[string]string) string {
	var result strings.Builder
	
	// Apply replacements and show diff
	modifiedText := text
	for old, new := range replacements {
		modifiedText = strings.ReplaceAll(modifiedText, old, new)
	}
	
	// Show replacements summary
	result.WriteString("\nReplacements to be made:\n")
	for old, new := range replacements {
		count := strings.Count(text, old)
		if count > 0 {
			df.writeReplacementLine(&result, old, new, count)
		}
	}
	
	return result.String()
}

// writeDiffLine writes a single diff line with appropriate formatting
func (df *DiffFormatter) writeDiffLine(sb *strings.Builder, prefix string, content string, lineNum int) {
	if df.colorEnabled {
		switch prefix {
		case "-":
			sb.WriteString(Error.Sprintf("%s %4d: %s\n", prefix, lineNum, content))
		case "+":
			sb.WriteString(Success.Sprintf("%s %4d: %s\n", prefix, lineNum, content))
		default:
			sb.WriteString(fmt.Sprintf("  %4d: %s\n", lineNum, content))
		}
	} else {
		if prefix != " " {
			sb.WriteString(fmt.Sprintf("%s %4d: %s\n", prefix, lineNum, content))
		} else {
			sb.WriteString(fmt.Sprintf("  %4d: %s\n", lineNum, content))
		}
	}
}

// writeReplacementLine writes a replacement summary line
func (df *DiffFormatter) writeReplacementLine(sb *strings.Builder, old, new string, count int) {
	if df.colorEnabled {
		sb.WriteString(fmt.Sprintf("  %s → %s (%d occurrence%s)\n",
			Error.Sprint(old),
			Success.Sprint(new),
			count, pluralS(count)))
	} else {
		sb.WriteString(fmt.Sprintf("  '%s' → '%s' (%d occurrence%s)\n",
			old, new, count, pluralS(count)))
	}
}

// pluralS returns "s" if count != 1
func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// ShowSimpleDiff shows a simple before/after comparison
func ShowSimpleDiff(oldText, newText, filename string) {
	formatter := NewDiffFormatter(3)
	fmt.Print(formatter.FormatTextDiff(oldText, newText, filename))
}

// ShowReplacementPreview shows what replacements will be made
func ShowReplacementPreview(text string, replacements map[string]string, filename string) {
	formatter := NewDiffFormatter(3)
	
	if !color.NoColor {
		fmt.Printf("\n%s\n", Warning.Sprintf("=== %s ===", filename))
	} else {
		fmt.Printf("\n=== %s ===\n", filename)
	}
	
	fmt.Print(formatter.FormatReplacementDiff(text, replacements))
}