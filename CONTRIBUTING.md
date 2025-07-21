# Contributing to Miko-Manifest

Thank you for your interest in contributing to Miko-Manifest! This document provides guidelines and information about contributing to the project.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected and actual behavior
- Your environment details (OS, Go version, etc.)
- Relevant configuration files and logs

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- A clear and descriptive title
- A detailed description of the proposed enhancement
- Use cases for the enhancement
- Examples of how it would work

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`make test`)
6. Run linting (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

#### Pull Request Guidelines

- Follow the existing code style
- Add tests for new functionality
- Update documentation if needed
- Ensure all CI checks pass
- Provide a clear description of the changes

## Development Setup

### Prerequisites

- Go 1.20 or later
- Make
- Docker (optional, for integration tests)

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/jepemo/miko-manifest.git
cd miko-manifest

# Install dependencies
go mod download

# Run tests
make test

# Build the project
make build

# Run linting
make lint
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific tests
go test -v ./pkg/mikomanifest/...

# Run integration tests
make integration-test
```

### Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Follow the project's existing patterns

### Project Structure

```
├── cmd/                    # CLI commands
├── pkg/mikomanifest/      # Core library
├── templates/             # Example templates
├── config/               # Example configurations
├── .github/              # GitHub workflows and templates
├── docs/                 # Documentation
└── test/                 # Integration tests
```

## Testing

### Unit Tests

- Write unit tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests when appropriate
- Mock external dependencies

### Integration Tests

- Test real-world scenarios
- Test CLI commands end-to-end
- Test with different configuration files

## Documentation

- Update README.md for new features
- Add inline code comments
- Update examples if needed
- Consider adding blog posts for major features

## Release Process

Releases are handled by maintainers:

1. Update version in relevant files
2. Create a new tag (`git tag -a v1.0.0 -m "Release v1.0.0"`)
3. Push the tag (`git push origin v1.0.0`)
4. GitHub Actions will automatically create a release

## Getting Help

- Check existing issues and discussions
- Ask questions in GitHub Discussions
- Join our community channels (if available)

## Recognition

Contributors are recognized in:

- GitHub contributors list
- Release notes
- Special mentions for significant contributions

Thank you for contributing to Miko-Manifest! 🎉
