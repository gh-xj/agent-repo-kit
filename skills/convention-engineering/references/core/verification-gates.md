# Verification Gates

Every repo needs one canonical command that runs all quality gates. If it passes, the code is ready.

## Table of Contents

- [The Canonical Gate](#the-canonical-gate)
- [Gate Checklist (8 steps)](#gate-checklist-8-steps)
- [Agent-Debuggable Output Contract](#agent-debuggable-output-contract)
- [Pre-Commit Hook](#pre-commit-hook)
- [Taskfile Pattern](#taskfile-pattern)
- [Verification Before Claiming Done](#verification-before-claiming-done)

## The Canonical Gate

One command. Runs everything. Same command locally and in CI.

```bash
task verify    # or: make verify, npm run verify, etc.
```

## Gate Checklist (8 steps)

A complete gate includes these checks in order:

### 1. Format Check
- Go: `gofumpt -l .` (list unformatted files)
- TypeScript: `prettier --check .`
- Python: `ruff format --check .`

### 2. Lint
- Go: `golangci-lint run ./...`
- TypeScript: `eslint .`
- Python: `ruff check .`

### 3. Type Check
- Go: compiler (automatic via `go build`)
- TypeScript: `tsc --noEmit` (strict mode)
- Python: `mypy .` or `pyright .`

### 4. Unit Tests
- Go: `go test ./... -race`
- TypeScript: `vitest run` or `jest`
- Python: `pytest`

### 5. Architecture Boundary Check
- Go: depguard (runs as part of golangci-lint)
- TypeScript: eslint-plugin-boundaries or custom script
- Python: import-linter

### 6. Entropy/Drift Detection (recommended)
- Stale plans detection (plans older than N days)
- Orphaned references (dead links in docs)
- Mirrored doc drift (CLAUDE.md vs AGENTS.md)

### 7. Smoke Tests (if harness exists)
- Fast deterministic lifecycle checks
- Artifact contract validation (output format hasn't changed)

### 8. Regression Tests (if fixtures exist)
- Fixture-driven scenario replay
- Known-good output comparison

## Agent-Debuggable Output Contract

- Canonical gate should emit a machine-readable run summary (`summary.json` or equivalent).
- Canonical gate should emit per-step logs with stable paths so failures are inspectable without rerunning.
- Failure output should include: failed command, exit code, log path, and short tail excerpt.
- Dependency preflight should fail early with actionable hints (for example `bun` missing for frontend gates).

## Pre-Commit Hook

Minimum viable hook (must complete in <10 seconds):

```bash
#!/bin/bash
set -e

# Format
gofumpt -w .           # or: prettier --write .

# Lint (with auto-fix)
golangci-lint run --fix  # or: eslint --fix .

# Dependency tidy
go mod tidy             # or: nothing for TS/Python

# Stage modified files
git ls-files --modified | grep -E '\.(go|mod|sum)$' | xargs -r git add
```

Keep the hook fast. Move expensive checks (tests, full lint) to the verify command.

## Taskfile Pattern

```yaml
# taskfiles/core.yml
tasks:
  verify:
    desc: Canonical local verification command
    cmds:
      - task: lint
      - task: test
      - task: check:arch-boundaries  # if harness exists
      - task: smoke                   # if harness exists
      - task: regress                 # if harness exists
```

## Verification Before Claiming Done

Before any agent marks a task complete:
- Go changes: `task lint` + `task test`
- Frontend changes: `task lint:web` + `task build:web`
- DI changes: `task wire` (Wire code generation)
- Large refactors: `task verify` (full suite)
