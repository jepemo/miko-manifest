# Miko-Manifest

Miko-Manifest is a CLI application written in Go that provides powerful configuration management for Kubernetes manifests with templating capabilities.

## Features

- **Template Processing**: Supports Go templates with three processing patterns:
  - Simple file processing
  - Same-file repeat (multiple sections in one file)
  - Multiple-files repeat (separate files per key)
- **Environment Configuration**: YAML-based environment-specific configurations
- **Variable Override**: Command-line variable overrides
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
- `--debug-config`: Show the final merged configuration
- `--show-config-tree`: Show the hierarchy of included resources

**Examples:**

```bash
# Basic build
miko-manifest build --env dev --output-dir output

# With hierarchical configuration debugging
miko-manifest build --env dev --output-dir output --show-config-tree --debug-config

# With custom directories
miko-manifest build --env prod --output-dir dist --config prod-config --templates prod-templates

# With variable overrides
miko-manifest build --env dev --output-dir output --var app_name=my-app --var replicas=5
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

**Custom Resource Validation:**

```bash
miko-manifest lint --dir output --schema-config config/schemas.yaml
```

Extended validation with support for Custom Resource Definitions (CRDs):

- **Custom Schema Loading**: Load CRDs from URLs, files, or directories
- **Multi-source Support**: Combine schemas from different sources
- **Automatic Discovery**: Infer GVK information from CRD definitions

### Custom Schema Configuration

Create a `schemas.yaml` file to define custom resource schemas:

```yaml
schemas:
  # Load from URL (e.g., Crossplane)
  - https://raw.githubusercontent.com/crossplane/crossplane/master/cluster/crds/apiextensions.crossplane.io_compositions.yaml

  # Load from local file
  - ./schemas/my-operator-crd.yaml

  # Load from directory (recursive)
  - ./schemas/operators/
```

The tool automatically:

- Detects source type (URL, file, or directory)
- Downloads and caches remote schemas
- Extracts GVK information from CRDs
- Validates custom resources against their schemas

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
  - base.yaml          # Include base configuration
  - components/        # Include all YAML files from directory

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

variables:
  - name: replicas
    value: "1"        # Override for development
  - name: environment
    value: development

include:
  - file: service.yaml  # Additional service for dev
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
