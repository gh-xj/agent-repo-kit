# Taskfile Anti-Patterns

Common mistakes with before/after fixes. Each entry states why the pattern
fails, then shows the minimal correction.

## 1. `echo`-as-docs

**Why this fails:** users run `task --list` to discover commands. `echo`-ing
documentation inside `cmds:` means the text only surfaces when the task runs,
and there is no way for `task --list` to display it.

```yaml
# BEFORE
tasks:
  build:
    cmds:
      - echo "Builds the project. Outputs to bin/."
      - go build -o bin/app .
```

```yaml
# AFTER
tasks:
  build:
    desc: Build the project; outputs to bin/
    cmds:
      - go build -o bin/app .
```

## 2. Variable Triangulation

**Why this fails:** three vars encoding one fact (`bin/myapp`) means every
rename touches three lines, and drift is silent.

```yaml
# BEFORE
vars:
  BIN_DIR: bin
  BIN_NAME: myapp
  BIN_PATH: "{{.BIN_DIR}}/{{.BIN_NAME}}"
```

```yaml
# AFTER
vars:
  BIN: bin/myapp
```

## 3. Unquoted Colon-Space In YAML

**Why this fails:** YAML treats `key: value` as a mapping. A shell string with
`: ` in it confuses the parser and produces a cryptic error or, worse, silently
reinterprets the line as nested keys.

```yaml
# BEFORE
tasks:
  release:
    cmds:
      - echo Failed: see logs
```

```yaml
# AFTER
tasks:
  release:
    cmds:
      - 'echo "Failed: see logs"'
```

The same trap applies to `desc:` lines (and any other scalar value with a
literal `: `) — the YAML parser doesn't care which field it is. Wrap in single
or double quotes:

```yaml
# WRONG — YAML parses the inner colon as a nested mapping; `task` errors out
verify:banned-strings:
  desc: Defense-in-depth privacy: regex list of strings that must never ship

# RIGHT
verify:banned-strings:
  desc: 'Defense-in-depth privacy: regex list of strings that must never ship'
```

## 4. `task install` Redundant With The Tool

**Why this fails:** Go and uv already have install verbs (`go install`,
`uv tool install`). Wrapping them adds surface without adding value, and
confuses users about which is canonical.

```yaml
# BEFORE
tasks:
  install:
    cmds:
      - go install ./...
```

```yaml
# AFTER — delete the task; document `go install ./...` in README if needed
```

## 5. `flatten: true` With Colliding Names

**Why this fails:** with `flatten: true`, an included task with the same name as
a root task silently shadows it. Six months later nobody remembers which
definition wins.

```yaml
# BEFORE — root and include both define `test`
includes:
  conventions:
    taskfile: ./.conv/Taskfile.yml
    flatten: true

tasks:
  test:
    cmds:
      - go test ./...
# …and ./.conv/Taskfile.yml also defines `test`
```

```yaml
# AFTER — exclude the colliding name, or drop flatten
includes:
  conventions:
    taskfile: ./.conv/Taskfile.yml
    flatten: true
    excludes: [test]

tasks:
  test:
    cmds:
      - go test ./...
```

## 6. Dotenv In An Included File

**Why this fails:** go-task only loads `dotenv:` at the **root** Taskfile.
Declaring it in an included file produces no error and loads nothing.

```yaml
# BEFORE — .conv/Taskfile.yml (included)
version: "3"
dotenv: [".env"] # silently ignored
tasks:
  verify:
    cmds: [go run ./scripts/audit.go]
```

```yaml
# AFTER — move dotenv to the root Taskfile
# root Taskfile.yml
version: "3"
dotenv: [".env"]
includes:
  conventions:
    taskfile: ./.conv/Taskfile.yml
```

## 7. Too Many User-Facing Tasks

**Why this fails:** once `task --list` shows eight or more tasks, the canonical
surface (`fmt`, `lint`, `test`, `build`, `ci`, `verify`) gets buried. New
contributors cannot tell which one to run.

```yaml
# BEFORE — user-facing tasks for every directory
tasks:
  test:unit: { cmds: [go test ./internal/unit/...] }
  test:integration: { cmds: [go test ./internal/integration/...] }
  test:e2e: { cmds: [go test ./internal/e2e/...] }
  build:cli: { cmds: [go build -o bin/cli ./cmd/cli] }
  build:server: { cmds: [go build -o bin/server ./cmd/server] }
  # ...
```

```yaml
# AFTER — one user-facing task, filter with CLI_ARGS when needed
tasks:
  test:
    desc: Run tests
    cmds:
      - go test ./...

  build:
    desc: Build all binaries
    cmds:
      - go build -o bin/ ./cmd/...
```

## 8. `method: timestamp` In CI

**Why this fails:** `git clone` and most CI cache restores reset mtimes. With
`method: timestamp`, go-task sees the sources as "newer than the output" every
time and re-runs the task unconditionally — defeating the cache.

```yaml
# BEFORE
build:
  sources: ["**/*.go"]
  generates: ["bin/app"]
  method: timestamp
```

```yaml
# AFTER
build:
  sources: ["**/*.go"]
  generates: ["bin/app"]
  method: checksum
```

## 9. Missing `.task/` In `.gitignore`

**Why this fails:** `.task/` holds the per-file checksum database. Committing it
means every contributor's cache is seeded with whoever pushed last, causing
spurious "up to date" skips on trees that differ.

```gitignore
# BEFORE — .gitignore
bin/
dist/
```

```gitignore
# AFTER
bin/
dist/
.task/
```

## 10. Hard-Coded `../` In Included Commands

**Why this fails:** included Taskfiles can be relocated (dropped into a
different repo, moved up or down a directory). Relative paths like `../` assume
a fixed layout and break silently when the assumption changes.

```yaml
# BEFORE — .conv/Taskfile.yml
tasks:
  audit:
    cmds:
      - go run ../scripts/audit.go
```

```yaml
# AFTER
tasks:
  audit:
    cmds:
      - go run {{.TASKFILE_DIR}}/scripts/audit.go
```

## 11. Dynamic `sh:` Var Relied On As Fingerprint

**Why this fails:** go-task's checksum fingerprint reads `sources:` file
content. It does **not** include the values of `vars:` (dynamic or static). A
task whose output depends on `git rev-parse HEAD` will not rebuild when HEAD
moves unless you encode that into `status:`.

```yaml
# BEFORE — expecting var change to invalidate cache
vars:
  COMMIT:
    sh: git rev-parse HEAD

tasks:
  release:
    sources: ["**/*.go"]
    generates: ["bin/app"]
    method: checksum
    cmds:
      - go build -ldflags="-X main.Commit={{.COMMIT}}" -o bin/app .
```

```yaml
# AFTER — encode the external state in status:
vars:
  COMMIT:
    sh: git rev-parse HEAD

tasks:
  release:
    method: none
    status:
      - test -f bin/app
      - test "$(./bin/app --version-commit 2>/dev/null)" = "{{.COMMIT}}"
    cmds:
      - go build -ldflags="-X main.Commit={{.COMMIT}}" -o bin/app .
```

## 12. `dir:` Plus A Shell Redirect Or A Relative-Path CLI Arg

**Why this fails:** `dir:` shifts the cwd for **everything** in `cmds:` —
including shell redirects (`2> $LOG`) and any relative path that the spawned
binary itself receives (`--cd sandbox/foo`). The path you wrote relative to the
_repo root_ now resolves relative to the _task's `dir:`_, and the failure mode
is silent (file not found, or worse, written to a different location than
expected). The mistake is invisible at task-lint time and only surfaces at run
time.

```yaml
# BEFORE — looks fine, breaks at runtime
tasks:
  trace:
    dir: codex/codex-rs
    vars:
      LOG: '{{.LOG | default "trace.log"}}'
    cmds:
      # Both of these resolve from codex/codex-rs/, NOT the repo root:
      - cargo run --bin codex -- {{.CLI_ARGS}} 2> {{.LOG}}
      # If CLI_ARGS contains "--cd sandbox/foo", that resolves wrong too.
```

Two safe patterns. **Pick whichever keeps the binary's relative-path
expectations intact.**

```yaml
# AFTER (a) — don't change dir; use --manifest-path. cwd stays at repo
# root, so the user's $LOG and --cd args resolve where they expect.
tasks:
  trace:
    vars:
      LOG: '{{.LOG | default "trace.log"}}'
    cmds:
      - cargo run --manifest-path {{.TASKFILE_DIR}}/codex/codex-rs/Cargo.toml \
        --bin codex -- {{.CLI_ARGS}} 2> {{.LOG}}
```

```yaml
# AFTER (b) — keep dir, but absolutize every path that crosses the
# task boundary, using {{.TASKFILE_DIR}} as the anchor.
tasks:
  trace:
    dir: codex/codex-rs
    vars:
      LOG: '{{.LOG | default "trace.log"}}'
      LOG_ABS:
        '{{if hasPrefix "/"
        .LOG}}{{.LOG}}{{else}}{{.TASKFILE_DIR}}/{{.LOG}}{{end}}'
    cmds:
      - mkdir -p "$(dirname '{{.LOG_ABS}}')"
      - cargo run --bin codex -- {{.CLI_ARGS}} 2> '{{.LOG_ABS}}'
```

The `LOG_ABS` template handles both forms — pass-through if the user already
provided an absolute path, otherwise anchor it at the repo root.

**Heuristic:** if a `dir:` task takes either `LOG=…` or `--cd …` style arguments
from the caller, prefer pattern (a) — the caller almost certainly meant their
paths to resolve from the repo root, not from inside whatever directory you
happened to chose.

## 13. Literal `{{` Anywhere In A `cmd:` Block (Including Comments)

**Why this fails:** go-task processes every line in a `cmd:` block as a Go
template before passing it to the shell. Any double-curly pair triggers
template evaluation. `{{ op://VAULT/ITEM }}` is read as a template expression
(unknown function `op`), and even literal `{{` inside shell-quoted strings
or shell-line comments errors with `template: <line>: missing value for
command`. The error message points at the template line — not the source —
so the cause is hard to find.

Most common bite: writing a heredoc/printf that emits another templated
config (1Password `op inject`, Mustache, Jinja, Helm, etc.) where the
emitted content itself uses `{{ }}` as its own delimiters.

```yaml
# BEFORE (fails: go-task tries to evaluate {{ op://... }})
tasks:
  render-cfg:
    cmds:
      - |
        # Build a section with an op:// secret ref:
        printf '[Chase]\naccess_token = {{ op://Private/chase }}\n' >> cfg.tpl
```

```yaml
# AFTER — build the delimiters at runtime from single characters
# so go-task never sees two adjacent open-curlies in the source.
tasks:
  render-cfg:
    cmds:
      - |
        b1='{'; b2='}'
        ob="${b1}${b1}"; cb="${b2}${b2}"
        printf '[Chase]\naccess_token = %s op://Private/chase %s\n' \
          "$ob" "$cb" >> cfg.tpl
```

The same trap applies in comments: a shell-line `# foo {{ bar }}` is parsed
by go-task before the shell ever sees it. Strip the curlies from prose
("the double-curly form" instead of `{{ … }}`), or move the comment outside
the `cmd:` block as a YAML `# comment` line.

## 14. `{{.CLI_ARGS}}` Includes The Caller's Shell Quotes Verbatim

**Why this fails:** `task foo -- "Display Name"` makes go-task substitute
`{{.CLI_ARGS}}` with the literal string `'Display Name'` (single-quoted by
the user's shell as it passes the arg through `task`). If the task simply
assigns `name="{{.CLI_ARGS}}"` and uses `$name` downstream, the quotes
become part of the value — `[Display Name]` ends up `['Display Name']` in
any file the task writes, and `op item edit` / `git commit -m` may use the
wrong slug.

```yaml
# BEFORE (name carries the outer quotes)
tasks:
  link:
    cmds:
      - |
        name="{{.CLI_ARGS}}"        # name == "'Chase Credit Cards'"
        echo "Linking $name"        # Linking 'Chase Credit Cards'
        write_section "$name"       # section header becomes ['Chase Credit Cards']
```

```yaml
# AFTER — strip a single layer of surrounding ' or "
tasks:
  link:
    cmds:
      - |
        name="{{.CLI_ARGS}}"
        name="${name#[\'\"]}"; name="${name%[\'\"]}"   # safe: only outermost layer
        if [ -z "$name" ]; then
          echo "usage: task link -- \"Display Name\"" >&2
          exit 2
        fi
        echo "Linking $name"        # Linking Chase Credit Cards
```

This handles the common case (one quoting layer added by zsh/bash). If
callers may pre-quote deliberately, document the expected form in `desc:`
so users don't double-quote.

`{{.CLI_ARGS}}` is split across multi-arg calls too, so `task foo -- a b c`
yields `a b c` (space-joined) — fine for tools that re-tokenize on spaces,
not fine for an arg with embedded spaces. If you need true argv handling,
parse `{{.CLI_ARGS}}` through `eval set -- {{.CLI_ARGS}}` cautiously or
require quoted single-arg form in `desc:`.
