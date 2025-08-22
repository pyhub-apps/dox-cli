# dox Examples

This directory contains practical examples demonstrating various features of the dox CLI tool.

## 📁 Directory Structure

```
examples/
├── replace/          # Text replacement examples
├── create/          # Markdown to document conversion examples  
├── template/        # Template processing examples
├── generate/        # AI content generation examples
└── scripts/         # Automation scripts
```

## 🔄 Text Replacement (`replace/`)

### Basic Rules (`rules-basic.yml`)
Simple text replacement rules for common use cases like updating years, versions, and company information.

```bash
# Preview changes
dox replace --rules examples/replace/rules-basic.yml --path ./docs --dry-run

# Apply changes with backup
dox replace --rules examples/replace/rules-basic.yml --path ./docs --backup
```

### Korean Rules (`rules-korean.yml`)
Text replacement rules for Korean documents with proper encoding support.

```bash
# Process Korean documents
dox --lang ko replace --rules examples/replace/rules-korean.yml --path ./문서
```

## 📝 Document Creation (`create/`)

### Presentation (`presentation.md`)
A complete presentation example that converts to PowerPoint format.

```bash
# Convert to PowerPoint
dox create --from examples/create/presentation.md --output presentation.pptx

# View the structure
cat examples/create/presentation.md
```

### Report (`report.md`)
A comprehensive report template with tables, lists, and formatting.

```bash
# Convert to Word document
dox create --from examples/create/report.md --output monthly-report.docx
```

## 📋 Template Processing (`template/`)

### Invoice Template (`invoice-values.yml`)
Complete invoice data in YAML format for template processing.

```bash
# Process invoice template
dox template \
  --template invoice-template.docx \
  --values examples/template/invoice-values.yml \
  --output invoice-2025-0142.docx
```

### Contract Template (`contract-values.json`)
Contract data in JSON format showing alternative data format support.

```bash
# Process contract template
dox template \
  --template contract-template.docx \
  --values examples/template/contract-values.json \
  --output contract-final.docx
```

## 🤖 AI Content Generation (`generate/`)

### Prompts Collection (`prompts.txt`)
Various prompt examples for different content types and use cases.

```bash
# Generate a blog post
dox generate --type blog \
  --prompt "Best practices for Go testing" \
  --output blog-post.md

# Generate with custom parameters
dox generate --type report \
  --prompt "Q4 2024 performance analysis" \
  --model gpt-4 \
  --temperature 0.7 \
  --max-tokens 2000 \
  --output performance-report.md
```

## 🔧 Automation Scripts (`scripts/`)

### Batch Processing (`batch-process.sh`)
Complete workflow script demonstrating multiple dox commands in sequence.

```bash
# Run the batch processing script
./examples/scripts/batch-process.sh

# Or source it for step-by-step execution
source examples/scripts/batch-process.sh
```

The script performs:
1. Creates backups of all documents
2. Updates year references
3. Updates company information
4. Generates changelog
5. Processes templates
6. Creates AI summaries (if API key is set)
7. Produces final report

## 🚀 Quick Start Examples

### Example 1: Update All Documents for New Year
```bash
# Create rules file
cat > year-update.yml << EOF
- old: "2024"
  new: "2025"
- old: "FY2024"
  new: "FY2025"
EOF

# Process all documents
dox replace --rules year-update.yml --path ./documents --recursive --backup
```

### Example 2: Create Presentation from Markdown
```bash
# Write presentation in markdown
echo "# My Presentation

## Slide 1
- Point 1
- Point 2

# Slide 2
## Content
More content here" > presentation.md

# Convert to PowerPoint
dox create --from presentation.md --output presentation.pptx
```

### Example 3: Process Multiple Templates
```bash
# Process all templates in a directory
for template in templates/*.docx; do
  output="output/$(basename $template .docx)-$(date +%Y%m%d).docx"
  dox template --template "$template" \
    --values common-values.yml \
    --output "$output" \
    --set date="$(date +"%B %d, %Y")"
done
```

### Example 4: Generate Multiple Reports
```bash
# Generate reports for different departments
departments=("Sales" "Marketing" "Engineering" "Support")

for dept in "${departments[@]}"; do
  dox generate --type report \
    --prompt "Generate Q4 2024 report for $dept department" \
    --output "reports/${dept,,}-q4-2024.md"
  
  # Convert to Word
  dox create --from "reports/${dept,,}-q4-2024.md" \
    --output "reports/${dept,,}-q4-2024.docx"
done
```

## 🎯 Best Practices

1. **Always Preview First**: Use `--dry-run` flag to preview changes before applying
2. **Create Backups**: Use `--backup` flag when modifying important documents
3. **Use Configuration Files**: Store common settings in `~/.pyhub/config.yml`
4. **Batch Operations**: Use `--concurrent` flag for better performance with multiple files
5. **Version Control Rules**: Keep your replacement rules in version control
6. **Template Validation**: Test templates with sample data before production use

## 📚 Additional Resources

- [Main Documentation](../README.md)
- [Command Reference](../docs/commands.md)
- [API Documentation](../docs/api.md)
- [Contributing Guide](../CONTRIBUTING.md)

## 💡 Tips

- Use `--verbose` flag for detailed output during debugging
- Set `OPENAI_API_KEY` environment variable for AI features
- Use `--lang ko` for Korean interface
- Combine multiple rules files: `cat rules1.yml rules2.yml > combined.yml`
- Process files in parallel: `--concurrent --max-workers 8`

---

For more examples and use cases, visit the [GitHub repository](https://github.com/pyhub-kr/pyhub-documents-cli).