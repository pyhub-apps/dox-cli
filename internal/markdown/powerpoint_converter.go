package markdown

import (
	"fmt"
)

// PowerPointConverter converts markdown to PowerPoint presentation
// Conversion rules:
// - H1 headers create new slides
// - H2 headers become slide titles (if first in section) or bold content
// - H3-H6 headers become bold content within slides
// - Lists, paragraphs, code blocks, and quotes are preserved as slide content
type PowerPointConverter struct {
	builder *PowerPointBuilder
}

// NewPowerPointConverter creates a new PowerPoint converter
func NewPowerPointConverter() *PowerPointConverter {
	return &PowerPointConverter{
		builder: NewPowerPointBuilder(),
	}
}

// Convert converts markdown document to PowerPoint
func (p *PowerPointConverter) Convert(doc *Document) error {
	// If no sections (no H1), treat entire document as one slide
	if len(doc.Sections) == 0 {
		// Create a title slide with all content
		p.builder.AddTitleSlide("Presentation", "")
		
		// Add content slides for remaining blocks
		if len(doc.Blocks) > 0 {
			p.convertBlocksToSlides(doc.Blocks)
		}
		return nil
	}
	
	// Convert sections to slides
	for i, section := range doc.Sections {
		if i == 0 && section.Title != "" {
			// First section becomes title slide
			subtitle := ""
			if len(section.Blocks) > 0 && section.Blocks[0].Type == BlockParagraph {
				subtitle = section.Blocks[0].Content
			}
			p.builder.AddTitleSlide(section.Title, subtitle)
			
			// Add remaining blocks as content slides
			startIdx := 0
			if subtitle != "" {
				startIdx = 1
			}
			if len(section.Blocks) > startIdx {
				p.convertBlocksToSlides(section.Blocks[startIdx:])
			}
		} else {
			// Regular content slide
			p.convertSectionToSlide(section)
		}
	}
	
	return nil
}

// SaveAs saves the PowerPoint presentation to the specified path
func (p *PowerPointConverter) SaveAs(path string) error {
	if p.builder == nil {
		return fmt.Errorf("no presentation to save")
	}
	
	return p.builder.Build(path)
}

// convertSectionToSlide converts a section to a slide
func (p *PowerPointConverter) convertSectionToSlide(section Section) {
	// Create slide with title
	slide := &Slide{
		Title: section.Title,
	}
	
	// Add content from blocks
	for _, block := range section.Blocks {
		switch block.Type {
		case BlockHeading:
			if block.Level == 2 && slide.Title == "" {
				// H2 can be slide title if no H1
				slide.Title = block.Content
			} else {
				// Other headings become content
				slide.Content = append(slide.Content, fmt.Sprintf("**%s**", block.Content))
			}
			
		case BlockParagraph:
			slide.Content = append(slide.Content, block.Content)
			
		case BlockList:
			for _, item := range block.Items {
				slide.Bullets = append(slide.Bullets, item)
			}
			
		case BlockOrderedList:
			for i, item := range block.Items {
				slide.Bullets = append(slide.Bullets, fmt.Sprintf("%d. %s", i+1, item))
			}
			
		case BlockCodeBlock:
			slide.Content = append(slide.Content, fmt.Sprintf("```\n%s\n```", block.Content))
			
		case BlockQuote:
			slide.Content = append(slide.Content, fmt.Sprintf("> %s", block.Content))
		}
	}
	
	p.builder.AddContentSlide(slide)
}

// convertBlocksToSlides converts blocks to slides
// H2 headers become content within the current slide, not new slides
func (p *PowerPointConverter) convertBlocksToSlides(blocks []Block) {
	if len(blocks) == 0 {
		return
	}
	
	slide := &Slide{}
	hasContent := false
	
	for _, block := range blocks {
		switch block.Type {
		case BlockHeading:
			// H2 and below become content, not new slides
			if block.Level == 2 {
				// H2 becomes bold content
				slide.Content = append(slide.Content, fmt.Sprintf("**%s**", block.Content))
			} else if block.Level > 2 {
				// H3 and below become regular bold content
				slide.Content = append(slide.Content, fmt.Sprintf("**%s**", block.Content))
			}
			hasContent = true
			
		case BlockParagraph:
			slide.Content = append(slide.Content, block.Content)
			hasContent = true
			
		case BlockList:
			for _, item := range block.Items {
				slide.Bullets = append(slide.Bullets, item)
			}
			hasContent = true
			
		case BlockOrderedList:
			for i, item := range block.Items {
				slide.Bullets = append(slide.Bullets, fmt.Sprintf("%d. %s", i+1, item))
			}
			hasContent = true
			
		case BlockCodeBlock:
			slide.Content = append(slide.Content, fmt.Sprintf("```\n%s\n```", block.Content))
			hasContent = true
			
		case BlockQuote:
			slide.Content = append(slide.Content, fmt.Sprintf("> %s", block.Content))
			hasContent = true
		}
	}
	
	// Add the slide if it has any content
	if hasContent {
		p.builder.AddContentSlide(slide)
	}
}