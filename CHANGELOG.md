# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/pyhub-kr/pyhub-documents-cli/releases/tag/v0.1.0