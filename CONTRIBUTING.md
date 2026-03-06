# Contributing to iKuai SDK

Thank you for your interest in contributing to the iKuai SDK! This document provides guidelines and instructions for contributing.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

This project adheres to the Contributor Covenant Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed and expected**
- **Include your environment details** (OS, Go version, iKuai version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a detailed description of the proposed feature**
- **Explain why this enhancement would be useful**
- **List some examples of how it would be used**

### Pull Requests

- Fill in the required template
- Do not include issue numbers in the PR title
- Include screenshots and animated GIFs in your pull request whenever possible
- Follow the coding standards
- Include tests for new features
- Update documentation for any changes

## Development Setup

### Prerequisites

- Go 1.19 or higher
- Git
- Access to an iKuai router (for integration testing)

### Setup Steps

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ikuai-api.git
   cd ikuai-api
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. Make your changes and test them:
   ```bash
   go test ./...
   ```

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` to format your code
- Use `goimports` to manage imports
- Run `go vet` and `golint` before committing

### Code Organization

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use meaningful package and variable names
- Add godoc comments for all exported functions and types

### Error Handling

- Always check and handle errors
- Use meaningful error messages
- Wrap errors with context when appropriate:
  ```go
  if err != nil {
      return fmt.Errorf("failed to do X: %w", err)
  }
  ```

### Interface Design

- Define interfaces in the `service/interface.go` file
- Keep interfaces small and focused
- Use descriptive names that reflect the purpose

### Example Code Pattern

```go
// Package example demonstrates SDK usage
package example

import (
    "context"
    "fmt"
    
    ikuaisdk "github.com/zy84338719/ikuai-api"
    "github.com/zy84338719/ikuai-api/service"
)

// ExampleFunction demonstrates how to use the SDK
func ExampleFunction() {
    ctx := context.Background()
    
    // Create client
    client, err := ikuaisdk.NewClientWithLogin(
        "192.168.1.1",
        "admin",
        "password",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Use the service
    api := service.NewAPIClient(client)
    devices, err := api.Monitor().GetLanIP(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d devices\n", len(devices))
}
```

## Testing

### Unit Tests

- Write unit tests for all new functions and methods
- Use table-driven tests for multiple test cases
- Aim for at least 80% code coverage
- Place tests in the same package with `_test.go` suffix

Example:
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Use build tag `//go:build integration` for integration tests
- Read credentials from environment variables
- Never hardcode credentials in tests

Example:
```go
//go:build integration

func TestIntegrationFeature(t *testing.T) {
    addr := os.Getenv("IKUAI_TEST_ADDR")
    username := os.Getenv("IKUAI_TEST_USERNAME")
    password := os.Getenv("IKUAI_TEST_PASSWORD")
    
    if password == "" {
        t.Skip("IKUAI_TEST_PASSWORD not set")
    }
    
    // ... test code
}
```

### Running Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Run specific package tests
go test ./service -v
```

## Pull Request Process

1. **Update Documentation**: Update the README.md and CHANGELOG.md with details of changes

2. **Add Tests**: Ensure all new code has corresponding tests

3. **Update Types**: If adding new API methods, add corresponding type definitions in `types/`

4. **Follow Conventions**: Match the existing code style and structure

5. **Test Thoroughly**: Run all tests and ensure they pass:
   ```bash
   go test ./...
   go vet ./...
   ```

6. **Commit Messages**: Write clear commit messages:
   - Use the present tense ("Add feature" not "Added feature")
   - Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
   - Limit the first line to 72 characters or less
   - Reference issues and pull requests liberally after the first line

7. **Create Pull Request**: Fill in the PR template completely

### PR Checklist

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No hardcoded credentials or sensitive data
- [ ] Code follows project style guidelines

## Questions?

Feel free to open an issue for any questions or concerns. We're here to help!

Thank you for contributing! 🎉
