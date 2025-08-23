# Getting Started with dox

This guide will help you install and start using dox for document automation.

## 📦 Installation

### Quick Install (Recommended)

#### macOS/Linux
```bash
curl -fsSL https://raw.githubusercontent.com/pyhub/pyhub-docs/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/pyhub/pyhub-docs/main/install.ps1 | iex
```

### Manual Installation

1. Download the appropriate binary for your platform from [Releases](https://github.com/pyhub/pyhub-docs/releases)
2. Extract the archive
3. Move the binary to a directory in your PATH
4. Verify installation: `dox version`

### Building from Source

```bash
# Clone repository
git clone https://github.com/pyhub/pyhub-docs.git
cd pyhub-docs

# Build
go build -o dox

# Install
sudo mv dox /usr/local/bin/
```

## 🚀 First Steps

### 1. Check Installation
```bash
dox version
dox --help
```

### 2. Initialize Configuration
```bash
# Create default config file
dox config --init

# View current configuration
dox config --list
```

### 3. Your First Document Operation

#### Replace Text in Documents
```bash
# Create a simple rule file
cat > rules.yml << EOF
- old: "2024"
  new: "2025"
- old: "old@email.com"
  new: "new@email.com"
EOF

# Apply to documents
dox replace --rules rules.yml --path ./documents
```

#### Convert Markdown to Word
```bash
# Create a markdown file
cat > report.md << EOF
# Monthly Report

## Summary
This month we achieved significant progress.

## Details
- Task 1: Completed
- Task 2: In Progress
EOF

# Convert to Word
dox create --from report.md --output report.docx
```

## 🔑 Setting Up AI Features

To use AI generation features, you need an OpenAI API key:

### 1. Get an API Key
Visit [OpenAI Platform](https://platform.openai.com/api-keys) to create an API key.

### 2. Configure the Key

#### Option 1: Environment Variable
```bash
export OPENAI_API_KEY="your-api-key-here"
```

#### Option 2: Config File
```bash
dox config --set openai.api_key="your-api-key-here"
```

### 3. Test AI Features
```bash
dox generate --type summary --prompt "Summarize the benefits of automation" --output summary.md
```

## 📚 Core Concepts

### Documents
dox works with:
- **Word Documents** (.docx)
- **PowerPoint Presentations** (.pptx)
- **Markdown Files** (.md)
- **Text Files** (.txt)

### Operations
Main operations you can perform:
- **Replace**: Bulk text replacement across documents
- **Create**: Convert between formats
- **Template**: Fill templates with data
- **Generate**: Create content with AI

### Rule Files
YAML files that define replacement rules:
```yaml
- old: "search text"
  new: "replacement text"
```

### Template Variables
Use variables in templates:
- Word/PowerPoint: `{{variable_name}}`
- Values provided via YAML/JSON files

## 🎯 Common Use Cases

### Annual Updates
Update year references across all company documents:
```bash
dox replace --rules year-update.yml --path ./company-docs --recursive
```

### Report Generation
Convert markdown reports to professional Word documents:
```bash
dox create --from quarterly-report.md --template company-template.docx --output Q4-Report.docx
```

### Invoice Creation
Generate invoices from templates:
```bash
dox template --template invoice-template.docx --values client-data.yml --output invoice-2025-001.docx
```

### AI Content Creation
Generate blog posts or summaries:
```bash
dox generate --type blog --prompt "Cloud computing trends in 2025" --output blog-post.md
```

## 🔧 Configuration Tips

### Default Paths
Set default document directory:
```bash
dox config --set defaults.path="./documents"
```

### Backup Settings
Always create backups:
```bash
dox config --set replace.backup=true
```

### Performance Tuning
Enable concurrent processing:
```bash
dox config --set performance.concurrent=true
```

## 📖 Next Steps

- Read the [Command Reference](./commands.md) for detailed command documentation
- Explore [Templates Guide](./templates.md) for advanced template usage
- Check [Examples](../examples/) for real-world scenarios
- Learn about [AI Generation](./ai-generation.md) capabilities

## 🆘 Getting Help

If you encounter issues:
1. Check error messages - they include helpful suggestions
2. Use `--help` flag with any command
3. Visit our [GitHub Issues](https://github.com/pyhub/pyhub-docs/issues)
4. See [Troubleshooting Guide](./troubleshooting.md)