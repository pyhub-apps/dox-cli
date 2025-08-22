# dox - Document Automation CLI 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/pyhub-kr/pyhub-documents-cli?include_prereleases)](https://github.com/pyhub-kr/pyhub-documents-cli/releases)
[![HeadVer](https://img.shields.io/badge/versioning-HeadVer-blue)](https://github.com/line/headver)
[![Issues](https://img.shields.io/github/issues/pyhub-kr/pyhub-documents-cli)](https://github.com/pyhub-kr/pyhub-documents-cli/issues)

A powerful CLI tool for document automation, text replacement, and AI-powered content generation. Process Word/PowerPoint documents efficiently with beautiful progress tracking and colored output.

[한국어 문서](README.md) | English

## ✨ Features

### 🔄 Bulk Text Replacement
- Replace text across multiple Word (.docx) and PowerPoint (.pptx) files
- YAML-based rule configuration for easy management
- Recursive directory processing with pattern exclusion
- Concurrent processing for improved performance (40-70% faster)
- Automatic backup creation before modifications
- Progress bars and colored output for better UX

### 📝 Document Creation
- Convert Markdown files to Word or PowerPoint documents
- Template-based document generation
- Style and format preservation
- Support for complex document structures
- Code blocks, lists, tables, and more

### 🤖 AI Content Generation
- Generate blog posts, reports, and summaries using OpenAI
- Multiple content types and customizable parameters
- Temperature and token control for output fine-tuning
- Support for GPT-3.5 and GPT-4 models
- Configuration file support for API keys

### 📋 Template Processing
- Process Word/PowerPoint templates with placeholders
- YAML/JSON-based value injection
- Support for complex data structures
- Validation and missing placeholder detection
- Batch processing capabilities

### 🎨 Beautiful UI
- Colored output for better readability
- Progress bars for long operations
- Loading spinners for AI operations
- File type-specific coloring
- Summary statistics with visual formatting
- Support for NO_COLOR environment variable

### 🌍 Internationalization
- Full support for English and Korean interfaces
- Automatic language detection based on system locale
- Easy language switching with --lang flag

## 📦 Installation

### Quick Install (Pre-built Binaries)

#### macOS/Linux
```bash
# Download the latest release
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-$(uname -s)-$(uname -m) -o dox

# Make it executable
chmod +x dox

# Move to PATH (optional)
sudo mv dox /usr/local/bin/
```

#### Windows
Download the latest `.exe` from [Releases](https://github.com/pyhub-kr/pyhub-documents-cli/releases) and add to your PATH.

```powershell
# PowerShell
Invoke-WebRequest -Uri "https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-windows-amd64.exe" -OutFile "dox.exe"
Move-Item dox.exe C:\Windows\System32\
```

### Build from Source
```bash
# Clone the repository
git clone https://github.com/pyhub-kr/pyhub-documents-cli.git
cd pyhub-documents-cli

# Build
go build -o dox

# Or install globally
go install
```

## 🚀 Quick Start

### Text Replacement
```bash
# Create a rules file (rules.yml)
cat > rules.yml << EOF
- old: "2023"
  new: "2024"
- old: "Version 1.0"
  new: "Version 2.0"
- old: "Company A"
  new: "Company B"
EOF

# Replace in a single file
dox replace --rules rules.yml --path document.docx

# Replace in all documents in a directory
dox replace --rules rules.yml --path ./docs --recursive

# Preview changes without applying
dox replace --rules rules.yml --path ./docs --dry-run

# Create backups before modification
dox replace --rules rules.yml --path ./docs --backup

# Use concurrent processing for better performance
dox replace --rules rules.yml --path ./docs --concurrent --max-workers 8
```

### Document Creation
```bash
# Convert Markdown to Word
dox create --from report.md --output report.docx

# Convert Markdown to PowerPoint
dox create --from presentation.md --output slides.pptx

# Use a template for styling
dox create --from content.md --template company-template.docx --output final.docx

# Force overwrite existing files
dox create --from report.md --output report.docx --force
```

### Template Processing
```bash
# Create a template with placeholders
# In your Word/PowerPoint: {{name}}, {{date}}, {{amount}}

# Create values file (values.yml)
cat > values.yml << EOF
name: "John Doe"
date: "2024-01-01"
amount: "$1,000"
items:
  - "Item 1"
  - "Item 2"
  - "Item 3"
EOF

# Process template
dox template --template invoice.docx --values values.yml --output invoice-final.docx

# Use inline values
dox template --template report.pptx --output final.pptx \
  --set title="Q4 Report" \
  --set year="2024" \
  --set author="Jane Smith"
```

### AI Content Generation
```bash
# Set OpenAI API key
export OPENAI_API_KEY="your-api-key"

# Or use configuration file
dox config --set openai.api_key "your-api-key"

# Generate a blog post
dox generate --type blog --prompt "Best practices for Go testing" --output blog.md

# Generate a report with GPT-4
dox generate --type report --prompt "Q3 sales analysis" --model gpt-4 --output report.md

# Generate with custom parameters
dox generate --type custom \
  --prompt "Write a technical tutorial about Docker" \
  --temperature 0.7 \
  --max-tokens 2000 \
  --output tutorial.md
```

### Configuration Management
```bash
# Initialize configuration
dox config --init

# List all settings
dox config --list

# Set a configuration value
dox config --set openai.api_key "your-key"
dox config --set global.lang "en"
dox config --set replace.concurrent true

# Get a configuration value
dox config --get openai.model
```

## ⚙️ Configuration

dox supports both command-line flags and configuration files. The precedence order is:
1. Command-line flags (highest priority)
2. Configuration file
3. Environment variables (lowest priority)

### Configuration File

Create `~/.pyhub/config.yml`:

```yaml
# OpenAI settings
openai:
  api_key: "your-api-key"  # Or use OPENAI_API_KEY env var
  base_url: "https://api.openai.com/v1"
  timeout: 120
  max_retries: 3

# Document replacement settings
replace:
  backup: true
  recursive: true
  concurrent: true
  max_workers: 8

# Content generation settings
generate:
  model: "gpt-3.5-turbo"
  max_tokens: 2000
  temperature: 0.7
  content_type: "blog"

# Global settings
global:
  verbose: false
  quiet: false
  lang: "en"  # or "ko" for Korean
```

## 📊 Command Reference

### Global Flags
- `--config` - Specify config file location
- `--verbose, -v` - Verbose output with detailed information
- `--quiet, -q` - Suppress non-error output
- `--no-color` - Disable colored output
- `--lang` - Set interface language (en, ko)

### Commands

#### `replace` - Text replacement in documents
```bash
dox replace --rules <file> --path <path> [flags]
```

**Flags:**
- `--rules, -r` - YAML file with replacement rules (required)
- `--path, -p` - Target file or directory (required)
- `--dry-run` - Preview changes without applying
- `--backup` - Create backups before modification
- `--recursive` - Process subdirectories (default: true)
- `--exclude` - Glob pattern for exclusion
- `--concurrent` - Enable concurrent processing
- `--max-workers` - Number of workers (default: CPU count)

#### `create` - Create documents from Markdown
```bash
dox create --from <file> --output <file> [flags]
```

**Flags:**
- `--from, -f` - Input Markdown file (required)
- `--output, -o` - Output document path (required)
- `--template, -t` - Template document for styling
- `--format` - Output format (docx, pptx)
- `--force` - Overwrite existing files

#### `template` - Process document templates
```bash
dox template --template <file> --output <file> [flags]
```

**Flags:**
- `--template, -t` - Template file path (required)
- `--output, -o` - Output file path (required)
- `--values` - YAML/JSON file with values
- `--set` - Set individual values (key=value)
- `--force` - Overwrite existing files

#### `generate` - AI content generation
```bash
dox generate --prompt <text> [flags]
```

**Flags:**
- `--prompt, -p` - Generation prompt (required)
- `--type, -t` - Content type (blog, report, summary, custom)
- `--output, -o` - Output file path
- `--model` - AI model (gpt-3.5-turbo, gpt-4)
- `--max-tokens` - Maximum response tokens
- `--temperature` - Creativity level (0.0-1.0)
- `--api-key` - OpenAI API key

#### `config` - Configuration management
```bash
dox config [flags]
```

**Flags:**
- `--init` - Initialize configuration file
- `--list` - List all configuration values
- `--get <key>` - Get a specific value
- `--set <key=value>` - Set a configuration value

## 📁 Examples

### PowerPoint Generation from Markdown

Create a presentation from markdown:

```markdown
# Project Status Update

## Completed Tasks
- Feature A implemented
- Bug fixes completed
- Documentation updated

# Next Steps

## Q1 Goals
- Launch beta version
- User testing
- Performance optimization

## Q2 Planning
- Scale infrastructure
- Add new features
- International expansion
```

Convert to PowerPoint:
```bash
dox create --from status.md --output status.pptx
```

### Batch Document Processing

Process multiple documents with different rules:

```bash
#!/bin/bash
# batch-process.sh

# Update year in all documents
dox replace --rules year-update.yml --path ./reports --concurrent

# Update company name
dox replace --rules company-update.yml --path ./contracts --backup

# Generate summary report
dox generate --type report \
  --prompt "Summarize all changes made today" \
  --output changes-summary.md

# Convert to Word
dox create --from changes-summary.md --output changes-summary.docx
```

### CI/CD Integration

```yaml
# .github/workflows/docs.yml
name: Document Processing
on:
  push:
    paths:
      - 'docs/**'
jobs:
  process:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dox
        run: |
          curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-Linux-x86_64 -o dox
          chmod +x dox
      - name: Process documents
        run: |
          ./dox replace --rules ci-rules.yml --path docs/
          ./dox create --from CHANGELOG.md --output CHANGELOG.docx
      - name: Upload artifacts
        uses: actions/upload-artifact@v2
        with:
          name: processed-docs
          path: |
            docs/
            CHANGELOG.docx
```

## 🔧 Advanced Usage

### Performance Optimization

For large document sets, use concurrent processing:
```bash
# Process with 16 workers
dox replace --rules rules.yml --path ./large-docs \
  --concurrent --max-workers 16

# Monitor progress with verbose output
dox replace --rules rules.yml --path ./docs \
  --concurrent --verbose
```

### Complex Template Processing

```yaml
# template-data.yml
company:
  name: "Tech Corp"
  address: "123 Main St"
  
invoice:
  number: "INV-2024-001"
  date: "2024-01-01"
  
items:
  - name: "Service A"
    quantity: 10
    price: 100
  - name: "Service B"
    quantity: 5
    price: 200
    
totals:
  subtotal: 2000
  tax: 200
  total: 2200
```

```bash
dox template --template invoice-template.docx \
  --values template-data.yml \
  --output invoice-2024-001.docx
```

## 🔢 Versioning (HeadVer)

This project uses [HeadVer](https://github.com/line/headver) versioning.

### Version Format
```
{head}.{yearweek}.{build}
```

- **head**: Major version (manually managed, incremented on breaking changes)
- **yearweek**: Year (2 digits) + Week number (2 digits) - auto-generated
- **build**: Build number within the week - auto-generated

### Examples
- `1.2534.0`: Version 1, Year 2025 Week 34, First build
- `1.2534.5`: Same week, 5th build
- `2.2601.0`: Version 2 (Breaking Change), Year 2026 Week 1

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Setup
```bash
# Clone the repository
git clone https://github.com/pyhub-kr/pyhub-documents-cli.git
cd pyhub-documents-cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o dox

# Run with debug output
./dox --verbose [command]
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [unioffice](https://github.com/unidoc/unioffice) - Office document processing
- [goldmark](https://github.com/yuin/goldmark) - Markdown parsing
- [progressbar](https://github.com/schollz/progressbar) - Progress indicators
- [color](https://github.com/fatih/color) - Terminal colors

## 📞 Support

- 📧 Email: support@pyhub.kr
- 🐛 Issues: [GitHub Issues](https://github.com/pyhub-kr/pyhub-documents-cli/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/pyhub-kr/pyhub-documents-cli/discussions)

## 🗺️ Roadmap

- [ ] Excel file support (.xlsx)
- [ ] PDF generation and processing
- [ ] HWP (Hangul) format support
- [ ] Cloud storage integration (S3, Google Drive)
- [ ] Web UI interface
- [ ] Plugin system for extensibility
- [ ] More AI providers (Claude, Gemini, Local LLMs)
- [ ] Document comparison and diff
- [ ] Batch processing improvements
- [ ] Docker container support

---

Made with ❤️ by [PyHub Korea](https://pyhub.kr)