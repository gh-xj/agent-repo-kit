# Verification Gates

Every repo needs one canonical command that runs all quality gates. If it
passes, the work is ready.

## The Canonical Gate

One command. Same locally and in CI.

```bash
task verify
```

What `task verify` does is the repo's choice. The skill does not prescribe
specific tools — that is part of the repo's `.conventions.yaml`.

## How `task verify` Relates to `.conventions.yaml`

The descriptor declares which gates the repo opts into. The Taskfile wires
each opt-in to a concrete command. For the recommended pure-shell
"mechanical floor" pattern that asserts every typed opt-in via `yq`, see
`references/core/verify-script-pattern.md`.

```yaml
# .conventions.yaml
taskfile: true
pre_commit: true
checks:
  - "task verify exits 0 from a clean checkout."
  - "task verify runs at least: format check, lint, type check, tests."
```

```yaml
# Taskfile.yml
tasks:
  verify:
    cmds:
      - task: fmt:check
      - task: lint
      - task: typecheck
      - task: test
```

The agent reads `.conventions.yaml` to know what should pass; `task verify`
is the single entry point that exercises it.

## Output Contract

The canonical gate emits agent-debuggable output:

- Per-step exit codes are visible (one task per check, not a monolith).
- Failures include: failed command, exit code, log path, short tail excerpt.
- Optional: a machine-readable summary (`summary.json` or equivalent) for
  automated downstream consumers.

## Pre-Commit Hook

A fast hook (under ~10 seconds) for format / auto-fix / dep-tidy concerns.
Heavy checks belong in `task verify`, not the hook.

The hook is intentionally repo-defined. The skill only requires that one
exists when `pre_commit: true` is set in `.conventions.yaml`, and that
`core.hooksPath` (or installed equivalent) wires it up.

## Verification Before Claiming Done

Before any agent marks a task complete:

- Run `task verify`.
- Read its output. Silence is not success — confirm the gate actually
  exercised the relevant code paths.
- If a fast subset is appropriate (e.g. one of `task lint`, `task test`),
  use that, then run `task verify` once before commit.
