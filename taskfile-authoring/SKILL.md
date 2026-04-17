---
name: taskfile-authoring
description: Use when writing, creating, or refactoring Taskfile.yml for any project (Go CLI, Python uv, or generic). Covers go-task v3 features — includes/sub-Taskfiles, sources/generates/method build cache, deps vs cmds, canonical CI/verify surface. Use before adding tasks to an existing Taskfile, scaffolding a new one, or when fixing `ark taskfile lint` findings. Triggers on "write Taskfile", "create Taskfile.yml", "refactor Taskfile", "Taskfile.dev", "go-task", "sub-taskfile", "build cache", "task lint", "includes", "dotenv".
---

# Taskfile Authoring

Write lean, composable, verification-first Taskfiles for go-task v3. Keep the
surface small, the cache correct, and the composition transparent. Pair with
`ark taskfile lint` to catch structural mistakes automatically.

## When To Use

Use this skill when:

- Writing a new `Taskfile.yml` from scratch.
- Adding tasks to an existing Taskfile (before, not after, the diff).
- Refactoring an overgrown Taskfile (too many tasks, duplicated cmds).
- Splitting work into included sub-Taskfiles (`includes:`).
- Wiring `sources:` / `generates:` / `method:` for build caching.
- Resolving `ark taskfile lint` findings.

Do not use for:

- CI platform configs (`.github/workflows/*.yml`) — that is not go-task.
- Makefiles, `just` recipes, or npm scripts — different tooling, different semantics.
- Pure shell scripts with no task runner — use a `scripts/` directory instead.

## Core Premise

Three properties keep a Taskfile healthy:

1. **Lean surface**: few top-level tasks, each with a clear `desc:`. If a
   contributor cannot read the `task --list` output and guess what to run,
   there are already too many.
2. **Composable**: included sub-Taskfiles with explicit `{{.TASKFILE_DIR}}`
   paths, so an overlay dropped into any repo Just Works.
3. **Verification-first**: one canonical `ci` task that CI runs verbatim, and
   one `verify` task that wraps `ci` plus any repo-specific extras (ticket
   checks, wiki lint, etc.). No branching logic in CI.

## The Canonical Aggregate Surface (advisory)

Most repos benefit from this shape:

- `fmt` — format in place
- `lint` — static checks
- `test` — run tests
- `build` — produce artifact(s)
- `ci` — aggregate: depends on `lint`, `test`, `build` (plus `smoke` where applicable). This is what CI runs.
- `verify` — wraps `ci` and adds repo-specific gates (e.g. ticket/wiki checks).

This is **not** lint-enforced. It is a sizing nudge — see
`references/canonical-surface.md` for rationale and per-stack task counts.

## Build Cache: sources / status / preconditions

Three mechanisms, three different jobs:

- **`preconditions:`** — fail loudly if the environment is wrong (tool
  missing, file absent). Runs before the task body.
- **`status:`** — skip if "already up to date." You write the check. All
  commands must exit 0 for the task to be skipped. No automatic source tracking.
- **`sources:` + `generates:` + `method:`** — automatic fingerprinting. Use
  `method: checksum` (default, CI-safe). Never `method: timestamp` in CI —
  `git clone` resets mtimes, so the task always runs.

Full decision flow and per-use-case recipes in `references/build-cache.md`.

## Composition (`includes:`)

Two composition modes:

- `flatten: true` — transparent overlay. Included tasks become top-level. Use
  when a convention pack ships a Taskfile fragment (`.convention-engineering/Taskfile.yml`)
  and callers should type `task verify`, not `task conventions:verify`.
- Explicit `dir:` with a namespace key — genuine sub-domain. Use for
  `tickets:test`, `wiki:lint`, `docs:build` — things that feel like separate
  areas of the repo.

Variable precedence is surprising: **the included file's own `vars:` win over
the includer's `includes.ns.vars:`** (pattern: use `{{.FOO | default "..."}}`
inside the included file to make values overridable). Full precedence table and
path-resolution rules in `references/composition.md`.

## Running `ark taskfile lint`

```bash
ark taskfile lint --repo-root .
```

V1 ships ten structural rules. Each finding points at a specific line and a
one-line fix. See `references/lint-rules.md` for the full table: rule ID,
trigger example, rationale, and fix. Common outputs:

- `version-required` / `version-is-three` — add `version: '3'` at the top.
- `cmd-and-cmds-mutex` — a task defines both `cmd:` and `cmds:`; pick one.
- `fingerprint-dir-gitignored` — add `.task/` to `.gitignore`.
- `dotenv-files-gitignored` — the `.env` file you declare in `dotenv:` is tracked by git.

Strict-mode, stack-specific rules, per-repo config, and integration with
`ark check` are planned for v2 — see the DEFERRED section of
`references/lint-rules.md`.

## Reference & Template Index

| Topic                                  | File                                        |
| -------------------------------------- | ------------------------------------------- |
| Canonical `ci` / `verify` pattern      | `references/canonical-surface.md`           |
| `includes:` deep dive, variable scope  | `references/composition.md`                 |
| `sources` / `status` / `preconditions` | `references/build-cache.md`                 |
| Go CLI patterns (advisory)             | `references/stack-go-cli.md`                |
| Python + uv patterns (advisory)        | `references/stack-uv-python.md`             |
| Before/after anti-patterns             | `references/anti-patterns.md`               |
| `ark taskfile lint` V1 rule catalog    | `references/lint-rules.md`                  |
| Minimal generic starter                | `references/templates/Taskfile.generic.yml` |
| Go CLI starter (with build cache)      | `references/templates/Taskfile.go-cli.yml`  |
| uv + typer starter                     | `references/templates/Taskfile.uv.yml`      |

## Boundaries

- Stack guidance (Go, uv) is advisory — surfaced in references, never enforced by lint.
- The skill does not opinionate on CI platform (`.github/workflows/*.yml`).
- `ark taskfile lint` V1 is structural-only; per-stack rules are deferred to v2.
