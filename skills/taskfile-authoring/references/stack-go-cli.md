# Stack Pattern: Go CLI

Advisory guidance for Go CLI projects. Conventions that have shaken out
across several internal repos. Apply them with judgment.

## Canonical Task Names

A Go CLI Taskfile usually fits in ten tasks or fewer:

- `deps` — `go mod tidy`
- `fmt` — write-in-place formatter (`gofmt -w .` or `gofumpt -w .`)
- `fmt:check` — fails if formatter would change files (gate for other tasks)
- `lint` — `go vet ./...` or `golangci-lint run`
- `test` — `go test ./...`
- `build` — produce `bin/<name>`
- `run` — `go run . {{.CLI_ARGS}}` for the dev loop
- `smoke` — exercise the built artifact in `--json` mode, validate parseability
  with `jq -e .`
- `ci` — aggregate: `deps: [lint, test, build, smoke]`
- `verify` — `deps: [ci]` plus repo-specific extras

## `fmt:check` Gates The Rest

Each mutating task depends on `fmt:check` so CI fails fast on unformatted
trees:

```yaml
lint:
  deps: [deps, fmt:check]
  cmds:
    - go vet ./...

test:
  deps: [deps, fmt:check]
  cmds:
    - go test ./...

build:
  deps: [deps, fmt:check]
  sources: ["**/*.go", "go.mod", "go.sum"]
  generates: ["bin/myapp"]
  method: checksum
  cmds:
    - mkdir -p bin
    - go build -o bin/myapp .
```

## `build` Must Cache

Declare `sources`, `generates`, `method: checksum`. Without these, every
`task ci` rebuilds from scratch, which blows up smoke-test latency.

## `smoke` Validates JSON Output

Agents consume these CLIs non-interactively. The smoke task runs the built
binary with `--json` (or `--ndjson`) and validates the output against a JSON
schema committed under `test/smoke/`:

```yaml
smoke:
  desc: Deterministic smoke checks
  deps: [build]
  cmds:
    - mkdir -p test/smoke
    - rm -f test/smoke/version.output.json
    - ./bin/myapp --json version > test/smoke/version.output.json
    - jq -e . test/smoke/version.output.json >/dev/null
```

## `--json` / `--ndjson` Are Unstyled

Never emit ANSI color, spinners, or progress bars in machine-mode output.
Agents parse line-by-line; styling breaks JSON decoding. Keep color and
Bubble Tea-style UI behind an explicit `--tty` or "no `--json` flag"
condition.

## Tool Choice Is Adopter-Preference

`gofmt` vs `gofumpt`, `go vet` vs `golangci-lint` — both pairs are legitimate.
Pick one per repo and stick with it. The lint layer does **not** enforce a
specific choice.

## `run` Uses `{{.CLI_ARGS}}`

`task run -- version --json` threads CLI args through to the binary:

```yaml
run:
  desc: Run the CLI without building
  cmds:
    - go run . {{.CLI_ARGS}}
```

## Anti-Patterns

- **`task install`** — wraps `go install ./...` in one line. Redundant; users
  already know `go install`. Drop it.
- **`task help`** — Cobra's `--help` and the default `task` listing cover
  this. A separate task is noise.
- **Unquoted colon-space in shell strings** — `cmd: echo "Failed: see logs"`
  parses as a YAML mapping and fails obscurely. Always quote commands that
  contain `: `.
- **Per-package test tasks** (`test:foo`, `test:bar`, `test:baz`) — consolidate
  into `go test ./...` and filter with `-run` when needed.
- **`build` without `sources:`** — every CI run recompiles. Add
  `sources: ['**/*.go', 'go.mod', 'go.sum']`, `generates: ['bin/<name>']`,
  `method: checksum`.
