# Miko-Manifest Documentation

> Comprehensive user & operator guide for the `miko-manifest` CLI and library.
>
> If you are just getting started, read the Quick Start section first. This document goes deeper than the README
> and is intended as the canonical, exhaustive reference.

---

## 1. Introduction

`miko-manifest` is a lightweight yet expressive CLI (and Go library) for generating and validating Kubernetes manifests using:

- **Environment-scoped YAML configuration** (hierarchical & composable)
- **Go text/templates** for flexible manifest templating
- **Structured include & repetition patterns** for re‑use without copy‑paste
- **Integrated schema validation** (Kubernetes native + custom CRDs via configuration)

It aims to sit in the “sweet spot” between plain hand‑written YAML and heavier ecosystems (Helm/Kustomize/Jsonnet), providing:

| Goal                   | How Miko-Manifest Approaches It                    |
| ---------------------- | -------------------------------------------------- |
| Predictable builds     | Deterministic merging + explicit includes          |
| Low cognitive overhead | Straightforward YAML + familiar Go templates       |
| Safe evolution         | Linting + schema validation at two distinct stages |
| Repeatable patterns    | "repeat" strategies (same-file & multi-file)       |
| Extensibility          | Library-first architecture                         |

### 1.1 Philosophy

- **Clarity over magic**: All expansion is explicit in config YAML.
- **Separation of stages**: Validate config _before_ build; validate manifests _after_ build.
- **Composable configuration**: Build larger environments from smaller bases.
- **Schema-driven rigor**: CRDs validated from declarative schema references.

### 1.2 Core Workflow (Mental Model)

```
config (input)  -->  check  -->  build (templates + config)  -->  validate (output)
```

You eliminate input errors early, then generate, then validate produced artifacts.

---

## 2. Installation

### 2.1 Quick Install (Recommended)

The easiest way to install miko-manifest is using our installation script:

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash
```

This script will:

- ✅ Detect your platform (Linux/macOS/Windows, amd64/arm64)
- ✅ Download the latest release binary from GitHub
- ✅ Remove any existing installations automatically
- ✅ Install to `/usr/local/bin` (or `~/.local/bin` if no sudo)
- ✅ Verify the installation works correctly
- ✅ Clean up temporary files

**Options:**

```bash
# View installation script help
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --help

# Uninstall miko-manifest completely
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --uninstall

# Custom installation directory (set before running)
export INSTALL_DIR="$HOME/bin"
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash
```

**Uninstallation:**

The installation script also provides an easy uninstallation option that will:

- ✅ Find all miko-manifest installations on the system
- ✅ Show current version before removal
- ✅ Remove binaries from all standard locations
- ✅ Provide sudo instructions for system-wide installations

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --uninstall
```

### 2.2 From Source (Go >= 1.21 recommended)

```bash
go install github.com/jepemo/miko-manifest@latest
```

`$GOBIN` (or `$GOPATH/bin`) should be on your `PATH`.

### 2.3 Docker Image

```bash
docker pull ghcr.io/jepemo/miko-manifest:latest
```

Run with your workspace mounted:

```bash
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest --help
```

### 2.4 Verifying Installation

```bash
miko-manifest --help
```

You should see all subcommands (config, check, build, validate, init, version).

#### Version Information

Check the installed version and build details:

```bash
# Show version information
miko-manifest version

# Example output:
# miko-manifest version 1.0.0
# commit: abc1234
# built: 2025-08-25_16:28:17
```

### 2.5 Docker Usage Examples

All commands work with Docker using the following pattern:

```bash
# Basic syntax
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest [command] [flags]

# Initialize a new project
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest init

# Check configuration
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest check

# Build manifests
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest build --env dev --output-dir output

# Validate manifests
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest validate --dir output

# Inspect configuration
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest config --env dev --tree
```

**Note**: The Docker image has the entrypoint pre-configured, so you only need to specify the command and flags.

---

## 3. Quick Start

```bash
# 1. Initialize a project scaffold
miko-manifest init

# 2. Inspect environment configuration
tail -n +1 config/*.yaml

# 3. Validate configuration (input stage)
miko-manifest check

# 4. Build manifests
miko-manifest build --env dev --output-dir output

# 5. Validate generated manifests
miko-manifest validate --dir output
```

Optional: combine build + post-build validation:

```bash
miko-manifest build --env dev --output-dir output --validate
```

---

## 4. Output System

Miko-manifest provides a dual-output system designed for both human-readable debugging and automation-friendly results.

### 4.1 Standard Mode (Default)

Shows only essential information:

- File validation results (`VALID`, `WARNING`, `ERROR`)
- Processing status (`PROCESSED`)
- Summary information (`SUMMARY`)

**Example:**

```bash
$ miko-manifest check
WARNING: config/schemas.yaml - Null value for key 'schemas'
SUMMARY: 2 file(s) validated successfully, 0 error(s)
SUMMARY: All YAML configuration files are valid
```

**Use cases:**

- CI/CD pipelines
- Automated scripts
- Production deployments
- Clean, parseable output

### 4.2 Verbose Mode (--verbose flag)

Shows detailed process information plus all standard output:

- Process steps (`INFO`, `STEP`, `DEBUG`)
- Internal workflow information
- Configuration loading details
- All standard mode messages

**Example:**

```bash
$ miko-manifest check --verbose
INFO: Using config directory: config
STEP: Checking YAML files in directory: config
STEP: Linting YAML files in config using native Go YAML parser
WARNING: config/schemas.yaml - Null value for key 'schemas'
SUMMARY: 2 file(s) validated successfully, 0 error(s)
SUMMARY: All YAML configuration files are valid
```

**Use cases:**

- Debugging configuration issues
- Understanding tool behavior
- Learning workflow steps
- Troubleshooting problems

### 4.3 Message Categories

| Prefix      | Mode     | Purpose               | Example                                                 |
| ----------- | -------- | --------------------- | ------------------------------------------------------- |
| `VALID`     | Standard | Successful validation | `VALID: deployment.yaml - Valid Deployment manifest`    |
| `WARNING`   | Standard | Non-blocking issues   | `WARNING: config.yaml - Null value for key 'schemas'`   |
| `ERROR`     | Standard | Blocking errors       | `ERROR: template.yaml - Invalid YAML syntax`            |
| `PROCESSED` | Standard | File operations       | `PROCESSED: template.yaml -> output/deployment.yaml`    |
| `SUMMARY`   | Standard | Final results         | `SUMMARY: 5 file(s) validated successfully, 0 error(s)` |
| `INFO`      | Verbose  | Informational         | `INFO: Using config directory: config`                  |
| `STEP`      | Verbose  | Process steps         | `STEP: Linting YAML files in config`                    |
| `DEBUG`     | Verbose  | Debug details         | `DEBUG: Loading schema from file.json`                  |

---

## 5. Command Deep Dive

### 5.0 Global Flags

#### Version Information

Get version, commit, and build information:

```bash
miko-manifest version

# Example output:
# miko-manifest version 1.0.0
# commit: abc1234
# built: 2025-08-25_16:28:17
```

#### Help

Get command help:

```bash
miko-manifest --help           # General help
miko-manifest [command] --help # Command-specific help
```

### 5.1 `init`

Scaffolds directories and example templates.

```bash
miko-manifest init
```

Creates (if absent):

- `templates/`
- `config/`
- Example environment YAML & template files

### 5.2 `config`

Inspect merged environment configuration.

```bash
miko-manifest config --env dev
```

Flags:

- `--variables` – print only `key=value` pairs (automation friendly)
- `--schemas` – list configured schema sources
- `--tree` – show hierarchical resource inclusion order
- `--verbose`, `-v` – show detailed processing information and loading steps

Example (variables only):

```bash
miko-manifest config --env prod --variables
```

Example (verbose tree display):

```bash
miko-manifest config --env dev --tree --verbose
```

### 5.3 `check`

Validates _configuration_ YAML prior to generation.

```bash
miko-manifest check
```

Performs:

- YAML syntax validation
- Structural config validation
- Variable definitions and references verification

Flags:

- `--config`, `-c`: Configuration directory path (default: "config")
- `--verbose`, `-v`: Show detailed processing information

**Output Modes:**

- **Standard**: Shows only validation results and summary
- **Verbose**: Shows step-by-step processing information

### 5.4 `build`

Generates manifests from templates + environment variables.

```bash
miko-manifest build --env staging --output-dir build-out
```

Useful flags:

- `--var NAME=VALUE` (repeatable) – ad‑hoc overrides
- `--templates` / `--config` – non-default layout
- `--validate` – run post-build validation automatically
- `--verbose` – show detailed build and validation information
- `--debug-config` / `--show-config-tree` – introspection aids

### 5.5 `validate`

Validates _generated_ manifests (output stage):

```bash
miko-manifest validate --dir build-out
```

Performs:

1. YAML parsing
2. Kubernetes resource schema validation
3. Custom resource schema validation (from environment config if env detectable or provided)

Flags:

- `--dir`, `-d`: Directory to validate (can also be positional argument)
- `--env`, `-e`: Environment to load schemas from (auto-detected if not specified)
- `--config`, `-c`: Configuration directory path (default: "config")
- `--skip-schema-validation`: Skip schema loading for faster YAML-only validation
- `--verbose`: Show detailed validation information

**Output Modes:**

- **Standard**: Shows only validation results and file status
- **Verbose**: Shows detailed validation steps and process information

Auto environment detection: if the build wrote its marker file, you can simply:

```bash
miko-manifest validate output
```

### 5.6 Structured Workflow Summary

| Stage             | Command  | Purpose                      | Typical Failure Sources       |
| ----------------- | -------- | ---------------------------- | ----------------------------- |
| Inspect           | config   | Understand what will be used | Wrong variable values         |
| Input validation  | check    | Catch config errors early    | YAML typos, missing includes  |
| Generation        | build    | Produce target manifests     | Template errors, missing vars |
| Output validation | validate | Ensure deployability         | CRD mismatch, schema drift    |

---

## 6. Configuration Model

### 6.1 Environment File Anatomy

```yaml
# config/dev.yaml
resources:
  - base.yaml # Reusable base
  - components/ # Directory expansion (alphabetical)

variables:
  - name: app_name
    value: my-app
  - name: replicas
    value: "2"

include:
  - file: deployment.yaml
  - file: service.yaml
    repeat: multiple-files
    list:
      - key: frontend
        values:
          - name: service_name
            value: frontend-svc
          - name: service_port
            value: "80"
      - key: backend
        values:
          - name: service_name
            value: backend-svc
          - name: service_port
            value: "8080"

schemas:
  - ./schemas/my-crd.yaml
  - https://raw.githubusercontent.com/example/operator/main/crd.yaml
```

### 6.2 Sections Explained

| Section     | Purpose                                 | Notes                                               |
| ----------- | --------------------------------------- | --------------------------------------------------- |
| `resources` | Hierarchical composition                | Order matters; later can override earlier vars      |
| `variables` | Key/value pairs injected into templates | Later duplicates override earlier                   |
| `include`   | Templating instructions                 | Drives which templates render & repetition behavior |
| `schemas`   | External CRDs for validation            | Local paths, directories, or URLs                   |

### 6.3 Repetition Patterns

1. **Simple File** — just `file: deployment.yaml`
2. **Same-File Repeat** — `repeat: same-file` consolidates multiple rendered fragments separated by `---`
3. **Multiple Files** — `repeat: multiple-files` creates suffixed outputs (`name-key.yaml`)

Same-file example:

```yaml
include:
  - file: configmap.yaml
    repeat: same-file
    list:
      - key: db
        values:
          - name: config_name
            value: database-config
      - key: cache
        values:
          - name: config_name
            value: cache-config
```

### 6.4 Hierarchical Resource Merging

Rules:

1. Process `resources` in order.
2. Merge `variables` (last win).
3. Append `include` items.
4. Deduplicate schema entries (stable order maintained).

Diagnostics:

```bash
miko-manifest build --env dev --output-dir out --show-config-tree
miko-manifest build --env dev --output-dir out --debug-config
```

### 6.5 Safety Controls

- Circular include detection
- Max recursion depth (fails fast if exceeded)
- Clear error messages for invalid sections

---

## 7. Template Authoring

### 7.1 Basics

Templates are standard Go `text/template` files. Variables declared in configuration become top-level keys.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{ .app_name }}"
spec:
  replicas: { { .replicas } }
  template:
    spec:
      containers:
        - name: main
          image: "{{ .image }}:{{ .tag }}"
```

### 7.2 Helpful Patterns

| Need          | Pattern                                                              |
| ------------- | -------------------------------------------------------------------- |
| Default value | `{{ or .optional_var "default" }}`                                   |
| Uppercase     | `{{ .app_name \| upper }}` (add custom funcs by library integration) |
| Join list     | `{{ join "," .listVar }}` (if custom func registered)                |

### 7.3 Debugging Templates

Add a temporary debug template:

```yaml
# debug.tmpl
# Available keys:
# {{ printf "%#v" . }}
```

Render selectively by adding to `include` during troubleshooting.

### 7.4 Common Pitfalls

| Issue                      | Cause                            | Fix                                             |
| -------------------------- | -------------------------------- | ----------------------------------------------- |
| Empty rendered value       | Missing variable                 | Add to environment or override with `--var`     |
| Bad YAML (during validate) | Template braces inside YAML keys | Quote dynamic keys                              |
| Mixed indentation          | Incorrect spacing inside loops   | Keep indentation static, only substitute values |

---

## 8. Schema Validation

### 7.1 Where Schemas Come From

Configured in environment YAML under `schemas:`. Accepts:

- Absolute/relative local file paths
- Directories (recursive)
- Remote URLs (raw CRD YAML)

### 7.2 Validation Stages

| Stage                       | Trigger      | What is validated                                         |
| --------------------------- | ------------ | --------------------------------------------------------- |
| Config Stage (`check`)      | Before build | Structure + optional custom resource schemas (if present) |
| Manifest Stage (`validate`) | After build  | Full Kubernetes objects + CRDs                            |

### 7.3 Skipping & Performance

Speed up quick iterations:

```bash
miko-manifest validate --dir output --skip-schema-validation
```

### 7.4 Auto Environment Detection

When you build, a marker enables `validate` to infer which environment's schemas to load — reduces flags for CI pipelines.

---

## 9. Advanced Usage

### 8.1 Variable Overrides

```bash
miko-manifest build --env dev --output-dir out \
  --var image=nginx --var tag=1.27 --var replicas=4
```

CLI overrides always win (highest precedence).

### 8.2 Partial Builds (Selective Include Edits)

Comment out `include` entries in env YAML to temporarily isolate a subset.

### 8.3 Using as a Library

```go
import "github.com/jepemo/miko-manifest/pkg/mikomanifest"

opts := mikomanifest.BuildOptions{
    Environment:  "dev",
    OutputDir:    "output",
    ConfigDir:    "config",
    TemplatesDir: "templates",
    Variables:    map[string]string{"app_name": "svc"},
}
mm := mikomanifest.New(opts)
if err := mm.Build(); err != nil { /* handle */ }

lintOpts := mikomanifest.LintOptions{Directory: "output"}
_ = mikomanifest.LintDirectory(lintOpts)
```

### 8.4 CI/CD Examples

**GitHub Actions** (minimal):

```yaml
jobs:
  manifests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate config
        run: miko-manifest check
      - name: Build
        run: miko-manifest build --env dev --output-dir output --validate
```

**GitLab CI** (three stages):

```yaml
stages: [validate, build, verify]

validate-config:
  stage: validate
  image: ghcr.io/jepemo/miko-manifest:latest
  script: miko-manifest check

generate:
  stage: build
  image: ghcr.io/jepemo/miko-manifest:latest
  script: miko-manifest build --env dev --output-dir output
  artifacts: { paths: [output/] }

verify:
  stage: verify
  image: ghcr.io/jepemo/miko-manifest:latest
  script: miko-manifest validate --dir output
```

### 8.5 Debug Strategy Checklist

| Symptom                | Try                                              |
| ---------------------- | ------------------------------------------------ |
| Wrong variables        | `config --env X --variables`                     |
| Missing include        | `config --env X --tree`                          |
| Unexpected manifest    | Inspect final output + template source           |
| CRD validation failing | Confirm schema URL reachable / file path correct |

---

## 10. Examples (Patterns Cookbook)

### 9.1 Multi-Environment Evolution

```
config/base.yaml
config/staging.yaml
config/prod.yaml
```

`base.yaml` holds shared vars; staging & prod override scale parameters only.

### 9.2 Dynamic Port + Image Tag

```bash
miko-manifest build --env dev --output-dir out \
  --var image=myrepo/api --var tag=$(git rev-parse --short HEAD)
```

### 9.3 Generating a Matrix of Services

`config/services.yaml`:

```yaml
include:
  - file: service.yaml
    repeat: multiple-files
    list:
      - key: users
        values:
          - name: service_name
            value: users-api
          - name: service_port
            value: "8001"
      - key: orders
        values:
          - name: service_name
            value: orders-api
          - name: service_port
            value: "8002"
```

### 9.4 Layered CRD Schemas

```yaml
# base.yaml
schemas:
  - ./schemas/platform/
# dev.yaml
resources:
  - base.yaml
schemas:
  - ./schemas/experimental/
```

### 9.5 Selective Validation (Fast Feedback)

```bash
miko-manifest check
```

Then do a full manifest validation before merging:

```bash
miko-manifest validate --dir output --env dev
```

---

## 11. Comparison with Ecosystem Tools

| Tool         | Primary Abstraction                  | Strengths                            | Where Miko-Manifest Differs                    |
| ------------ | ------------------------------------ | ------------------------------------ | ---------------------------------------------- |
| Helm         | Charts (packaged templates + values) | Large ecosystem, packaging, releases | Miko is lighter, no release management layer   |
| Kustomize    | Patch & overlay YAML transforms      | Patch-based, no templating logic     | Miko uses templating + explicit include logic  |
| Jsonnet      | Data templating language             | Powerful language features           | Miko favors simplicity & YAML familiarity      |
| ytt          | YAML templating with Starlark        | Rich templating + schema annotations | Miko relies on Go templates only               |
| kpt          | Package-centric, function pipelines  | Advanced packaging + GitOps flows    | Miko focuses purely on generation + validation |
| Raw K8s YAML | Direct manifests                     | Zero tooling overhead                | Miko adds structure, reuse, validation         |

### 10.1 When to Choose Miko-Manifest

- You want _just enough_ structure without adopting Helm’s packaging semantics.
- You prefer plain Go templates over DSLs (Jsonnet/Starlark).
- You need both **input** and **output** validation stages.
- You favor explicit over magical layering.

### 10.2 When Another Tool Might Fit Better

| Scenario                                              | Consider      |
| ----------------------------------------------------- | ------------- |
| Need chart dependency management & release lifecycles | Helm          |
| Prefer patching existing vendor YAML                  | Kustomize     |
| Want programmable graph transformations               | Jsonnet / CUE |
| Need function pipelines & packaging                   | kpt           |
| Heavy schema authoring & validations integrated       | ytt           |

---

## 11. FAQ

**Q: Can I reference one variable inside another?**  
A: Currently resolution is single-pass; pre-compose in your environment file or override via CLI.

**Q: Does build order matter?**  
A: Only `resources` ordering affects override precedence.

**Q: How do I add custom template functions?**  
A: Use the Go library, wrap `New()` and supply your own template.FuncMap before calling `Build()`.

**Q: Are remote schemas cached?**  
A: Remote fetch strategy is implementation-dependent; plan for network access on first validation run.

**Q: Can I skip output validation in CI?**  
A: Not recommended, but you can omit the `validate` step or use `--skip-schema-validation` flag on the `validate` command for speed.

---

## 12. Troubleshooting Quick Table

| Problem                          | Likely Cause             | Resolution                                            |
| -------------------------------- | ------------------------ | ----------------------------------------------------- |
| Missing generated file           | Include block omitted    | Inspect env with `config --tree`                      |
| Wrong replica count              | Override not applied     | Check variable precedence / CLI override              |
| Validation fails on CRD          | Schema not found         | Confirm path/URL; run with verbose logs (future flag) |
| YAML parse error referencing `{` | Unquoted template markup | Quote dynamic values in YAML                          |

---

## 13. Conclusion

`miko-manifest` fills a pragmatic niche: **structured, validated manifest generation** using familiar primitives without imposing a packaging ecosystem. It complements tools like Helm (packaging) or Kustomize (patching) by offering a middle path—ideal for teams wanting reproducibility, environment layering, and schema assurance with minimal ceremony.

You can adopt it incrementally: start by wrapping existing templates, then layer in hierarchical configs, then enable schemas for CRDs as maturity grows.

> If this tool streamlines your delivery pipeline, consider contributing examples, schemas, or enhancements.

---

_End of Documentation_
