# Miko-Shell Integration for Miko-Manifest

This document describes how to use `miko-shell` as an alternative to the traditional Makefile-based development workflow.

## 🚀 Quick Start

### 1. Install miko-shell

```bash
# Install via script (recommended)
curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash

# Or install a specific version
curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash -s -- --version v1.0.0
```

### 2. Build the development environment

```bash
# Build the containerized development environment
miko-shell image build

# Or let it auto-build on first run
miko-shell run test
```

### 3. Available Commands

```bash
# List all available scripts
miko-shell run

# Run tests
miko-shell run test
miko-shell run test-coverage
miko-shell run test-race

# Build the project
miko-shell run build           # Build with default version
miko-shell run build v1.2.3    # Build with specific version
miko-shell run build-dev       # Build development version

# Code quality
miko-shell run lint
miko-shell run fmt
miko-shell run precommit       # Run all pre-commit checks

# CI/CD commands
miko-shell run ci-test         # Run tests for CI
miko-shell run ci-build        # Build for CI

# Development
miko-shell run dev-setup       # Set up development environment
miko-shell run clean          # Clean build artifacts
miko-shell run clean-all      # Clean everything

# Examples
miko-shell run example-init    # Initialize example project
miko-shell run example-build   # Build example

# Get help
miko-shell run help           # Show all available commands
```

### 4. Interactive Development

```bash
# Open an interactive shell in the development environment
miko-shell open

# Inside the container you have access to:
# - Go 1.24
# - All project dependencies
# - golangci-lint, staticcheck
# - make (for compatibility)
```

## 🔧 Development Workflow

### Traditional Makefile vs Miko-Shell

| Task | Makefile | Miko-Shell |
|------|----------|------------|
| Run tests | `make test` | `miko-shell run test` |
| Build | `make build` | `miko-shell run build` |
| Lint | `make lint` | `miko-shell run lint` |
| Pre-commit | `make precommit` | `miko-shell run precommit` |
| Coverage | `make test-coverage` | `miko-shell run test-coverage` |
| Clean | `make clean` | `miko-shell run clean` |

### Advantages of Miko-Shell

1. **Reproducible Environment**: Same Go version, dependencies, and tools across all machines
2. **No Local Dependencies**: Only Docker/Podman required on host
3. **Isolation**: Development environment doesn't affect host system
4. **Cross-platform**: Works identically on Linux, macOS, and Windows
5. **Version Control**: Development environment configuration is versioned in `miko-shell.yaml`

### Host Architecture Support

The build scripts automatically detect and use host architecture variables:

- `MIKO_HOST_OS`: Host operating system (linux, darwin, windows)
- `MIKO_HOST_ARCH`: Host architecture (amd64, arm64)

This ensures binaries are built for the correct target platform.

## 📋 Configuration

The `miko-shell.yaml` configuration includes:

- **Base Image**: `golang:1.24-alpine`
- **Tools**: git, make, curl, bash, golangci-lint, staticcheck
- **Scripts**: All Makefile targets plus additional convenience commands
- **Environment**: Proper Go module support, build flags, and host architecture detection

## 🚀 CI/CD Integration

### GitHub Actions

A dedicated workflow (`.github/workflows/miko-shell-ci.yml`) tests the miko-shell integration:

- Installs miko-shell automatically
- Builds the development environment
- Runs tests and builds in parallel with traditional Makefile
- Ensures compatibility between both approaches

### Migration Strategy

1. **Phase 1** (Current): Both Makefile and miko-shell available
2. **Phase 2**: Gradually prefer miko-shell in documentation
3. **Phase 3**: Eventually deprecate Makefile (future)

## 🛠️ Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `miko-shell.yaml not found` | Run from project root directory |
| Docker/Podman not found | Install Docker or Podman |
| Permission denied | Add user to docker group or use sudo |
| Build fails | Check `miko-shell image build --force` |

### Debugging

```bash
# Check miko-shell version
miko-shell version

# Rebuild environment
miko-shell image build --force

# Clean all images
miko-shell image clean --all

# Interactive debugging
miko-shell open
```

## 📚 Further Reading

- [Miko-Shell Documentation](https://github.com/jepemo/miko-shell/blob/main/DOCS.md)
- [Miko-Shell Examples](https://github.com/jepemo/miko-shell/tree/main/examples)
- [Project README](../README.md)

## 🤝 Contributing

When contributing to this project, you can use either approach:

1. **Traditional**: Use existing Makefile commands
2. **Modern**: Use miko-shell commands for reproducible development

Both approaches are supported and tested in CI.
