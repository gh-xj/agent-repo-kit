# agent-repo-kit

[![CI](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/gh-xj/agent-repo-kit/pulls)

## TL;DR

`agent-repo-kit` is a drop-in set of repo conventions and tooling for
AI-agent-assisted development. The convention surfaces are
**harness-agnostic** and can be adopted from any editor or agent runtime.
This repo ships canonical open-skill surfaces plus thin compatibility
adapters for `claude-code` and `codex`; `cursor/` remains placeholder
adapter docs. The kit gives any repo a small default core:

1. A repo-root `.conventions.yaml` descriptor that an agent reads to scaffold
   and audit the rest.
2. A local-first **work tracker** (`.work/`) — inbox, triage, views, and a
   JSON-native `work` CLI.

Optional packs, such as `.wiki/`, are available when a repo has source-backed
knowledge that earns the extra surface area.

## Install

Install the skills:

```bash
npx skills add gh-xj/agent-repo-kit -g -a claude-code -a codex --skill '*' -y
```

This step requires Node.js so `npx` is available.

Build the `work` CLI from source:

```bash
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit/cli
go install ./cmd/work
```

Requires Go ≥ 1.25. The resulting `work` binary lives in `$(go env GOBIN)`
or `$GOPATH/bin`; ensure that directory is on your `PATH`.

### Supported harnesses

- **claude-code** — install with
  `npx skills add gh-xj/agent-repo-kit -g -a claude-code --skill '*' -y`
- **codex** — install with
  `npx skills add gh-xj/agent-repo-kit -g -a codex --skill '*' -y`

### Maintainer Setup

If you are actively editing this repo's `skills/` sources, prefer the local
symlink workflow over `npx skills add /path/to/repo ...`.

Why: installing from a local filesystem path with `npx skills add` copies the
skill directories into the managed runtime roots. That is useful for smoke
tests and release verification, but edits in your checkout do not live-update
the installed skills.

For day-to-day maintenance from a local clone, run:

```bash
task skills:link-dev
```

That helper:

- symlinks every `skills/*/SKILL.md` repo skill into `~/.agents/skills/`
- symlinks Claude entries from `~/.claude/skills/` to `~/.agents/skills/`
- does not modify `~/.codex/skills/`, which can keep runtime-owned entries
  such as `.system`

After linking, restart Codex and Claude Code so they rescan the skill roots.
Manual maintainer symlinks may not appear in `npx skills ls`; validate them
by checking the filesystem directly (for example
`~/.agents/skills/<name>/SKILL.md`).

## Bootstrap A Repo

Use the `convention-engineering` skill (loaded via the supported harness) to
scaffold the kit-owned repo surface. The agent reads or creates a
`.conventions.yaml` at your repo root and writes the artifacts it declares —
agent contract files (`CLAUDE.md` / `AGENTS.md`), docs taxonomy, Taskfile,
pre-commit hook — following
`skills/convention-engineering/references/operations/bootstrap-workflow.md`.

The generated repo then supports:

```bash
task verify
```

Prerequisites:

- `work` on `PATH` (built from source above) for the work tracker.
- `task`, `bash`, and standard Unix tools for `task verify`.

## What you get

- **`skills/`** — canonical, harness-free skill sources:
  - `skills/convention-engineering/` — repo conventions: `.conventions.yaml`,
    agent docs, docs taxonomy, verification gates, repo-local skill placement,
    and optional work / wiki scaffolds.
  - `skills/convention-evaluator/` — skeptical scoring of a repo's adoption
    of its declared conventions. Produces a graded report with evidence.
  - `skills/skill-builder/` — skill for creating, refactoring, and auditing
    agent skills.
  - `skills/taskfile-authoring/` — skill for writing canonical Taskfiles.
  - `skills/attack-architecture/` — adversarial architecture-review skill.
  - `skills/harness-router/` — proposal-only router for deciding where
    session learnings and harness improvements should persist.
  - `skills/work-cli/` — operating the `.work/` tracker.
  - `skills/paper-vetting/` — three-lens credibility vetting for research
    papers before reading them.
- **`adapters/`** — thin wrappers that expose `skills/` to a specific
  harness. `claude-code/` and `codex/` are shipped as compatibility targets;
  `cursor/` is placeholder docs today.
- **`cli/`** — the `work` CLI (Go) that powers `.work/`.

## Quick example

```bash
cd /path/to/your-repo
# Use the convention-engineering skill via your harness to bootstrap, then:
task verify             # runs whatever .conventions.yaml declares
task work -- view ready # inspect ready work
```

## Architecture

```
     +---------+
     | skills/ |
     +----+----+
          ^
   +------+-------+
   |  adapters/   |
   | claude-code  |
   | codex        |
   | cursor*      |
   +--------------+
```

`*` placeholder adapter docs only; not an installable target today.

Content lives under `skills/`. Adapters don't own content; they re-export
via `adapters/manifest.json`.

## Contributing

PRs welcome. Keep everything under `skills/` harness-free; put harness
specifics under `adapters/<name>/`.

## License

MIT — see `LICENSE`.
