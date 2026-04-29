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

1. A local-first **work tracker** (`.work/`) — inbox, triage, views, and a
   JSON-native `work` CLI.
2. An **audit / bootstrap workflow** that scores a repo against the contract
   and guides adoption.

Optional packs, such as `.wiki/`, are available when a repo has source-backed
knowledge that earns the extra surface area.

## Install

Install the skills:

```bash
npx skills add gh-xj/agent-repo-kit -g -a claude-code -a codex --skill '*' -y
```

This step requires Node.js so `npx` is available.

Install the shipped binaries (prebuilt by default, `~/.local/bin`):

```bash
curl -sSL https://raw.githubusercontent.com/gh-xj/agent-repo-kit/main/install.sh | sh
```

`npx skills` installs the canonical repo skills under the supported agent
runtime roots. `install.sh` installs shipped binaries such as `ark` and
`work` into the selected prefix when they are available in the release archive.

### From source

If you want to build from a local clone (or Go is available and you'd
rather not pull a prebuilt):

```bash
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
./install.sh --from-source
```

`--from-source` forces local `go build` of the `ark` and `work` entrypoints
into the install prefix instead of downloading a release archive. Requires
Go ≥ 1.25.

### Prefix and PATH

The default prefix is `~/.local/bin`. For a system-wide install use
`--prefix /usr/local/bin` (you may need `sudo` for writes there). Make
sure the chosen prefix is on your `PATH`.

Other useful flags:

```bash
./install.sh --prefix /usr/local/bin
./install.sh --from-source
./install.sh --dry-run              # preview without changes
```

### Upgrade

```bash
ark upgrade
```

`ark upgrade` upgrades the shipped `ark` and `work` binaries. If `ark`
lives inside a git clone it runs `git pull` + rebuild; otherwise it downloads
the latest release archive and replaces the installed binaries in place.

Refresh installed skills separately with `npx skills update -g`, or re-run
the `npx skills add gh-xj/agent-repo-kit ...` command above.

### Supported harnesses

- **claude-code** — install with
  `npx skills add gh-xj/agent-repo-kit -g -a claude-code --skill '*' -y`
- **codex** — install with
  `npx skills add gh-xj/agent-repo-kit -g -a codex --skill '*' -y`

`adapters/manifest.json` and `ark adapters link` remain available as a
low-level compatibility path for local or legacy symlink workflows. They
are no longer the recommended end-user install path.

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

Use the `ark` CLI to scaffold the kit-owned repo surface in one pass:

```bash
ark init \
  --repo-root /path/to/your-repo \
  --profiles go,typescript-react
```

This writes a tracked `.convention-engineering.json`, `docs/` taxonomy
READMEs, `.work/`, repo-local convention task wiring under
`.convention-engineering/`, and mirrored `AGENTS.md` / `CLAUDE.md`
convention blocks. The generated repo then supports:

```bash
task verify
```

Prerequisites:

- `ark` and `work` on `PATH` (installed by `./install.sh` or built from source) for bootstrap, convention checks, and work views
- `task`, `bash`, and standard Unix tools for `task verify`

## What you get

- **`skills/`** — canonical, harness-free skill sources:
  - `skills/convention-engineering/` — repo conventions: agent docs, docs
    taxonomy, stack profiles, verification gates, work scaffolds, and optional
    wiki scaffolds.
  - `skills/convention-evaluator/` — skeptical scoring of a repo's adoption
    of the contract. Produces a graded report with evidence.
  - `skills/skill-builder/` — skill for creating, refactoring, and auditing
    agent skills (trigger wording, portable structure, reference
    extraction, runtime placement).
  - `skills/taskfile-authoring/` — skill for writing canonical Taskfiles,
    used by `ark taskfile lint`.
  - `skills/attack-architecture/` — adversarial architecture-review skill
    (parallel lens attacks + debate).
  - `skills/harness-router/` — proposal-only router for deciding where
    session learnings and harness improvements should persist across
    instructions, skills, docs, work records, memory, and verification
    surfaces.
- **`examples/demo-repo/`** — a working repo showing lean `.work/` adoption
  end to end, wired to CI.
- **`adapters/`** — thin wrappers that expose `skills/` to a specific
  harness. `claude-code/` and `codex/` are shipped as compatibility targets
  (see `adapters/manifest.json`); `cursor/` is placeholder docs today.

## Quick example

```bash
ark init \
  --repo-root /path/to/your-repo \
  --profiles go

cd /path/to/your-repo
task verify             # conventions + work
task work -- view ready # inspect ready work
```

## Architecture

```
     +---------+             +-----------------------+
     | skills/ |<------------| examples/demo-repo/   |
     +----+----+             | (.work, CI)           |
          ^                  +-----------------------+
          |
   +------+-------+
   |  adapters/   |
   | claude-code  |
   | codex        |
   | cursor*      |
   +--------------+
```

`*` placeholder adapter docs only; not an installable target today.

Content lives under `skills/`. Adapters don't own content; they re-export
via `adapters/manifest.json`. Examples are concrete, runnable proof.

## Contributing

PRs welcome. Keep everything under `skills/` harness-free; put harness
specifics under `adapters/<name>/`.

## License

MIT — see `LICENSE`.
