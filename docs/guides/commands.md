# Command Reference

Complete reference for all dox commands and their options.

## Global Flags

These flags work with all commands:

| Flag | Description | Example |
|------|-------------|---------|
| `--help, -h` | Show help for command | `dox --help` |
| `--version, -v` | Show version information | `dox --version` |
| `--verbose` | Enable verbose output | `dox --verbose replace ...` |
| `--quiet, -q` | Suppress non-error output | `dox -q create ...` |
| `--config` | Specify config file | `dox --config custom.yml ...` |

## Commands

### `dox replace`

Replace text in Word and PowerPoint documents.

#### Synopsis
```bash
dox replace --rules <file> --path <path> [flags]
```

#### Required Flags
- `--rules, -r` - YAML file with replacement rules
- `--path, -p` - Target file or directory

#### Optional Flags
| Flag | Description | Default |
|------|-------------|---------|
| `--recursive` | Process subdirectories | false |
| `--backup, -b` | Create backup files | false |
| `--dry-run` | Preview changes without applying | false |
| `--include` | File patterns to include | *.docx,*.pptx |
| `--exclude` | File patterns to exclude | none |
| `--concurrent` | Process files in parallel | false |
| `--max-workers` | Max concurrent workers | 4 |

#### Rule File Format
```yaml
# rules.yml
- old: "text to find"
  new: "replacement text"
- old: "{{2024}}"
  new: "{{2025}}"
  regex: false  # Optional: treat as literal text
```

#### Examples
```bash
# Basic replacement
dox replace --rules updates.yml --path document.docx

# Recursive with backup
dox replace -r rules.yml -p ./docs --recursive --backup

# Dry run to preview
dox replace --rules changes.yml --path . --dry-run

# Parallel processing
dox replace --rules bulk.yml --path ./reports --concurrent --max-workers 8
```

### `dox create`

Convert Markdown to Word or PowerPoint documents.

#### Synopsis
```bash
dox create --from <file> --output <file> [flags]
```

#### Required Flags
- `--from, -f` - Source Markdown file
- `--output, -o` - Output file (.docx or .pptx)

#### Optional Flags
| Flag | Description | Default |
|------|-------------|---------|
| `--template, -t` | Template document | none |
| `--style` | Style preset | default |
| `--metadata` | Document metadata JSON | none |
| `--toc` | Generate table of contents | false |

#### Markdown Extensions
- **Slides**: Use `---` to separate slides in PowerPoint
- **Speaker Notes**: Use `Notes:` prefix for speaker notes
- **Columns**: Use `:::columns` blocks
- **Custom Styles**: Use `{.style-name}` attributes

#### Examples
```bash
# Basic conversion
dox create --from report.md --output report.docx

# With template
dox create -f presentation.md -o slides.pptx -t company-template.pptx

# With metadata
dox create --from article.md --output article.docx --metadata meta.json
```

### `dox template`

Fill document templates with data from YAML/JSON files.

#### Synopsis
```bash
dox template --template <file> --values <file> --output <file> [flags]
```

#### Required Flags
- `--template, -t` - Template document with variables
- `--values, -v` - YAML/JSON file with values
- `--output, -o` - Output file

#### Optional Flags
| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Values file format | auto |
| `--missing` | Handle missing variables | error |
| `--set` | Set individual values | none |

#### Template Syntax
```
# In Word/PowerPoint documents:
Dear {{customer_name}},
Your invoice total is {{total_amount}}.

# Conditionals:
{{if premium_customer}}
  Thank you for being a premium member!
{{end}}

# Loops:
{{range items}}
  - {{.name}}: {{.price}}
{{end}}
```

#### Values File
```yaml
# values.yml
customer_name: "John Doe"
total_amount: "$1,234.56"
premium_customer: true
items:
  - name: "Product A"
    price: "$100"
  - name: "Product B"
    price: "$200"
```

#### Examples
```bash
# Basic template filling
dox template -t invoice.docx -v client.yml -o invoice-001.docx

# With individual values
dox template -t letter.docx -v base.yml -o output.docx --set "date=2025-01-15"

# JSON values
dox template --template report.pptx --values data.json --output final.pptx
```

### `dox generate`

Generate content using AI (OpenAI).

#### Synopsis
```bash
dox generate --type <type> --prompt <text> --output <file> [flags]
```

#### Required Flags
- `--type, -t` - Content type (blog, report, summary, email, proposal)
- `--prompt, -p` - Generation prompt
- `--output, -o` - Output file

#### Optional Flags
| Flag | Description | Default |
|------|-------------|---------|
| `--model, -m` | AI model to use | gpt-3.5-turbo |
| `--temperature` | Creativity (0.0-2.0) | 0.7 |
| `--max-tokens` | Maximum response length | 2000 |
| `--format` | Output format | markdown |
| `--language` | Output language | English |
| `--api-key` | OpenAI API key | env/config |

#### Content Types
- **blog**: Blog posts and articles
- **report**: Business reports
- **summary**: Document summaries
- **email**: Professional emails
- **proposal**: Business proposals
- **custom**: Custom content with full prompt control

#### Examples
```bash
# Generate blog post
dox generate --type blog --prompt "AI trends in healthcare" --output blog.md

# Business report with options
dox generate -t report -p "Q4 sales analysis" -o report.md --temperature 0.3

# Custom content
dox generate --type custom --prompt "Write a technical guide for Docker beginners" \
  --output guide.md --max-tokens 3000

# Non-English content
dox generate --type email --prompt "Schedule meeting" --language Korean --output email.md
```

### `dox config`

Manage dox configuration.

#### Synopsis
```bash
dox config [flags]
```

#### Flags
| Flag | Description | Example |
|------|-------------|---------|
| `--init` | Create default config | `dox config --init` |
| `--list, -l` | Show all settings | `dox config --list` |
| `--get <key>` | Get specific value | `dox config --get openai.model` |
| `--set <key=value>` | Set configuration | `dox config --set replace.backup=true` |
| `--unset <key>` | Remove setting | `dox config --unset custom.key` |
| `--edit` | Open in editor | `dox config --edit` |

#### Configuration Keys
```yaml
# Available configuration keys
openai:
  api_key: "sk-..."
  model: "gpt-3.5-turbo"
  max_tokens: 2000
  temperature: 0.7

defaults:
  path: "./documents"
  backup: true
  recursive: false

replace:
  backup: true
  concurrent: false
  max_workers: 4

performance:
  concurrent: true
  cache: true
  
ui:
  color: true
  progress: true
  verbose: false
```

#### Examples
```bash
# Initialize configuration
dox config --init

# Set API key
dox config --set openai.api_key="sk-your-key-here"

# View current settings
dox config --list

# Get specific value
dox config --get defaults.path

# Edit config file
dox config --edit
```

### `dox version`

Display version information.

#### Synopsis
```bash
dox version [flags]
```

#### Flags
| Flag | Description |
|------|-------------|
| `--short` | Show version only |
| `--json` | Output as JSON |

#### Examples
```bash
# Basic version
dox version
# Output: dox version 1.2534.0

# Short format
dox version --short
# Output: 1.2534.0

# JSON format
dox version --json
# Output: {"version":"1.2534.0","head":1,"yearweek":2534,"build":0}
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Misuse of command |
| 3 | File not found |
| 4 | Permission denied |
| 5 | Invalid configuration |
| 6 | API error |

## Environment Variables

| Variable | Description | Used By |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | generate |
| `DOX_CONFIG` | Config file path | all |
| `DOX_CACHE_DIR` | Cache directory | all |
| `NO_COLOR` | Disable colors | all |
| `DOX_DEBUG` | Debug mode | all |

## Performance Tips

### Concurrent Processing
Enable for bulk operations:
```bash
dox replace --rules rules.yml --path ./docs --concurrent --max-workers 8
```

### Caching
Templates and AI responses are cached:
```bash
dox config --set performance.cache=true
```

### Batch Operations
Process multiple files at once:
```bash
find . -name "*.md" | xargs -I {} dox create --from {} --output {}.docx
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Fix with sudo or check file permissions
sudo dox config --init
chmod 644 document.docx
```

#### API Key Not Found
```bash
# Set via environment or config
export OPENAI_API_KEY="sk-..."
# OR
dox config --set openai.api_key="sk-..."
```

#### Out of Memory
```bash
# Reduce concurrent workers
dox replace --rules rules.yml --path . --max-workers 2
```

### Debug Mode
```bash
# Enable debug output
export DOX_DEBUG=1
dox replace --rules rules.yml --path test.docx
```

### Getting Help
```bash
# Command help
dox <command> --help

# Visit GitHub
# https://github.com/pyhub/pyhub-docs/issues
```