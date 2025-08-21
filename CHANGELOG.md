# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [HeadVer](https://github.com/line/headver) versioning.

## [v1.2534.0] - 2025-08-21

### 🚀 Major Release with Template Support

This release introduces powerful template-based document generation and Markdown conversion features, completing the MVP phase of the project.

### ✨ New Features

#### Template-Based Document Generation (#12)
- **New `template` command** for processing Word/PowerPoint templates
- **Placeholder replacement** with `{{placeholder}}` syntax
- **Nested value support** (e.g., `{{departments.sales.head}}`)
- **Array value support** with automatic list formatting
- **Multiple input methods**: YAML/JSON files or inline CLI flags (`--set`)
- **Template validation** with missing placeholder warnings

#### Markdown to Office Document Conversion (#11)
- **New `create` command** for document conversion
- **Markdown to Word (.docx)** conversion with full formatting
- **Markdown to PowerPoint (.pptx)** conversion
  - H1 headers create new slides
  - H2 headers become slide titles or bold content
  - Lists, code blocks, and quotes preserved
- **Structure preservation** during conversion

#### PowerPoint Text Replacement (#10)
- **PowerPoint support** in `replace` command
- **Slide content replacement** across all slides
- **Format preservation** while replacing text
- **Batch processing** for multiple .pptx files

### 🛡️ Security Improvements
- **Atomic file operations** for PowerPoint handler
- **Proper XML escaping** for all user inputs
- **Path traversal prevention** in file operations

### 🔧 Technical Improvements
- **Enhanced error handling** with clear user messages
- **Comprehensive test coverage** for all new features
- **Improved CLI help documentation**
- **Better code organization** with dedicated packages

### 📦 Installation
Binary releases available for Windows, macOS (Intel/ARM), and Linux.

## [0.1.0] - 2024-01-21

### 🎉 Initial Release

This is the first official release of pyhub-documents-cli, a powerful CLI tool for document automation with a focus on Word document text replacement.

### ✨ Features

#### Core Functionality
- **Text Replacement in Word Documents** (#3)
  - Replace text across multiple .docx files using YAML rules
  - Preserve formatting while replacing text
  - Support for batch processing of entire directories
  - Recursive directory processing option

#### Document Processing
- **Word Document Handler** (#2)
  - Read and write .docx files while preserving structure
  - Handle text across multiple runs
  - XML injection prevention for security
  - Full OOXML compliance

#### Configuration
- **YAML Rules Parser** (#1)
  - Define replacement rules in simple YAML format
  - Support for multiple replacement rules
  - Rule validation and error handling

#### CLI Features
- **Replace Command**
  - `--rules`: Specify YAML file with replacement rules
  - `--path`: Target file or directory
  - `--dry-run`: Preview changes without applying
  - `--backup`: Create backups before modification
  - `--recursive`: Process subdirectories

#### Build & Distribution
- **Automated Release Pipeline**
  - GitHub Actions workflow for automatic releases
  - Multi-platform binaries (Windows, macOS Intel/ARM, Linux)
  - SHA256 checksums for all binaries
  - Optimized binary size with stripped symbols

### 🛡️ Security
- XML injection vulnerability prevention in Word document processing
- Safe text escaping for special characters
- Input validation for all user inputs

### 📦 Installation

#### Windows
```powershell
# Download the executable
Invoke-WebRequest -Uri "https://github.com/pyhub-kr/pyhub-documents-cli/releases/download/v0.1.0/pyhub-documents-cli-windows-amd64.exe" -OutFile "pyhub-documents-cli.exe"
```

#### macOS
```bash
# Intel
curl -L -o pyhub-documents-cli https://github.com/pyhub-kr/pyhub-documents-cli/releases/download/v0.1.0/pyhub-documents-cli-darwin-amd64

# Apple Silicon
curl -L -o pyhub-documents-cli https://github.com/pyhub-kr/pyhub-documents-cli/releases/download/v0.1.0/pyhub-documents-cli-darwin-arm64

chmod +x pyhub-documents-cli
sudo mv pyhub-documents-cli /usr/local/bin/
```

#### Linux
```bash
curl -L -o pyhub-documents-cli https://github.com/pyhub-kr/pyhub-documents-cli/releases/download/v0.1.0/pyhub-documents-cli-linux-amd64
chmod +x pyhub-documents-cli
sudo mv pyhub-documents-cli /usr/local/bin/
```

### 👥 Contributors
- @allieus - Project lead and main contributor

### 🔗 Links
- [GitHub Repository](https://github.com/pyhub-kr/pyhub-documents-cli)
- [Issue Tracker](https://github.com/pyhub-kr/pyhub-documents-cli/issues)

---

## Version History Note

Starting from v1.2534.0, this project uses [HeadVer](https://github.com/line/headver) versioning:
- **Format**: `{head}.{yearweek}.{build}`
- **Example**: v1.2534.0 = Head version 1, Year 2025 Week 34, Build 0

Previous versions (v0.x.x) used Semantic Versioning.

---

[v1.2534.0]: https://github.com/pyhub-kr/pyhub-documents-cli/releases/tag/v1.2534.0
[0.1.0]: https://github.com/pyhub-kr/pyhub-documents-cli/releases/tag/v0.1.0