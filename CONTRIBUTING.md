# Contributing to pyhub-docs

Thank you for your interest in contributing to pyhub-docs! This document provides guidelines and instructions for contributing.

## 🎯 Development Philosophy

We follow **Test-Driven Development (TDD)**:
1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve the code while keeping tests green

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional but recommended)

### Setting Up Development Environment

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/pyhub-docs.git
cd pyhub-docs

# Install dependencies
go mod download

# Run tests to verify setup
make test
```

## 📋 Development Process

### 1. Check or Create an Issue

Before starting work:
- Check existing [issues](https://github.com/pyhub/pyhub-docs/issues)
- If none exists, create a new issue describing what you want to work on
- Wait for maintainer feedback/approval for significant changes

### 2. Create a Feature Branch

```bash
git checkout -b feature/#ISSUE_NUMBER-brief-description
# Example: feature/#5-yaml-parser
```

### 3. Write Tests First (TDD)

Create test file before implementation:

```go
// internal/replace/parser_test.go
func TestParseYAMLRules(t *testing.T) {
    // Write your test cases here
    tests := []struct {
        name    string
        input   string
        want    []Rule
        wantErr bool
    }{
        // Test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 4. Implement the Feature

Write the minimal code to make tests pass:

```go
// internal/replace/parser.go
func ParseYAMLRules(data []byte) ([]Rule, error) {
    // Implementation
}
```

### 5. Run Tests and Checks

```bash
# Run tests
make test

# Check coverage
make coverage

# Format code
make fmt

# Run linter
make lint

# Run all checks
make ci
```

### 6. Commit Your Changes

Follow conventional commit format:

```bash
git add .
git commit -m "feat: add YAML rules parser for replace command

- Implement ParseYAMLRules function
- Add comprehensive test cases
- Handle edge cases

Closes #5"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `chore`: Build/tooling changes

### 7. Push and Create Pull Request

```bash
git push origin feature/#5-yaml-parser
```

Then create a PR on GitHub with:
- Clear title and description
- Reference to the issue
- Test results/coverage
- Screenshots if applicable

## 🧪 Testing Guidelines

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := setupTestData()
    
    // Act
    result, err := FunctionToTest(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Test Categories

1. **Unit Tests**: Test individual functions
2. **Integration Tests**: Test component interactions
3. **E2E Tests**: Test complete workflows

### Test Coverage

- Minimum 80% coverage for new code
- 100% coverage for critical paths
- Use `testdata/` directory for test fixtures

## 📝 Code Style

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

### Example Code Style

```go
// ParseRules parses replacement rules from YAML data.
// It returns an error if the YAML is invalid or contains
// unsupported rule formats.
func ParseRules(data []byte) ([]Rule, error) {
    var rules []Rule
    
    if err := yaml.Unmarshal(data, &rules); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    
    // Validate rules
    for i, rule := range rules {
        if err := rule.Validate(); err != nil {
            return nil, fmt.Errorf("invalid rule at index %d: %w", i, err)
        }
    }
    
    return rules, nil
}
```

## 📚 Documentation

### Code Documentation

- Document all exported types and functions
- Use clear, concise GoDoc comments
- Include examples where helpful

```go
// Rule represents a text replacement rule.
// It defines what text to find (Old) and what to replace it with (New).
type Rule struct {
    Old string `yaml:"old"`
    New string `yaml:"new"`
}
```

### User Documentation

Update relevant documentation:
- `README.md` for features/usage
- `docs/` for detailed guides
- `CHANGELOG.md` for notable changes

## 🔍 Code Review Process

PRs will be reviewed for:
1. **Functionality**: Does it work as intended?
2. **Tests**: Are there comprehensive tests?
3. **Code Quality**: Is it clean and maintainable?
4. **Documentation**: Is it well documented?
5. **Performance**: Are there any performance concerns?

## 🐛 Reporting Issues

When reporting issues, include:
- Clear description
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version)
- Error messages/logs

## 💡 Suggesting Features

Feature suggestions should include:
- Use case description
- Proposed solution
- Alternative solutions considered
- Potential impact on existing features

## 📊 Project Structure

```
pyhub-docs/
├── cmd/            # CLI commands
│   ├── root.go
│   ├── replace.go
│   ├── create.go
│   └── generate.go
├── internal/       # Internal packages
│   ├── docx/       # Word document processing
│   ├── pptx/       # PowerPoint processing
│   ├── markdown/   # Markdown parsing
│   └── replace/    # Replacement logic
├── pkg/            # Public packages
│   └── documents/  # Document API
├── tests/          # Test files
│   └── testdata/   # Test fixtures
└── docs/           # Documentation
```

## 🎉 Recognition

Contributors will be:
- Listed in release notes
- Added to CONTRIBUTORS.md
- Mentioned in relevant documentation

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License.

## 🙋 Getting Help

- Create an issue for bugs
- Use discussions for questions
- Check existing issues/PRs first

Thank you for contributing to pyhub-docs! 🚀