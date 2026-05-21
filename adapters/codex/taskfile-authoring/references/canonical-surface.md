# Canonical Surface

The aggregate `fmt → lint → test → build → ci → verify` shape. Advisory, not
lint-enforced. Use as a sizing nudge when scaffolding a new Taskfile or
consolidating an overgrown one.

## The Shape

```
fmt
fmt:check         (optional; gate for lint/test/build in strict repos)
lint      deps: [fmt:check]
test      deps: [fmt:check]
build     deps: [fmt:check]
smoke     deps: [build]             (optional; run built artifact)
ci        deps: [lint, test, build, smoke]
verify    deps: [ci]                (plus repo-specific extras)
```

`ci` is the single contract CI runs. `verify` wraps `ci` and chains any
repo-specific gates — work status check, wiki freshness lint, generated
artifact audit. Keeping them separate means extras stay out of the CI hot
path, and local contributors still have one command.

## Why `ci` and `verify` Are Distinct

Without the split, every time a repo adds a local-only check (e.g. "warn if
`.work/` has stale items"), one of two bad things happens:

- The check goes into `ci` and slows down every PR for a rule that has
  nothing to do with shipping correctness.
- The check lives in a separate task that contributors forget to run.

With the split, `ci` is the shipping contract and `verify` is the
local-trust contract. They never diverge — `verify` always calls `ci` first.

## Example Dependency Graph

```
         fmt:check
         /   |   \
       lint test build
                    \
                     smoke
                      |
                      ci
                      |
                      verify
```

## Sizing Guidance (Not Lint-Enforced)

Rough targets for user-facing task counts (tasks that show up in `task --list`):

| Stack   | Aim | Hard ceiling |
| ------- | --- | ------------ |
| Generic | ≤ 8 | 12           |
| uv      | ≤ 5 | 8            |
| Go CLI  | ≤ 8 | 10           |

If you are past the ceiling, the usual causes are:

- Tasks that do one shell command and wrap nothing (`task install` → just `go install`).
- Per-subdir tasks that should be loops or single commands (`task test:pkg1`, `task test:pkg2` → `go test ./...`).
- Tasks that paper over missing composition (three copies of `test:x`, `test:y`, `test:z` → one included sub-Taskfile).

## `desc:` Is Mandatory For User-Facing Tasks

Any task that appears in `task --list` should have a `desc:`. Internal
helpers can set `internal: true` to hide them. Lint does **not** enforce
this in V1 (high false-positive rate on legitimate private tasks) — see
`lint-rules.md` deferred section.

## Non-Negotiables

- `ci` must not branch on environment (no `if [ "$CI" = "true" ]` logic inside the task).
- `verify` depends on `ci`; `ci` does not depend on `verify`.
- The task names `ci`, `verify`, `fmt`, `lint`, `test`, `build` are reserved for this shape — do not redefine them to mean something else.
