# pyhub-docs

[![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/pyhub-kr/pyhub-docs)](https://github.com/pyhub-kr/pyhub-docs/releases)
[![HeadVer](https://img.shields.io/badge/versioning-HeadVer-blue)](https://github.com/line/headver)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Document automation and AI-powered content generation CLI tool for developers and content creators.

## 🎯 Features

- ✅ **Document Conversion**: Convert Markdown to Word (.docx) and PowerPoint (.pptx)
- ✅ **Bulk Text Replacement**: Replace text across multiple Word and PowerPoint documents using YAML rules
- ✅ **Template Processing**: Use Word/PowerPoint templates with placeholder replacement
- 🌍 **Internationalization (i18n)**: English and Korean language support with automatic locale detection
- 🤖 **AI Content Generation**: Generate content using OpenAI (Phase 2)
- 🚀 **Cross-Platform**: Single binary with no dependencies

## 📦 Installation

### Download Binary

Download the latest release (v1.2534.0) for your platform from the [releases page](https://github.com/pyhub-kr/pyhub-docs/releases).

#### Quick Install

**Windows (PowerShell)**:
```powershell
Invoke-WebRequest -Uri "https://github.com/pyhub-kr/pyhub-docs/releases/latest/download/pyhub-docs.exe" -OutFile "pyhub-docs.exe"
```

**macOS/Linux**:
```bash
# macOS Intel
curl -L -o pyhub-docs https://github.com/pyhub-kr/pyhub-docs/releases/latest/download/pyhub-docs-darwin-amd64

# macOS Apple Silicon
curl -L -o pyhub-docs https://github.com/pyhub-kr/pyhub-docs/releases/latest/download/pyhub-docs-darwin-arm64

# Linux
curl -L -o pyhub-docs https://github.com/pyhub-kr/pyhub-docs/releases/latest/download/pyhub-docs-linux-amd64

chmod +x pyhub-docs
sudo mv pyhub-docs /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/pyhub/pyhub-docs.git
cd pyhub-docs

# Build
make build

# Or build for specific platform
make build-windows  # Windows
make build-darwin   # macOS
make build-linux    # Linux
```

## 🚀 Quick Start

### Replace Text in Documents

Create a YAML file with replacement rules:
```yaml
# rules.yml
- old: "v1.0.0"
  new: "v2.0.0"
- old: "2023"
  new: "2024"
```

Run the replacement:
```bash
pyhub-docs replace --rules rules.yml --path ./docs
```

### Create Document from Markdown

Convert Markdown files to Word or PowerPoint:

```bash
# Convert to Word document
pyhub-docs create --from report.md --output report.docx

# Convert to PowerPoint presentation
pyhub-docs create --from slides.md --output presentation.pptx

# Format is auto-detected from extension, or specify explicitly
pyhub-docs create --from content.md --output output.docx --format docx

# With template (Coming Soon)
pyhub-docs create --from content.md --template company.docx --output final.docx
```

**Markdown to PowerPoint Conversion:**
- H1 headers (`#`) create new slides
- H2 headers (`##`) become slide titles when first in a section, otherwise bold content
- H3-H6 headers become bold content within slides
- Lists, paragraphs, code blocks, and quotes are preserved as slide content

**Markdown to Word Conversion:**
- All Markdown elements are converted to Word formatting
- Heading hierarchy is preserved
- Lists, code blocks, and quotes are styled appropriately

### Generate AI Content (Coming Soon)

```bash
pyhub-docs generate --type blog --prompt "Go best practices" --output blog.md
```

## 🌍 Internationalization (i18n)

The CLI supports multiple languages with automatic locale detection from your system settings.

### Supported Languages
- 🇺🇸 English (en)
- 🇰🇷 Korean (ko)

### Language Selection
The CLI automatically detects your system language from environment variables (`LANG`, `LC_ALL`). You can also explicitly set the language:

```bash
# Use Korean interface
pyhub-docs --lang ko replace --rules rules.yml --path ./docs

# Use English interface
pyhub-docs --lang en create --from report.md --output report.docx

# Auto-detect from system (default)
pyhub-docs replace --rules rules.yml --path ./docs
```

### System Language Detection Priority
1. `--lang` flag (highest priority)
2. `LANG` environment variable
3. `LC_ALL` environment variable
4. Default to English

### Examples
```bash
# Korean system (LANG=ko_KR.UTF-8)
$ pyhub-docs create --from report.md --output report.docx
report.md를 Word 문서로 변환 중...
✅ report.docx 생성 완료

# English system or with --lang en
$ pyhub-docs --lang en create --from report.md --output report.docx
Converting report.md to Word document...
✅ Successfully created report.docx
```

## 📚 Documentation

- [User Guide](docs/user-guide.md) (Coming soon)
- [API Reference](docs/api-reference.md) (Coming soon)
- [Examples](docs/examples.md) (Coming soon)

## 🛠️ Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Project Structure

```
pyhub-docs/
├── cmd/            # CLI commands
├── internal/       # Internal packages
├── pkg/            # Public packages
└── tests/          # Test files and fixtures
```

### Building

```bash
# Run tests
make test

# Run with coverage
make coverage

# Format code
make fmt

# Run linter
make lint

# Build all platforms
make build-all
```

### Testing

We follow Test-Driven Development (TDD):
1. Write failing tests first
2. Implement functionality
3. Refactor while keeping tests green

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific test
go test -run TestFunctionName ./package
```

## 🗺️ Roadmap

### Phase 1: MVP (Current)
- [x] Project setup and CLI structure
- [x] Text replacement in Word documents
- [x] Text replacement in PowerPoint
- [x] Markdown to Word conversion
- [x] Markdown to PowerPoint conversion
- [x] Template-based generation with placeholder replacement

### Phase 2: AI Integration
- [ ] OpenAI API integration
- [ ] Content generation commands
- [ ] Prompt templates

### Phase 3: Advanced Features
- [ ] HWP format support
- [ ] Batch processing optimization
- [ ] Plugin system

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Process

1. Check existing issues or create a new one
2. Fork the repository
3. Create a feature branch
4. Write tests first (TDD)
5. Implement the feature
6. Submit a pull request

## 🔢 Versioning

This project uses [HeadVer](https://github.com/line/headver) versioning system:
- **Format**: `{head}.{yearweek}.{build}`
- **Example**: v1.2534.0 = Head version 1, Year 2025 Week 34, Build 0
- **Benefits**: Clear release timeline, sprint alignment, automatic build tracking

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Document processing powered by open-source Go libraries
- AI features powered by OpenAI

## 📞 Support

- Create an [issue](https://github.com/pyhub/pyhub-docs/issues) for bugs or features
- Check [discussions](https://github.com/pyhub/pyhub-docs/discussions) for Q&A

---

**Note**: This project is under active development. Features marked as "Coming Soon" are planned for future releases.