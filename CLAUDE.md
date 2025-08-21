# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**pyhub-documents-cli** is a Go-based CLI application for document automation and AI-powered content generation. It automates repetitive document editing tasks and integrates OpenAI capabilities for content generation directly from the command line.

## Agent System

This project uses a specialized agent system for optimal development support. See `.claude/AGENTS.md` for the complete agent architecture. Key agents include:
- **GoMaster**: Go language expertise and best practices
- **DocProcessor**: Document format handling (OOXML, templates)
- **CLIArchitect**: CLI/UX design and command structure
- **AIIntegrator**: OpenAI API integration and prompt engineering
- **TestGuardian**: Testing strategies and quality assurance

Claude Code will automatically activate the appropriate agent based on the task context.

## Core Features & Implementation Priorities

### Phase 1 (MVP) - Current Focus
1. **Bulk Content Replacement** (`replace` command)
   - Replace text across Word/PPT documents using YAML rule files
   - Support recursive directory processing
   
2. **Document Conversion** (`create` command)
   - Convert Markdown to Word (.docx) and PowerPoint (.pptx)
   - Support template-based generation with `--template` flag
   
3. **Basic CLI Structure**
   - Build single Windows executable (.exe) file
   - Implement intuitive command structure with `--help` support

### Phase 2
- **AI Content Generation** (`generate` command) with OpenAI integration
- HWP format support research

## Technical Requirements

### Language & Framework
- **Language**: Go (Golang)
- **CLI Framework**: Cobra or urfave/cli
- **Target Platform**: Windows (amd64) initially
- **Distribution**: Single executable without dependencies

### Key Libraries (Open-source only)
- **Word Processing**: `nguyenthenguyen/docx` or `baliance/gooxml`
- **PowerPoint Processing**: Consider `officenum/go-prezi`
- **Markdown Parsing**: `goldmark` or `gomarkdown/markdown`
- **YAML Parsing**: `gopkg.in/yaml.v3`

### Build Considerations
- Use build flags to reduce antivirus false positives: `-ldflags="-s -w"`
- Avoid UPX packing to prevent AV detection issues
- Focus on text-based processing in MVP (complex objects like images/charts are out of scope)

## Development Commands

### Project Setup
```bash
# Initialize Go module
go mod init github.com/pyhub/pyhub-documents-cli

# Install dependencies (after adding to go.mod)
go mod tidy
```

### Build Commands
```bash
# Build for current platform
go build -o pyhub-documents-cli

# Build Windows executable (from any platform)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o pyhub-documents-cli.exe

# Run tests
go test ./...

# Run specific test
go test -run TestFunctionName ./package_name
```

## Project Structure Recommendations

```
pyhub-documents-cli/
├── cmd/
│   └── root.go          # Main CLI entry point and command definitions
├── internal/
│   ├── docx/            # Word document processing
│   ├── pptx/            # PowerPoint processing
│   ├── markdown/        # Markdown parsing
│   └── replace/         # Bulk replacement logic
├── pkg/                 # Exportable library packages
│   └── documents/       # Public API for document operations
├── go.mod
├── go.sum
└── main.go
```

## Command Interface Design

### Replace Command
```bash
pyhub-documents-cli replace --rules vars.yml --path ./docs
```

### Create Command
```bash
pyhub-documents-cli create --from report.md --template template.docx --output output.docx
```

### Generate Command (Phase 2)
```bash
pyhub-documents-cli generate --type blog --prompt "..." --output draft.md
```

## Implementation Notes

1. **Library Package**: Core functionality should be implemented as importable Go packages under `pkg/` for reuse in other applications

2. **Error Handling**: Use clear error messages that help users understand what went wrong and how to fix it

3. **YAML Rule File Format**:
```yaml
- old: "old_text"
  new: "new_text"
- old: "v1.2.0"
  new: "v1.3.0"
```

4. **Template Processing**: When using `--template`, preserve all formatting/styles from the template document while replacing only the content

5. **OpenAI Integration** (Phase 2): Support API key configuration via `OPENAI_API_KEY` environment variable or CLI flag