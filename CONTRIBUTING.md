# Contributing to dox

Thank you for your interest in contributing to dox! We welcome contributions from the community and are grateful for any help you can provide.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## 📜 Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## 🚀 Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a branch** for your changes
4. **Make your changes** and commit them
5. **Push to your fork** and submit a pull request

## 💻 Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional but recommended)
- Your favorite code editor (we recommend VS Code with Go extension)

### Initial Setup

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/pyhub-documents-cli.git
cd pyhub-documents-cli

# Add upstream remote
git remote add upstream https://github.com/pyhub-kr/pyhub-documents-cli.git

# Install dependencies
go mod download

# Build the project
go build -o dox

# Run tests
go test ./...
```

### Development Tools

```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

## 🤝 How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

1. **Clear title and description**
2. **Steps to reproduce**
3. **Expected behavior**
4. **Actual behavior**
5. **System information** (OS, Go version, dox version)
6. **Relevant logs or error messages**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

1. **Clear title and description**
2. **Use case and motivation**
3. **Possible implementation approach**
4. **Alternative solutions considered**

### Code Contributions

1. **Find an issue** to work on or create a new one
2. **Comment on the issue** to let others know you're working on it
3. **Follow the development workflow** below

## 🔄 Development Workflow

### 1. Sync with Upstream

```bash
git checkout main
git fetch upstream
git rebase upstream/main
```

### 2. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 3. Make Your Changes

Follow these guidelines:
- Write clear, concise commit messages
- Keep commits focused and atomic
- Include tests for new features
- Update documentation as needed

### 4. Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or modifications
- `chore`: Maintenance tasks

Examples:
```bash
git commit -m "feat(replace): add support for Excel files"
git commit -m "fix(template): handle missing placeholders gracefully"
git commit -m "docs: update installation instructions"
```

### 5. Test Your Changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/replace

# Run tests with race detection
go test -race ./...
```

### 6. Update Documentation

- Update README.md if needed
- Update command help text
- Add examples if introducing new features
- Update API documentation for library changes

## 📝 Coding Standards

### Go Code Style

We follow the standard Go style guidelines:

1. **Format code** with `gofmt`
2. **Organize imports** with `goimports`
3. **Follow [Effective Go](https://golang.org/doc/effective_go.html)**
4. **Use meaningful variable names**
5. **Comment exported functions and types**

### Code Quality Checks

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Security scan
gosec ./...

# Vet code
go vet ./...
```

### Project-Specific Guidelines

1. **Error Handling**
   - Always check errors
   - Use custom error types in `internal/errors`
   - Provide context with error wrapping

2. **Internationalization**
   - All user-facing strings must support i18n
   - Add translations to `locales/` directory
   - Use `i18n.T()` for translatable strings

3. **Testing**
   - Write unit tests for all new functions
   - Maintain >80% code coverage
   - Use table-driven tests where appropriate
   - Mock external dependencies

4. **Performance**
   - Use concurrent processing where beneficial
   - Implement progress indicators for long operations
   - Profile code for performance bottlenecks

## 🧪 Testing

### Unit Tests

```go
// Example test structure
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "expected",
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Place integration tests in `*_integration_test.go` files and use build tags:

```go
//go:build integration
// +build integration

package package_test

func TestIntegration(t *testing.T) {
    // Integration test code
}
```

Run integration tests:
```bash
go test -tags=integration ./...
```

## 📚 Documentation

### Code Documentation

- Document all exported types, functions, and methods
- Use clear, concise comments
- Include examples in documentation where helpful

```go
// ProcessDocument processes a Word or PowerPoint document by applying
// the specified replacement rules. It returns the number of replacements
// made and any error encountered.
//
// Example:
//   count, err := ProcessDocument("document.docx", rules)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Made %d replacements\n", count)
func ProcessDocument(path string, rules []Rule) (int, error) {
    // Implementation
}
```

### User Documentation

- Update README.md for user-facing changes
- Add examples to `examples/` directory
- Update command help text in Cobra commands

## 📤 Submitting Changes

### Pull Request Process

1. **Update your branch** with the latest upstream changes
2. **Push your changes** to your fork
3. **Create a Pull Request** with:
   - Clear title and description
   - Reference to related issues
   - List of changes made
   - Screenshots (if UI changes)
   - Test results

### Pull Request Template

```markdown
## Description
Brief description of changes

## Related Issue
Fixes #(issue number)

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Checklist
- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Changelog updated (if needed)
```

### Review Process

1. Automated checks must pass
2. At least one maintainer review required
3. All feedback addressed
4. Branch up to date with main

## 🚢 Release Process

We use [HeadVer](https://github.com/line/headver) versioning:

```
{head}.{yearweek}.{build}
```

### Creating a Release

1. **Update version** in code
2. **Update CHANGELOG.md**
3. **Create release PR**
4. **Tag release** after merge
5. **Build binaries** for all platforms
6. **Create GitHub release** with binaries

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Binaries built for all platforms
- [ ] Release notes written
- [ ] GitHub release created

## 🎯 Areas for Contribution

### Good First Issues

Look for issues labeled `good first issue` - these are great for newcomers.

### Priority Areas

- **Excel Support**: Add support for .xlsx files
- **PDF Generation**: Implement PDF export functionality
- **Performance**: Optimize document processing speed
- **Testing**: Increase test coverage
- **Documentation**: Improve user guides and examples
- **Internationalization**: Add more language support

### Feature Ideas

- Cloud storage integration (S3, Google Drive)
- Web UI interface
- Plugin system
- More AI providers (Claude, Gemini)
- Document comparison features
- Batch processing improvements

## 📞 Getting Help

- **Discord**: Join our community (coming soon)
- **GitHub Discussions**: Ask questions and share ideas
- **Issue Tracker**: Report bugs and request features
- **Email**: support@pyhub.kr

## 🙏 Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to dox! Your efforts help make document automation better for everyone.

---

**Note**: This contributing guide is a living document. If you find something confusing or have suggestions for improvement, please let us know!