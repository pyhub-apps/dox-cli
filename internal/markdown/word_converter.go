package markdown

import (
	"fmt"
)

// WordConverter converts markdown to Word document
type WordConverter struct {
	builder *WordBuilder
}

// NewWordConverter creates a new Word converter
func NewWordConverter() *WordConverter {
	return &WordConverter{
		builder: NewWordBuilder(),
	}
}

// Convert converts markdown document to Word
func (w *WordConverter) Convert(doc *Document) error {
	// Convert blocks to Word content
	for _, block := range doc.Blocks {
		if err := w.convertBlock(block); err != nil {
			return fmt.Errorf("failed to convert block: %w", err)
		}
	}
	
	return nil
}

// SaveAs saves the Word document to the specified path
func (w *WordConverter) SaveAs(path string) error {
	if w.builder == nil {
		return fmt.Errorf("no document to save")
	}
	
	return w.builder.Build(path)
}

// convertBlock converts a markdown block to Word content
func (w *WordConverter) convertBlock(block Block) error {
	switch block.Type {
	case BlockHeading:
		w.builder.AddHeading(block.Level, block.Content)
		
	case BlockParagraph:
		w.builder.AddParagraph(block.Content)
		
	case BlockList:
		w.builder.AddList(block.Items, false)
		
	case BlockOrderedList:
		w.builder.AddList(block.Items, true)
		
	case BlockCodeBlock:
		w.builder.AddCodeBlock(block.Content)
		
	case BlockQuote:
		w.builder.AddQuote(block.Content)
	}
	
	return nil
}