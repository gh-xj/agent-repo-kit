---
name: skill-builder
description: Use when creating, refactoring, auditing, or migrating Claude/Codex skills, especially when trigger wording, portable structure, reference extraction, or runtime placement need design.
---

# skill-builder (Claude Code adapter)

This file is the Claude Code thin wrapper for `agent-repo-kit`'s
harness-free `skill-builder/` surface. It exists so Claude's skill loader
can discover the skill via the frontmatter above.

**All substantive content lives in `../../../skill-builder/`.** Do not
duplicate rules, templates, or scripts into this adapter.

## How install.sh wires this up

`install.sh` with `--target claude-code` symlinks:

- `skill-builder/` → `~/.claude/skills/skill-builder/`

so this adapter file is only needed while iterating in-tree. The
published form is the symlinked repo-root skill surface itself.
