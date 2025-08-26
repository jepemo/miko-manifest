## Miko-Manifest

[![Release](https://img.shields.io/github/v/release/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/jepemo/miko-manifest/ci-cd.yml)](https://github.com/jepemo/miko-manifest/actions)
[![CodeQL](https://github.com/jepemo/miko-manifest/actions/workflows/codeql.yml/badge.svg)](https://github.com/jepemo/miko-manifest/actions/workflows/codeql.yml)

<!-- [![Go Report](https://goreportcard.com/badge/github.com/jepemo/miko-manifest)](https://goreportcard.com/report/github.com/jepemo/miko-manifest) -->

[![Go Version](https://img.shields.io/github/go-mod/go-version/jepemo/miko-manifest)](https://github.com/jepemo/miko-manifest/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jepemo/miko-manifest)](LICENSE)

Declarative, hierarchical configuration and robust templating for Kubernetes manifests. Fast, deterministic, script‑friendly.

---

### 1. Purpose

Miko-Manifest unifies three concerns normally spread across ad‑hoc scripts:

1. Structured, hierarchical environment configuration (YAML with controlled merging).
2. Deterministic manifest generation via Go templates (single, repeated-in-file, or repeated-to-multiple-files patterns).
3. Integrated validation (YAML structure, Kubernetes schemas, and custom CRDs) before and after generation.

The result: repeatable builds, early failure detection, and transparent configuration provenance.

### 2. Quick Start

**Easy installation (recommended):**

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash
```

**Easy uninstallation:**

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --uninstall
```

**Alternative installations:**

```bash
# Using Go
go install github.com/jepemo/miko-manifest@latest

# Using Docker
docker pull ghcr.io/jepemo/miko-manifest:latest
```

Scaffold a project:

```bash
miko-manifest init
```

Generate manifests for an environment:

```bash
miko-manifest build --env dev --output-dir output
```

Validate generated manifests:

```bash
miko-manifest validate --dir output
```

Inspect effective configuration (recommended before first build):

```bash
miko-manifest config --env dev --tree --verbose
```

For comprehensive flags and advanced scenarios consult: [DOCS.md](DOCS.md)

### 3. Core Concepts (Essentials Only)

| Concept                | Summary                                                                                                                                                                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Hierarchical Resources | `resources:` lists files or directories. They are merged in order. Later variables override earlier ones. Includes and schema lists are concatenated (deduplicated logically). |
| Templates              | Standard Go templates located under `templates/`. Supports: plain render, same-file repetition (multiple YAML docs), multi-file repetition (one output per list item).         |
| Variables              | Declared in config YAML (`variables:` as name/value pairs) or overridden with `--var key=value` at build time. Unified into a final map fed to templates.                      |
| Includes               | `include:` drives which template(s) render and how (optionally with `repeat` + `list`).                                                                                        |
| Schemas                | Optional `schemas:` entries (local file, directory, or URL). Used to validate generated manifests (including CRDs).                                                            |
| Output Modes           | Standard (concise) vs `--verbose` (steps + context). Consistent across commands.                                                                                               |

### 4. Typical Workflow

```bash
# (Optional) Understand configuration and inheritance
miko-manifest config --env dev --tree

# Lint configuration inputs early
miko-manifest check

# Generate manifests
miko-manifest build --env dev --output-dir output

# Validate output (YAML + Kubernetes + custom schemas)
miko-manifest validate --dir output
```

### 5. Minimal Example

`config/dev.yaml`:

```yaml
variables:
  - name: app_name
    value: demo
include:
  - file: deployment.yaml
```

`templates/deployment.yaml` (fragment):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: { { .app_name } }
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: { { .app_name } }
          image: nginx:latest
```

Build:

```bash
miko-manifest build --env dev --output-dir output
```

Result: `output/deployment.yaml`

### 6. Command Overview (Condensed)

| Command    | Purpose                                                   | Typical Additions                                 |
| ---------- | --------------------------------------------------------- | ------------------------------------------------- |
| `init`     | Scaffold directories and example templates                | –                                                 |
| `config`   | Inspect merged configuration / schemas / tree / variables | `--tree`, `--schemas`, `--variables`, `--verbose` |
| `check`    | Validate configuration YAML before build                  | `--verbose`                                       |
| `build`    | Render templates into manifest files                      | `--var`, `--validate`, `--verbose`                |
| `validate` | Validate generated manifests (YAML + schemas)             | `--env`, `--skip-schema-validation`, `--verbose`  |

Complete flag descriptions: see [DOCS.md](DOCS.md).

### 7. Advanced Highlights

| Topic               | Detail                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------ |
| Deterministic Order | Directories processed alphabetically; merging order = declaration order.                         |
| Override Strategy   | Variable last-write-wins; template includes accumulate; schemas aggregated (duplicates ignored). |
| Safety              | Circular resource inclusion detection + maximum depth guard.                                     |
| Auto Environment    | `build` records environment; `validate` reuses it if `--env` omitted.                            |
| Schema Sources      | Local file, directory (recursive), or remote URL (fetched once per run).                         |

### 8. Programmatic Use

```go
import "github.com/jepemo/miko-manifest/pkg/mikomanifest"

opts := mikomanifest.BuildOptions{
    Environment:  "dev",
    OutputDir:    "output",
    ConfigDir:    "config",
    TemplatesDir: "templates",
    Variables:    map[string]string{"app_name": "demo"},
}
mm := mikomanifest.New(opts)
if err := mm.Build(); err != nil { /* handle */ }
```

More constructors, linting, and validation helpers are listed in [DOCS.md](DOCS.md).

### 9. Docker & CI

Run without local toolchain:

```bash
# Check configuration
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest check

# Build manifests
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest build --env dev --output-dir output

# Validate generated manifests
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/jepemo/miko-manifest:latest validate --dir output
```

Typical pipeline: check -> build -> validate. Minimal GitHub Actions and GitLab CI templates are provided in the documentation.

### 10. Contributing

Contributions are welcome. Please open an issue to propose significant changes before submitting a pull request. Ensure tests cover new behaviour and run `go test ./...` locally.

### 11. License

MIT — see [LICENSE](LICENSE).

---

Further detail (all flags, schema handling, repetition patterns, error modes): [DOCS.md](DOCS.md)
