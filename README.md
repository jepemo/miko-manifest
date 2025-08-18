# Miko-Manifest

[![GitHub Release](https://img.shields.io/github/v/release/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/releases)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/jepemo/miko-manifest/ci.yml)](https://github.com/jepemo/miko-manifest/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jepemo/miko-manifest)](https://goreportcard.com/report/github.com/jepemo/miko-manifest)
[![GitHub License](https://img.shields.io/github/license/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/blob/main/LICENSE)
[![GitHub Issues](https://img.shields.io/github/issues/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/pulls)
[![GitHub Stars](https://img.shields.io/github/stars/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/stargazers)

Miko-Manifest is a CLI application written in Go that provides powerful configuration management for Kubernetes manifests with templating capabilities.

## Features

- **Template Processing**: Supports Go templates with three processing patterns:
  - Simple file processing
  - Same-file repeat (multiple sections in one file)
  - Multiple-files repeat (separate files per key)
- **Environment Configuration**: YAML-based environment-specific configurations
- **Variable Override**: Command-line variable overrides
- **Hierarchical Configuration**: Include and merge configurations with resources section
- **Integrated Schema Validation**: Schema definitions within environment configurations
- **Auto-Environment Detection**: Automatic environment detection from build artifacts
- **YAML Validation**: Native Go YAML validation and Kubernetes manifest validation
- **Library Architecture**: Core functionality available as a Go library

## Installation

### From Source

```bash
go install github.com/jepemo/miko-manifest@latest
```

### Using Docker

```bash
docker pull jepemo/miko-manifest:latest
```

## Usage

### Initialize a Project

```bash
miko-manifest init
```

This command creates:

- `templates/`: Directory for Go template files
- `config/`: Directory for environment-specific YAML configurations
- Example files demonstrating all three processing patterns

### Build Project

```bash
miko-manifest build --env dev --output-dir output
```

**Available Parameters:**

- `--env`, `-e`: Environment configuration to use (required)
- `--output-dir`, `-o`: Output directory for generated files (required)
- `--config`, `-c`: Configuration directory path (default: "config")
- `--templates`, `-t`: Templates directory path (default: "templates")
- `--var`: Override variables (format: `--var NAME=VALUE`)
- `--validate`: Perform validation after build (equivalent to build + lint)

**Examples:**

```bash
# Basic build
miko-manifest build --env dev --output-dir output

# With custom directories
miko-manifest build --env prod --output-dir dist --config prod-config --templates prod-templates

# With variable overrides
miko-manifest build --env dev --output-dir output --var app_name=my-app --var replicas=5

# Build and validate in one command
miko-manifest build --env dev --output-dir output --validate
```

### Display Configuration

```bash
miko-manifest config --env dev
```

Display configuration information for a specific environment with multiple viewing options:

**Available Parameters:**

- `--env`, `-e`: Environment configuration to use (required)
- `--config`, `-c`: Configuration directory path (default: "config")
- `--variables`: Show only variables in `var=value` format (one per line)
- `--schemas`: Show list of all configured schemas
- `--tree`: Show the hierarchy of included resources with detailed loading process

**Examples:**

```bash
# Show complete unified configuration
miko-manifest config --env dev

# Show only variables (useful for scripts)
miko-manifest config --env dev --variables

# Show configured schemas
miko-manifest config --env dev --schemas

# Show hierarchical resource loading tree
miko-manifest config --env dev --tree
```

### Validate Configuration

```bash
miko-manifest check --config config
```

Validates YAML files in the configuration directory using native Go YAML parsing.

### Lint Generated Files

```bash
miko-manifest lint --dir output
```

Performs two-step validation:

1. **YAML Linting**: Using native Go YAML parser
2. **Kubernetes Validation**: Schema validation for Kubernetes manifests

**Auto-Environment Detection:**

If you've previously built the project, the lint command automatically detects the environment:

```bash
# After building with --env dev
miko-manifest build --env dev --output-dir output

# Lint automatically detects dev environment and loads schemas
miko-manifest lint output
```

**Manual Environment Specification:**

```bash
# Explicitly specify environment
miko-manifest lint --env dev --dir output

# Skip schema validation for faster linting
miko-manifest lint --skip-schema-validation --dir output
```

**Integrated Schema Validation:**

Schemas are now defined directly in your environment configuration files:

```yaml
# config/dev.yaml
resources:
  - base.yaml

schemas:
  # Load from URL (e.g., Crossplane)
  - https://raw.githubusercontent.com/crossplane/crossplane/master/cluster/crds/apiextensions.crossplane.io_compositions.yaml
  # Load from local file
  - ./schemas/my-operator-crd.yaml
  # Load from directory (recursive)
  - ./schemas/operators/

variables:
  - name: environment
    value: development
```

## Command Reference

### Build Command Options

```bash
miko-manifest build [flags]
```

**Flags:**

- `--env`, `-e`: Environment configuration to use (required)
- `--output-dir`, `-o`: Output directory for generated files (required)
- `--config`, `-c`: Configuration directory path (default: "config")
- `--templates`, `-t`: Templates directory path (default: "templates")
- `--var`: Override variables (format: `--var NAME=VALUE`)
- `--validate`: Perform validation after build (equivalent to build + lint)
- `--debug-config`: Show the final merged configuration
- `--show-config-tree`: Show the hierarchy of included resources

### Lint Command Options

```bash
miko-manifest lint [directory] [flags]
```

**Flags:**

- `--dir`, `-d`: Directory to validate (can also be specified as positional argument)
- `--env`, `-e`: Environment to load schemas from (auto-detected if not specified)
- `--config`, `-c`: Configuration directory path (default: "config")
- `--skip-schema-validation`: Skip schema loading for faster YAML-only validation

### Check Command Options

```bash
miko-manifest check [flags]
```

**Flags:**

- `--config`, `-c`: Configuration directory path to validate (default: "config")

### Init Command Options

```bash
miko-manifest init [flags]
```

**Flags:**

- `--project-dir`: Directory to initialize project in (default: current directory)

## Template Processing Types

### 1. Simple File Processing

```yaml
include:
  - file: deployment.yaml
```

Template is processed once with global variables.

### 2. Same-File Repeat

```yaml
include:
  - file: configmap.yaml
    repeat: same-file
    list:
      - key: database-config
        values:
          - name: config_name
            value: database-config
          - name: database_url
            value: postgresql://localhost:5432/mydb
      - key: cache-config
        values:
          - name: config_name
            value: cache-config
          - name: database_url
            value: redis://localhost:6379
```

Generates multiple sections in a single file separated by `---`.

### 3. Multiple-Files Repeat

```yaml
include:
  - file: service.yaml
    repeat: multiple-files
    list:
      - key: frontend
        values:
          - name: service_name
            value: frontend-service
          - name: service_port
            value: "80"
      - key: backend
        values:
          - name: service_name
            value: backend-service
          - name: service_port
            value: "3000"
```

Generates separate files: `service-frontend.yaml`, `service-backend.yaml`.

## Configuration Structure

### Hierarchical Configuration

Miko-Manifest supports hierarchical configuration through the `resources` section, allowing you to create modular, reusable configurations.

#### Basic Structure

```yaml
# config/dev.yaml
resources:
  - base.yaml # Include base configuration
  - components/ # Include all YAML files from directory

variables:
  - name: environment
    value: development

include:
  - file: deployment.yaml
```

#### Configuration Merging Rules

1. **Load Order**: Resources are processed in the order they appear
2. **Variable Precedence**: Later definitions override earlier ones
3. **Include Combination**: All includes are merged (no duplicates)
4. **Directory Processing**: Files in directories are loaded alphabetically

#### Example: Environment Inheritance

**Base Configuration (`config/base.yaml`)**:

```yaml
variables:
  - name: app_name
    value: my-app
  - name: replicas
    value: "1"
  - name: image
    value: nginx

include:
  - file: deployment.yaml
  - file: service.yaml
```

**Component Configuration (`config/components/database.yaml`)**:

```yaml
variables:
  - name: database_host
    value: localhost
  - name: database_port
    value: "5432"

include:
  - file: configmap.yaml
    repeat: same-file
    list:
      - key: database-config
        values:
          - name: config_name
            value: database-config
```

**Development Configuration (`config/dev.yaml`)**:

```yaml
resources:
  - base.yaml
  - components/

schemas:
  # CRD schemas for validation
  - ./schemas/my-operator-crd.yaml
  - https://raw.githubusercontent.com/example/operator/main/crd.yaml

variables:
  - name: replicas
    value: "1" # Override for development
  - name: environment
    value: development

include:
  - file: service.yaml # Additional service for dev
    repeat: multiple-files
    list:
      - key: debug
        values:
          - name: service_name
            value: debug-service
```

**Result**: The final configuration combines all variables and includes, with development-specific overrides taking precedence.

#### Debugging Configuration

Use debug flags to understand configuration merging:

```bash
# Show configuration hierarchy
miko-manifest build --env dev --output-dir output --show-config-tree

# Show final merged configuration
miko-manifest build --env dev --output-dir output --debug-config
```

#### Safety Features

- **Circular Dependency Detection**: Prevents infinite loops
- **Maximum Depth Limit**: Configurable recursion depth (default: 5)
- **Path Resolution**: Relative paths are resolved correctly
- **Clear Error Messages**: Descriptive errors for configuration issues

### Schema Integration Features

Miko-Manifest now supports integrated schema validation directly within configuration files:

#### Schema Definition in Configuration

Define schemas alongside your environment configuration:

```yaml
# config/dev.yaml
resources:
  - base.yaml

schemas:
  # Remote CRD from operator
  - https://raw.githubusercontent.com/crossplane/crossplane/master/cluster/crds/apiextensions.crossplane.io_compositions.yaml
  # Local CRD file
  - ./schemas/database-operator-crd.yaml
  # Directory of CRDs (recursive)
  - ./schemas/operators/

variables:
  - name: environment
    value: development
```

#### Schema Merging and Inheritance

Schemas follow the same hierarchical merging rules as other configuration:

```yaml
# config/base.yaml
schemas:
  - https://raw.githubusercontent.com/kubernetes/api/master/core/v1/configmap.yaml

# config/dev.yaml
resources:
  - base.yaml

schemas:
  - ./schemas/dev-specific-crds/ # Adds to base schemas, no duplicates
```

#### Auto-Detection Workflow

1. **Build Phase**: Environment information is saved to `.miko-manifest-env`
2. **Lint Phase**: Automatically detects environment and loads corresponding schemas
3. **Seamless Validation**: No need to specify schemas separately for linting

```bash
# Step 1: Build saves environment info
miko-manifest build --env dev --output-dir output

# Step 2: Lint auto-detects 'dev' and loads schemas from config/dev.yaml
miko-manifest lint output
```

#### Performance Options

- `--skip-schema-validation`: Skip schema loading for faster YAML-only validation
- Auto-caching of remote schemas for improved performance
- Efficient schema deduplication in hierarchical configurations

### Environment Configuration (`config/dev.yaml`)

```yaml
variables:
  - name: app_name
    value: my-app
  - name: namespace
    value: default
  - name: replicas
    value: "3"

include:
  - file: deployment.yaml
  - file: configmap.yaml
    repeat: same-file
    list:
      - key: database-config
        values:
          - name: config_name
            value: database-config
  - file: service.yaml
    repeat: multiple-files
    list:
      - key: frontend
        values:
          - name: service_name
            value: frontend-service
```

### Template Example (`templates/deployment.yaml`)

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{.app_name}}"
  namespace: "{{.namespace}}"
spec:
  replicas: { { .replicas } }
  selector:
    matchLabels:
      app: "{{.app_name}}"
  template:
    metadata:
      labels:
        app: "{{.app_name}}"
    spec:
      containers:
        - name: "{{.app_name}}"
          image: "{{.image}}:{{.tag}}"
          ports:
            - containerPort: { { .port } }
```

## Docker Usage

### Using Pre-built Image

```bash
# Check configuration
docker run --rm -v "$(pwd):/workspace" jepemo/miko-manifest:latest check --config /workspace/config

# Build manifests
docker run --rm -v "$(pwd):/workspace" jepemo/miko-manifest:latest build \
  --env dev \
  --output-dir /workspace/output \
  --config /workspace/config \
  --templates /workspace/templates

# Lint generated files
docker run --rm -v "$(pwd):/workspace" jepemo/miko-manifest:latest lint --dir /workspace/output
```

### CI/CD Pipeline Examples

**GitHub Actions:**

```yaml
name: Miko-Manifest Build
on: [push, pull_request]

env:
  MIKO_IMAGE: jepemo/miko-manifest:latest

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Validate configuration
        run: docker run --rm -v "${{ github.workspace }}:/workspace" $MIKO_IMAGE check --config /workspace/config

      - name: Generate manifests
        run: |
          docker run --rm \
            -v "${{ github.workspace }}:/workspace" \
            $MIKO_IMAGE build \
            --env dev \
            --output-dir /workspace/output \
            --config /workspace/config \
            --templates /workspace/templates

      - name: Validate generated manifests
        run: docker run --rm -v "${{ github.workspace }}:/workspace" $MIKO_IMAGE lint --dir /workspace/output
```

**GitLab CI:**

```yaml
stages:
  - validate
  - build
  - verify

validate-config:
  stage: validate
  image: jepemo/miko-manifest:latest
  script:
    - miko-manifest check --config config/

generate-manifests:
  stage: build
  image: jepemo/miko-manifest:latest
  script:
    - miko-manifest build --env ${ENVIRONMENT:-dev} --output-dir output/ --config config/ --templates templates/
  artifacts:
    paths:
      - output/
    expire_in: 1 week

verify-manifests:
  stage: verify
  image: jepemo/miko-manifest:latest
  script:
    - miko-manifest lint --dir output/
```

## Library Usage

Miko-Manifest is designed with a library-first approach. You can use it programmatically:

```go
package main

import (
    "github.com/jepemo/miko-manifest/pkg/mikomanifest"
)

func main() {
    // Initialize a project
    initOptions := mikomanifest.InitOptions{
        ProjectDir: "/path/to/project",
    }
    mikomanifest.InitProject(initOptions)

    // Build project
    buildOptions := mikomanifest.BuildOptions{
        Environment:   "dev",
        OutputDir:     "output",
        ConfigDir:     "config",
        TemplatesDir:  "templates",
        Variables:     map[string]string{"app_name": "my-app"},
    }

    mikoManifest := mikomanifest.New(buildOptions)
    mikoManifest.Build()

    // Lint directory
    lintOptions := mikomanifest.LintOptions{
        Directory: "output",
    }
    mikomanifest.LintDirectory(lintOptions)
}
```

## Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/jepemo/miko-manifest.git
cd miko-manifest

# Build
go build -o miko-manifest .

# Run
./miko-manifest --help
```

### Run Tests

```bash
go test ./...
```

### Build Docker Image

```bash
docker build -f Dockerfile.go -t miko-manifest:latest .
```

## Migration from Konfig

If you're migrating from the Python version (Konfig), the command structure is very similar:

**Python (Konfig):**

```bash
uv run konfig build --env dev --output-dir output
```

**Go (Miko-Manifest):**

```bash
miko-manifest build --env dev --output-dir output
```

## Dependencies

- **CLI Framework**: [cobra](https://github.com/spf13/cobra)
- **YAML Processing**: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- **Kubernetes Validation**: [k8s.io/client-go](https://github.com/kubernetes/client-go)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License
