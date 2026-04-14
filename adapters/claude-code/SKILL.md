---
name: convention-engineering
description: Use when auditing or bootstrapping repo conventions (agent docs, tickets, wiki, verification gates, stack profiles). Triggers on "init tickets", "new ticket", "audit repo", "adopt wiki".
---

# convention-engineering (Claude Code adapter)

This file is the Claude Code thin wrapper for `agent-repo-kit`'s
harness-free content. It exists so Claude's skill loader can discover
the convention surface via the frontmatter above.

**All substantive content lives in `../../contract/`.** Do not duplicate
rules, templates, or scripts into this adapter — read from `contract/`.

For the evaluator, see the sibling adapter file that points at
`../../evaluator/` (or symlink it into `~/.claude/skills/convention-evaluator/`
via the repo's `install.sh`).

## How install.sh wires this up

`install.sh` with `--target claude-code` symlinks:

- `contract/` → `~/.claude/skills/convention-engineering/`
- `evaluator/` → `~/.claude/skills/convention-evaluator/`

so this adapter file is only needed while iterating in-tree. The
published form is the symlinked `contract/` itself.
