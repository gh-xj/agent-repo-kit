---
name: convention-engineering
description: Use when auditing or bootstrapping repo conventions (agent docs, tickets, wiki, verification gates, stack profiles). Triggers on "init tickets", "new ticket", "audit repo", "adopt wiki".
---

# convention-engineering (Claude Code adapter)

This file is the Claude Code thin wrapper for `agent-repo-kit`'s
harness-free content. It exists so Claude's skill loader can discover
the convention surface via the frontmatter above.

**All substantive content lives in `../../convention-engineering/`.** Do not
duplicate rules, templates, or scripts into this adapter.

The companion evaluator surface lives in `../../convention-evaluator/`.

## How install.sh wires this up

`install.sh` with `--target claude-code` symlinks:

- `convention-engineering/` → `~/.claude/skills/convention-engineering/`
- `convention-evaluator/` → `~/.claude/skills/convention-evaluator/`

so this adapter file is only needed while iterating in-tree. The
published form is the symlinked repo-root convention surface itself.
