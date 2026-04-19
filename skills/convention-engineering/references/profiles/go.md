# Go Convention Profile

Conventions for Go service repositories.

## Table of Contents

- [Build System: Taskfile.dev](#build-system-taskfiledev)
- [Linting: golangci-lint](#linting-golangci-lint)
- [Formatting: gofumpt](#formatting-gofumpt)
- [Dependency Injection: Google Wire](#dependency-injection-google-wire)
- [Testing](#testing)
- [Architecture Contract](#architecture-contract)
- [Pre-Commit Hook](#pre-commit-hook)
- [Binary Embedding](#binary-embedding)

## Build System: Taskfile.dev

Prefer Taskfile.dev over Makefile. Advantages: YAML syntax, includes for modularity, checksum-based caching.

### Root Structure

```yaml
# Taskfile.yml
version: "3"
includes:
  setup:
    taskfile: ./deploy/setup.yml
    dir: .
  core:
    taskfile: ./taskfiles/core.yml
    dir: .
    flatten: true # tasks available at root level
  harness:
    taskfile: ./taskfiles/harness.yml
    dir: .
    flatten: true
```

### Standard Task Names

| Task        | Purpose              | Example                                                  |
| ----------- | -------------------- | -------------------------------------------------------- |
| `build`     | Build all artifacts  | `go build -o bin/app ./cmd/app`                          |
| `test`      | Run safe local tests | `go test $(go list ./... \| grep -v '/expensive') -race` |
| `test:full` | Run ALL tests        | `go test ./... -v -race`                                 |
| `lint`      | Run linter           | `golangci-lint run ./...`                                |
| `fmt`       | Format code          | `gofumpt -w .`                                           |
| `serve`     | Build + run server   | Build then exec binary                                   |
| `wire`      | Regenerate DI code   | `cd service && wire gen`                                 |
| `verify`    | Canonical full gate  | lint + test + smoke + regress                            |

### Checksum Caching

```yaml
tasks:
  build:go:
    cmds: [go build -o bin/app ./cmd/app]
    sources: ["**/*.go", "go.mod", "go.sum"]
    generates: ["bin/app"]
    method: checksum # skip if sources unchanged
```

## Linting: golangci-lint

### Required Linters

- **revive**: Cognitive complexity (<=20), cyclomatic complexity (<=20)
- **depguard**: Layer boundary enforcement (THE single most valuable lint rule)

### Config Template

```yaml
# .golangci.yml
version: "2"
linters:
  enable: [revive, depguard]
  disable: [lll, wsl]
  settings:
    revive:
      confidence: 0.8
      rules:
        - name: cognitive-complexity
          arguments: [20]
        - name: cyclomatic
          arguments: [20]
    depguard:
      rules:
        handler-layer:
          list-mode: lax
          files: ["**/handler/**", "!$test"]
          deny:
            - pkg: "example.com/project/dal"
              desc: "handler must not import dal directly"
```

## Formatting: gofumpt

Use `gofumpt` (stricter than `gofmt`). Enforced in pre-commit hook.

## Dependency Injection: Google Wire

Compile-time DI via `google/wire`:

- Wire definition: `service/wire.go`
- Generated code: `service/wire_gen.go` (DO NOT EDIT)
- Regenerate: `task wire` after provider changes
- Provider functions return `(T, error)` or `(T, func(), error)`

## Testing

### Framework

- Standard `testing` package + `testify/assert` for assertions
- Colocated `*_test.go` files (same package)

### Test Split

- `task test`: Excludes expensive packages (e.g., YARA/CGo tests on dev machines)
- `task test:full`: Full suite for CI/server environments

### Mocking

- `mock/` package for test doubles (MemoryFileSystem, MockConfig, etc.)
- `dal/memory_store.go` for in-memory storage (no external deps)
- No external API mocks -- all tests use local fixtures

### E2E Tests

- Handler API tests in `handler/api_*_e2e_test.go`
- Test the HTTP surface, not internal functions

## Architecture Contract

5-layer depguard pattern — see `core/architecture-contracts.md` for the full reference diagram and enforcement config. Summary:

- `handler` → service, model, config (never dal, operator, pkg)
- `model` → nothing internal (zero deps guaranteed)

Interface assertions for every concrete-to-interface pairing:

```go
var _ Store = (*MongoStore)(nil)
var _ Store = (*MemoryStore)(nil)
```

## Pre-Commit Hook

```bash
#!/bin/bash
set -e
if command -v gofumpt &> /dev/null; then gofumpt -w .; fi
if command -v golangci-lint &> /dev/null; then golangci-lint run --fix; fi
go mod tidy
git ls-files --modified | grep -E '\.(go|mod|sum)$' | xargs -r git add
```

Install via `git config core.hooksPath .githooks` in setup task.

## Binary Embedding

Embed frontend SPA into Go binary for single-artifact deployment:

```go
//go:embed dist
var frontendFS embed.FS
```

Build order: frontend (bun/vite) -> embed -> Go build -> single binary.
