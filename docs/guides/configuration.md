# Configuration Guide

Learn how to configure dox for optimal performance and customize it to your workflow.

## 📁 Configuration File

### Location

dox looks for configuration in the following order:
1. File specified with `--config` flag
2. `.dox.yml` in current directory
3. `$HOME/.config/dox/config.yml` (user config)
4. `/etc/dox/config.yml` (system config)

### File Format

Configuration uses YAML format:
```yaml
# ~/.config/dox/config.yml
openai:
  api_key: "sk-your-api-key"
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
  cache_dir: "~/.cache/dox"
  
ui:
  color: true
  progress: true
  verbose: false
  
logging:
  enabled: false
  level: "info"
  path: "~/.local/share/dox/logs"
```

## 🚀 Quick Setup

### Initialize Configuration
```bash
# Create default config file
dox config --init

# Location: ~/.config/dox/config.yml
Configuration file created successfully!
```

### Essential Settings
```bash
# Set OpenAI API key
dox config --set openai.api_key="sk-your-key"

# Set default document directory
dox config --set defaults.path="/path/to/documents"

# Enable backups by default
dox config --set replace.backup=true

# Enable colored output
dox config --set ui.color=true
```

## ⚙️ Configuration Sections

### OpenAI Settings
Configure AI generation features:
```yaml
openai:
  api_key: "sk-..."           # Your OpenAI API key
  model: "gpt-3.5-turbo"      # Default model
  max_tokens: 2000            # Maximum response length
  temperature: 0.7            # Creativity level (0.0-2.0)
  timeout: 30                 # Request timeout in seconds
  retry_count: 3              # Number of retries on failure
  organization: ""            # Optional: Organization ID
```

### Default Behaviors
Set default command options:
```yaml
defaults:
  path: "./documents"         # Default directory for operations
  backup: true               # Always create backups
  recursive: false           # Process subdirectories
  dry_run: false            # Preview mode
  output_dir: "./output"     # Default output directory
  template_dir: "./templates" # Template directory
```

### Replace Command
Customize text replacement behavior:
```yaml
replace:
  backup: true               # Create .bak files
  backup_suffix: ".bak"      # Backup file extension
  concurrent: false          # Parallel processing
  max_workers: 4            # Concurrent workers
  include: "*.docx,*.pptx"   # File patterns to include
  exclude: "~$*,*.tmp"       # File patterns to exclude
  case_sensitive: true       # Case-sensitive search
```

### Performance Tuning
Optimize for speed and efficiency:
```yaml
performance:
  concurrent: true           # Enable parallel processing
  max_workers: 8            # Maximum parallel workers
  cache: true               # Enable caching
  cache_dir: "~/.cache/dox" # Cache directory
  cache_ttl: 3600           # Cache TTL in seconds
  buffer_size: 8192         # I/O buffer size
  compression: true         # Compress cached data
```

### User Interface
Customize output appearance:
```yaml
ui:
  color: true               # Colored output
  progress: true            # Progress bars
  verbose: false            # Detailed output
  quiet: false             # Suppress non-error output
  format: "pretty"          # Output format (pretty/json/plain)
  spinner: "dots"           # Progress spinner style
  timestamp: false          # Show timestamps
```

### Logging
Configure logging behavior:
```yaml
logging:
  enabled: true                      # Enable logging
  level: "info"                      # Log level (debug/info/warn/error)
  path: "~/.local/share/dox/logs"   # Log directory
  max_size: "10MB"                   # Max log file size
  max_age: 30                        # Days to keep logs
  format: "json"                     # Log format (text/json)
```

### Templates
Template processing settings:
```yaml
templates:
  dir: "./templates"         # Template directory
  cache: true               # Cache parsed templates
  strict: false             # Strict variable checking
  missing_var: "error"      # How to handle missing vars (error/warn/ignore)
  delimiters:
    left: "{{"              # Left delimiter
    right: "}}"             # Right delimiter
```

### Security
Security-related settings:
```yaml
security:
  verify_ssl: true          # Verify SSL certificates
  api_key_env: "OPENAI_API_KEY" # Environment variable for API key
  mask_secrets: true        # Mask secrets in logs
  allowed_paths:           # Restrict file access
    - "/home/user/documents"
    - "/var/shared/templates"
```

## 🔧 Management Commands

### View Configuration
```bash
# Show all settings
dox config --list

# Get specific value
dox config --get openai.model
# Output: gpt-3.5-turbo

# Get nested value
dox config --get performance.cache
# Output: true
```

### Modify Configuration
```bash
# Set single value
dox config --set ui.color=false

# Set nested value
dox config --set openai.temperature=0.5

# Unset (remove) value
dox config --unset logging.enabled
```

### Edit Configuration
```bash
# Open in default editor
dox config --edit

# Open in specific editor
EDITOR=vim dox config --edit
```

### Validate Configuration
```bash
# Check for errors
dox config --validate
# ✓ Configuration is valid

# Show effective configuration (merged from all sources)
dox config --show-effective
```

## 🌍 Environment Variables

Override configuration with environment variables:

| Environment Variable | Configuration Key | Description |
|---------------------|------------------|-------------|
| `OPENAI_API_KEY` | `openai.api_key` | OpenAI API key |
| `DOX_CONFIG` | - | Config file path |
| `DOX_CACHE_DIR` | `performance.cache_dir` | Cache directory |
| `DOX_DEBUG` | `logging.level` | Set to debug mode |
| `NO_COLOR` | `ui.color` | Disable colors |
| `DOX_WORKERS` | `performance.max_workers` | Worker count |

### Priority Order
1. Command-line flags (highest)
2. Environment variables
3. Config file in current directory
4. User config file
5. System config file (lowest)

## 📝 Configuration Profiles

Create different configurations for different scenarios:

### Development Profile
```yaml
# ~/.config/dox/profiles/dev.yml
defaults:
  backup: true
  dry_run: true
  
ui:
  verbose: true
  
logging:
  enabled: true
  level: debug
```

### Production Profile
```yaml
# ~/.config/dox/profiles/prod.yml
defaults:
  backup: true
  dry_run: false
  
performance:
  concurrent: true
  max_workers: 16
  
ui:
  verbose: false
  quiet: true
```

### Using Profiles
```bash
# Use specific profile
dox --config ~/.config/dox/profiles/dev.yml replace ...

# Set alias for convenience
alias dox-dev='dox --config ~/.config/dox/profiles/dev.yml'
alias dox-prod='dox --config ~/.config/dox/profiles/prod.yml'
```

## 🎯 Common Configurations

### Minimal Setup
Basic configuration for getting started:
```yaml
openai:
  api_key: "sk-your-key"

defaults:
  backup: true
```

### Power User
Advanced configuration with all features:
```yaml
openai:
  api_key: "sk-your-key"
  model: "gpt-4"
  max_tokens: 3000

defaults:
  path: "~/Documents"
  backup: true
  recursive: true

performance:
  concurrent: true
  max_workers: 8
  cache: true

ui:
  color: true
  progress: true

logging:
  enabled: true
  level: "info"
```

### CI/CD Environment
Configuration for automated environments:
```yaml
defaults:
  backup: false
  dry_run: false

performance:
  concurrent: true
  max_workers: 4

ui:
  color: false
  progress: false
  quiet: true
  format: "json"

logging:
  enabled: true
  format: "json"
```

## 🔐 Security Best Practices

### Protecting API Keys
```bash
# Never store API keys in config files committed to git
# Use environment variables instead
export OPENAI_API_KEY="sk-..."

# Or use a secrets manager
dox config --set openai.api_key="$(vault read -field=key secret/openai)"
```

### File Permissions
```bash
# Restrict config file access
chmod 600 ~/.config/dox/config.yml

# Verify permissions
ls -la ~/.config/dox/config.yml
# -rw------- 1 user user 1234 Jan 15 10:00 config.yml
```

### Path Restrictions
```yaml
# Limit file operations to specific directories
security:
  allowed_paths:
    - "/home/user/safe-documents"
    - "/var/shared/templates"
  blocked_paths:
    - "/etc"
    - "/sys"
    - "/root"
```

## 🐛 Troubleshooting

### Configuration Not Loading
```bash
# Check which config file is being used
dox config --show-source

# Validate syntax
dox config --validate

# Show parse errors
dox config --debug
```

### Precedence Issues
```bash
# Show effective configuration (after all merging)
dox config --show-effective

# Test with specific config
dox --config test.yml config --list
```

### Reset Configuration
```bash
# Backup current config
cp ~/.config/dox/config.yml ~/.config/dox/config.yml.backup

# Reset to defaults
dox config --reset

# Or manually delete
rm ~/.config/dox/config.yml
dox config --init
```

## 📊 Performance Optimization

### Concurrent Processing
```yaml
# Enable for bulk operations
performance:
  concurrent: true
  max_workers: 8  # Adjust based on CPU cores
```

### Caching Strategy
```yaml
# Configure caching for better performance
performance:
  cache: true
  cache_dir: "/fast-ssd/dox-cache"  # Use fast storage
  cache_ttl: 7200  # 2 hours
  compression: true  # Save space
```

### Memory Management
```yaml
# Optimize for large files
performance:
  buffer_size: 16384  # Larger buffer for big files
  max_file_size: "100MB"  # Limit for safety
```

## 🎨 Customization Examples

### Custom Delimiters
```yaml
# Use different template delimiters
templates:
  delimiters:
    left: "<%"
    right: "%>"
```

### Logging Customization
```yaml
# Detailed logging setup
logging:
  enabled: true
  level: "debug"
  path: "/var/log/dox"
  format: "json"
  outputs:
    - file: "dox.log"
    - file: "errors.log"
      level: "error"
```

### UI Themes
```yaml
# Custom UI settings
ui:
  color: true
  theme: "dark"  # future feature
  progress_style: "bar"  # bar/spinner/percentage
  success_symbol: "✓"
  error_symbol: "✗"
```

## 📚 Advanced Topics

### Dynamic Configuration
Load configuration based on conditions:
```bash
# Use different config based on environment
if [ "$ENV" = "production" ]; then
  export DOX_CONFIG="/etc/dox/prod.yml"
else
  export DOX_CONFIG="~/.config/dox/dev.yml"
fi
```

### Configuration Templating
Generate configuration from templates:
```bash
# Create config template
cat > config.template.yml << EOF
openai:
  api_key: "${OPENAI_KEY}"
  model: "${AI_MODEL:-gpt-3.5-turbo}"
EOF

# Generate actual config
envsubst < config.template.yml > ~/.config/dox/config.yml
```

### Distributed Configuration
Share configuration across team:
```bash
# Store in git (without secrets)
git clone https://github.com/team/dox-configs
ln -s $(pwd)/dox-configs/team.yml ~/.config/dox/config.yml

# Load secrets from secure source
dox config --set openai.api_key="$(security find-generic-password -w -s dox-api)"
```

## 🔗 Related Documentation

- [Getting Started](./getting-started.md) - Initial setup
- [Command Reference](./commands.md) - All commands
- [Environment Variables](#environment-variables) - Env var reference