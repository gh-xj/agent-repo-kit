# agent-repo-kit

[![CI](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/gh-xj/agent-repo-kit/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/gh-xj/agent-repo-kit/pulls)

## TL;DR

`agent-repo-kit` is a drop-in set of repo conventions and tooling for
AI-agent-assisted development. It is **harness-agnostic**: works equally well
with Claude Code, Codex, Cursor, or a plain editor with no harness at all. The
kit gives any repo three things out of the box:

1. A flat-file **work tracker** (`.tickets/`) — state machine, verb surface,
   Taskfile.
2. An LLM-maintained **knowledge base** (`.wiki/`) — page types, frontmatter,
   citation rules, lint.
3. An **audit / bootstrap workflow** that scores a repo against the contract
   and guides adoption.

## Install

```bash
git clone https://github.com/YOURORG/agent-repo-kit.git
cd agent-repo-kit
./install.sh                       # auto-detect harness
./install.sh --target claude-code  # force Claude Code adapter
./install.sh --target codex        # force Codex adapter
./install.sh --target none         # just print adoption instructions
./install.sh --dry-run             # preview without changes
```

## What you get

- **`contract/`** — repo conventions: agent docs, docs taxonomy, stack
  profiles, verification gates, tickets + wiki scaffolds. The canonical,
  harness-free source of truth.
- **`evaluator/`** — skeptical scoring of a repo's adoption of the contract.
  Produces a graded report with evidence.
- **`examples/demo-repo/`** — a working repo showing `.tickets/` + `.wiki/`
  adoption end to end, wired to CI.
- **`adapters/`** — thin wrappers that expose `contract/` and `evaluator/` to
  a specific harness (Claude Code skill, Codex agent, Cursor rules).

## Quick example

```bash
# After installing the Claude Code adapter, in any repo:
cp -R .../agent-repo-kit/examples/demo-repo/.tickets ./
cp -R .../agent-repo-kit/examples/demo-repo/.wiki ./
task -d .tickets test     # 10/10 scenarios pass
task -d .wiki lint        # OK
```

Then paste the `## Conventions` block from
`examples/demo-repo/AGENTS.md` into your repo's `AGENTS.md` and `CLAUDE.md`.

## Architecture

```
           +-----------+        +-------------+
           | contract/ |<-------|  evaluator/ |
           +-----+-----+        +------+------+
                 ^                     ^
                 |                     |
         +-------+-----+       +-------+------+
         |  adapters/  |       | examples/    |
         | claude-code |       | demo-repo/   |
         |   codex     |       | (.tickets,   |
         |   cursor    |       |  .wiki, CI)  |
         +-------------+       +--------------+
```

Content lives in `contract/` and `evaluator/`. Adapters don't own content;
they re-export. Examples are concrete, runnable proof.

## Contributing

PRs welcome. Keep `contract/` and `evaluator/` harness-free; put harness
specifics under `adapters/<name>/`.

## License

MIT — see `LICENSE`.
