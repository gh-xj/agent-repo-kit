# CLI Surfaces

Use this file when a workflow is too deterministic, too important, or too reusable to stay as prose inside a skill.

## Choose The Right Surface

| Surface                           | Best For                                                                             |
| --------------------------------- | ------------------------------------------------------------------------------------ |
| `SKILL.md`                        | routing, ownership, boundaries, high-level workflow                                  |
| `references/`                     | detailed but mostly read-only knowledge                                              |
| skill-local `scripts/`            | deterministic operations used mainly by that skill                                   |
| skill-local CLI in `cli/`         | one skill's durable commands, local verification, small command lifecycle            |
| repo-owned CLI in `tools/<name>/` | shared repo operations with policy, verification, traces, or multi-command lifecycle |

Bootstrap skill-local CLIs with whatever Go scaffolder your repo already uses. Keep the root `Taskfile.yml` as a thin wrapper over `go run ./...` or `task -d ... run -- ...`; do not reimplement command semantics in shell blocks.

Use `cli/` when most of these are true:

- the logic belongs to one skill boundary
- the command set is small and directly expresses that skill's contract
- verification should live next to the skill, not in a distant repo tool
- the root Taskfile only needs thin wrappers for human convenience

Use `tools/<name>/` when most of these are true:

- the workflow is stable across sessions
- multiple commands belong to one domain
- the repo needs a durable operating surface, not just a helper script
- machine-readable output matters
- verification or safety rules are part of the contract
- other skills, Taskfiles, or humans will call the same tool

## Design Rules

- Prefer Go for repo-owned CLIs; follow whatever Go style contract your repo already uses.
- Prefer Go for skill-local CLIs too; bootstrap with whatever scaffolder your repo already uses unless there is a stronger local convention.
- Keep the root `Taskfile.yml` thin and delegate into `tools/<name>/`.
- If exposing a skill-local CLI from the root, keep wrappers thin there too.
- Put domain semantics in the CLI, not in long Taskfile shell blocks.
- Provide read-only inspection commands before destructive commands.
- Emit machine-readable output for automation and stable human output for operators.
- Report artifact paths, trace ids, workflow urls, or log locations explicitly.
- Keep tests and a single verification gate close to the tool.

## Good Command Shapes

| Command Type                  | Purpose                                                    |
| ----------------------------- | ---------------------------------------------------------- |
| `check` or `status`           | read-only inspection                                       |
| `doctor`                      | environment and prerequisite diagnosis                     |
| `verify`                      | explicit verification surface                              |
| `apply`, `publish`, `release` | destructive or state-changing operation with strong guards |
| `trace` or `debug`            | inspect prior runs and artifacts                           |

## Representative Shape

A well-scoped repo-owned CLI tends to look like:

- One tool per stable lifecycle (e.g. a release-gate, deploy-gate, or migration CLI).
- The root `Taskfile.yml` delegates into the tool instead of reimplementing its logic inline.
- The CLI emits artifact paths, trace ids, or log locations for humans to follow.
- Structured `--json` output is available on every command, not just `version`.
- A `doctor` command verifies prerequisites (`gh`, `git`, `task`, auth, expected files).
- A single `verify` command is the canonical gate for both local and CI runs.

## Promotion Rule

If a skill keeps re-explaining a repo workflow that already has stable invariants, promote that workflow into `tools/<name>/` and leave the skill responsible for:

- when to use the CLI
- how to choose the right command
- what ownership and verification rules apply

If a workflow is stable but still owned by one skill, promote it into `cli/` first. Promote again into `tools/<name>/` only when multiple skills or humans need one repo-wide operating surface.
