# agent-repo-kit

[![CI](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/gh-xj/agent-repo-kit/pulls)

## TL;DR

`agent-repo-kit` is a drop-in set of repo conventions and tooling for
AI-agent-assisted development. The convention surfaces are
**harness-agnostic** and can be adopted from any editor or agent runtime.
This repo ships installable adapters for `claude-code` and `codex`;
`cursor/` remains placeholder adapter docs. The kit gives any repo three
things out of the box:

1. A flat-file **work tracker** (`.tickets/`) — state machine, verb surface,
   Taskfile.
2. An LLM-maintained **knowledge base** (`.wiki/`) — page types, frontmatter,
   citation rules, lint.
3. An **audit / bootstrap workflow** that scores a repo against the contract
   and guides adoption.

## Install

One-liner (default: Claude Code, prebuilt binary, `~/.local/bin`):

```bash
curl -sSL https://raw.githubusercontent.com/gh-xj/agent-repo-kit/main/install.sh | sh
```

This downloads the prebuilt `ark` binary for your OS/arch from the latest
GitHub Release, installs it to `~/.local/bin/ark`, then wires the skill
directories into your harness via `ark adapters link`.

### From source

If you want to build from a local clone (or Go is available and you'd
rather not pull a prebuilt):

```bash
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
./install.sh --from-source
```

`--from-source` forces a `go build` of `cli/` into the install prefix
instead of downloading a release archive. Requires Go ≥ 1.25.

### Prefix and PATH

The default prefix is `~/.local/bin`. For a system-wide install use
`--prefix /usr/local/bin` (you may need `sudo` for writes there). Make
sure the chosen prefix is on your `PATH`.

Other useful flags:

```bash
./install.sh --target claude-code   # override harness auto-detect
./install.sh --skip-symlinks        # install binary only; skip harness wiring
./install.sh --dry-run              # preview without changes
```

### Upgrade

```bash
ark upgrade
```

`ark upgrade` auto-detects how it was installed: if the binary lives
inside a git clone it runs `git pull` + rebuild; otherwise it downloads
the latest release archive and replaces itself in place.

### Supported harnesses

- **claude-code** — auto-detected when `~/.claude/skills` exists; symlinks
  the convention skills under that directory.
- **codex** — skill root `~/.codex/skills`; pass `--target codex` to
  select it explicitly.

`adapters/manifest.json` is the single source of truth for which skills
get linked into which harness. Pass `--target <name>` to `install.sh` or
`ark upgrade` to override auto-detection.

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
  to a specific harness. `claude-code/` and `codex/` are shipped as
  install targets (see `adapters/manifest.json`); `cursor/` is
  placeholder docs today.

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
           | codex        |                | (.tickets,  |
           | cursor*      |                |  .wiki, CI) |
           +--------------+                +-------------+
```

`*` placeholder adapter docs only; not an installable target today.

Content lives in `convention-engineering/`, `convention-evaluator/`, and
`skill-builder/`. Adapters don't own content; they re-export. Examples are
concrete, runnable proof.

## Contributing

PRs welcome. Keep `convention-engineering/`, `convention-evaluator/`, and
`skill-builder/` harness-free; put harness specifics under
`adapters/<name>/`.

## License

MIT — see `LICENSE`.
