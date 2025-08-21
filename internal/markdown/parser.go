package markdown

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Document represents a parsed markdown document
type Document struct {
	// Raw markdown content
	Raw []byte
	
	// Parsed AST
	Node ast.Node
	
	// Sections for PowerPoint conversion (split by H1)
	Sections []Section
	
	// All content for Word conversion
	Blocks []Block
}

// Section represents a slide in PowerPoint or a major section in Word
type Section struct {
	Title  string
	Level  int
	Blocks []Block
}

// Block represents a content block (paragraph, list, code, etc.)
type Block struct {
	Type    BlockType
	Content string
	Items   []string // For lists
	Level   int      // For headings
}

// BlockType represents the type of content block
type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockList
	BlockOrderedList
	BlockCode
	BlockCodeBlock
	BlockQuote
	BlockTable
)

// ParseFile parses a markdown file
func ParseFile(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	return Parse(content)
}

// Parse parses markdown content
func Parse(content []byte) (*Document, error) {
	doc := &Document{
		Raw: content,
	}
	
	// Parse markdown to AST
	parser := goldmark.DefaultParser()
	reader := text.NewReader(content)
	doc.Node = parser.Parse(reader)
	
	// Extract sections and blocks
	if err := doc.extractContent(); err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}
	
	return doc, nil
}

// extractContent walks the AST and extracts sections and blocks
func (d *Document) extractContent() error {
	var currentSection *Section
	
	// Walk the AST
	err := ast.Walk(d.Node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		
		switch node := n.(type) {
		case *ast.Heading:
			// Extract heading text
			title := d.extractText(node)
			
			if node.Level == 1 {
				// Start a new section for H1
				if currentSection != nil {
					d.Sections = append(d.Sections, *currentSection)
				}
				currentSection = &Section{
					Title: title,
					Level: node.Level,
				}
			}
			
			// Add as a block
			block := Block{
				Type:    BlockHeading,
				Content: title,
				Level:   node.Level,
			}
			
			d.Blocks = append(d.Blocks, block)
			if currentSection != nil && node.Level > 1 {
				currentSection.Blocks = append(currentSection.Blocks, block)
			}
			
		case *ast.Paragraph:
			text := d.extractText(node)
			if text != "" {
				block := Block{
					Type:    BlockParagraph,
					Content: text,
				}
				d.Blocks = append(d.Blocks, block)
				if currentSection != nil {
					currentSection.Blocks = append(currentSection.Blocks, block)
				}
			}
			
		case *ast.List:
			items := d.extractListItems(node)
			if len(items) > 0 {
				blockType := BlockList
				if node.IsOrdered() {
					blockType = BlockOrderedList
				}
				
				block := Block{
					Type:  blockType,
					Items: items,
				}
				d.Blocks = append(d.Blocks, block)
				if currentSection != nil {
					currentSection.Blocks = append(currentSection.Blocks, block)
				}
			}
			// Skip children as we've already processed them
			return ast.WalkSkipChildren, nil
			
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			code := d.extractCodeBlock(node)
			block := Block{
				Type:    BlockCodeBlock,
				Content: code,
			}
			d.Blocks = append(d.Blocks, block)
			if currentSection != nil {
				currentSection.Blocks = append(currentSection.Blocks, block)
			}
			
		case *ast.Blockquote:
			text := d.extractText(node)
			block := Block{
				Type:    BlockQuote,
				Content: text,
			}
			d.Blocks = append(d.Blocks, block)
			if currentSection != nil {
				currentSection.Blocks = append(currentSection.Blocks, block)
			}
			// Skip children as we've already processed them
			return ast.WalkSkipChildren, nil
		}
		
		return ast.WalkContinue, nil
	})
	
	// Add the last section if exists
	if currentSection != nil {
		d.Sections = append(d.Sections, *currentSection)
	}
	
	// If no H1 sections were found, create a single section with all content
	if len(d.Sections) == 0 && len(d.Blocks) > 0 {
		d.Sections = append(d.Sections, Section{
			Title:  "Document",
			Level:  1,
			Blocks: d.Blocks,
		})
	}
	
	return err
}

// extractText extracts plain text from a node
func (d *Document) extractText(node ast.Node) string {
	var buf bytes.Buffer
	
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		
		switch node := n.(type) {
		case *ast.Text:
			buf.Write(node.Text(d.Raw))
		case *ast.CodeSpan:
			buf.Write(node.Text(d.Raw))
		case *ast.String:
			buf.Write(node.Value)
		}
		
		return ast.WalkContinue, nil
	})
	
	return strings.TrimSpace(buf.String())
}

// extractListItems extracts items from a list node
func (d *Document) extractListItems(list ast.Node) []string {
	var items []string
	
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.ListItem); ok {
			text := d.extractText(child)
			if text != "" {
				items = append(items, text)
			}
		}
	}
	
	return items
}

// extractCodeBlock extracts code from a code block node
func (d *Document) extractCodeBlock(node ast.Node) string {
	var buf bytes.Buffer
	
	switch n := node.(type) {
	case *ast.FencedCodeBlock:
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(d.Raw))
		}
	case *ast.CodeBlock:
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(d.Raw))
		}
	}
	
	return strings.TrimSpace(buf.String())
}

// Converter interface for different output formats
type Converter interface {
	Convert(doc *Document) error
	SaveAs(path string) error
}

// ConvertFile converts a markdown file to the specified format
func ConvertFile(mdPath string, converter Converter, outputPath string) error {
	doc, err := ParseFile(mdPath)
	if err != nil {
		return fmt.Errorf("failed to parse markdown: %w", err)
	}
	
	if err := converter.Convert(doc); err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}
	
	if err := converter.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	
	return nil
}

// ConvertReader converts markdown from a reader
func ConvertReader(reader io.Reader, converter Converter, outputPath string) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}
	
	doc, err := Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse markdown: %w", err)
	}
	
	if err := converter.Convert(doc); err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}
	
	if err := converter.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	
	return nil
}