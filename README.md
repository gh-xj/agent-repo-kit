# agent-repo-kit

[![CI](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/gh-xj/agent-repo-kit/pulls)

## TL;DR

`agent-repo-kit` is a drop-in set of repo conventions and tooling for
AI-agent-assisted development. The convention surfaces are
**harness-agnostic** and can be adopted from any editor or agent runtime. This
repo currently ships one ready-to-install adapter (`claude-code`) plus manual
adoption guidance; `codex/` and `cursor/` are placeholder adapter docs, not
packaged installs. The kit gives any repo three things out of the box:

1. A flat-file **work tracker** (`.tickets/`) — state machine, verb surface,
   Taskfile.
2. An LLM-maintained **knowledge base** (`.wiki/`) — page types, frontmatter,
   citation rules, lint.
3. An **audit / bootstrap workflow** that scores a repo against the contract
   and guides adoption.

## Install

```bash
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
./install.sh                       # auto-detect Claude Code, else print manual instructions
./install.sh --target claude-code  # install the Claude Code surfaces
./install.sh --target none         # just print manual adoption instructions
./install.sh --dry-run             # preview without changes
./install.sh --skip-build          # skip the Go build of cli/bin/ark (requires Go otherwise)
```

Supported install targets today: `claude-code` and `none`. Codex and Cursor
can consume the generated repo surface manually, but this repo does not yet
ship installable adapters for them.

## Bootstrap A Repo

Use the `ark` CLI to scaffold the kit-owned repo surface in one pass:

```bash
ark init \
  --repo-root /path/to/your-repo \
  --profiles go,typescript-react
```

This writes a tracked `.convention-engineering.json`, `docs/` taxonomy
READMEs, `.tickets/`, `.wiki/`, repo-local convention task wiring under
`.convention-engineering/`, and mirrored `AGENTS.md` / `CLAUDE.md`
convention blocks. The generated repo then supports:

```bash
task verify
```

Prerequisites:

- `ark` on `PATH` (installed by `./install.sh`) for bootstrap and convention checks
- `task`, `bash`, and standard Unix tools for `task verify`

## What you get

- **`convention-engineering/`** — repo conventions: agent docs, docs
  taxonomy, stack profiles, verification gates, tickets + wiki scaffolds.
  The canonical, harness-free source of truth.
- **`convention-evaluator/`** — skeptical scoring of a repo's adoption of
  the contract. Produces a graded report with evidence.
- **`skill-builder/`** — harness-free skill for creating, refactoring, and
  auditing agent skills (trigger wording, portable structure, reference
  extraction, runtime placement).
- **`examples/demo-repo/`** — a working repo showing `.tickets/` + `.wiki/`
  adoption end to end, wired to CI.
- **`adapters/`** — thin wrappers that expose the repo-root skill surfaces
  to a specific harness. `claude-code/` is shipped; `codex/` and `cursor/`
  are placeholder docs today.

## Quick example

```bash
ark init \
  --repo-root /path/to/your-repo \
  --profiles go

cd /path/to/your-repo
task verify           # conventions + tickets + wiki
task -d .tickets test # 10/10 scenarios pass
task -d .wiki lint    # OK
```

## Architecture

```
     +-------------------------+      +-----------------------+
     | convention-engineering/ |<-----| convention-evaluator/ |
     +------------+------------+      +-----------+-----------+
                  ^                               ^
                  |                               |
           +------+-------+                +------+------+
           |  adapters/   |                | examples/   |
           | claude-code  |                | demo-repo/  |
           | codex*       |                | (.tickets,  |
           | cursor*      |                |  .wiki, CI) |
           +--------------+                +-------------+
```

`*` placeholder adapter docs only; not installable targets today.

Content lives in `convention-engineering/`, `convention-evaluator/`, and
`skill-builder/`. Adapters don't own content; they re-export. Examples are
concrete, runnable proof.

## Contributing

PRs welcome. Keep `convention-engineering/`, `convention-evaluator/`, and
`skill-builder/` harness-free; put harness specifics under
`adapters/<name>/`.

## License

MIT — see `LICENSE`.
