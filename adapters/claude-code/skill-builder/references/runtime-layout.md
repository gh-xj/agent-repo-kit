# Runtime Layout

Use this file when the main question is where a skill belongs, what the portable core should contain, or how runtime-specific metadata should be isolated.

## Canonical Runtime Table

| Aspect | Claude Code | Codex |
| --- | --- | --- |
| Personal skills | `~/.claude/skills/<name>/` | `~/.agents/skills/<name>/` |
| Project skills | `.claude/skills/<name>/` | `.agents/skills/<name>/` |
| Global instructions | `~/.claude/CLAUDE.md` | `~/.codex/AGENTS.md` |
| Invocation | `Skill` tool or slash commands | Auto-discovery by `description` |
| Portable core | `SKILL.md`, `references/`, `scripts/`, `assets/` | `SKILL.md`, `references/`, `scripts/`, `assets/` |
| Optional runtime metadata | none required | `agents/openai.yaml` when UI metadata is needed |

If another file disagrees with this table, fix the other file.

## Portable Core

For managed installs, prefer the open skills CLI instead of hand-copying
files:

```bash
npx skills add <source> -g -a claude-code -a codex --skill '*' -y
```

Portable custom skills should default to:

```yaml
---
name: skill-name
description: Use when ...
---
```

Rules:

- Treat `name` + `description` as the shared default.
- Keep runtime-specific metadata outside the portable core unless it is truly runtime-specific.
- Prefer sibling files such as `agents/openai.yaml` over portable-frontmatter expansion.

## Loading Model

| Level | Content | When Loaded |
| --- | --- | --- |
| 1 | metadata (`name`, `description`) | always |
| 2 | `SKILL.md` body | when the skill triggers |
| 3 | references, scripts, assets | on demand |

Implications:

- `SKILL.md` should stay small enough to act as a router.
- Put deep knowledge in references.
- Put deterministic operations in code, not long prose.

## Placement Choice

| Scope | Place It Here |
| --- | --- |
| Personal reusable skill | `~/.claude/skills/` and/or `~/.agents/skills/` |
| Repo-specific workflow | `.claude/skills/` and/or `.agents/skills/` |
| Runtime-specific UI metadata | `agents/openai.yaml` or runtime-only sibling files |
| Repo automation logic | `tools/` or `scripts/`, not the portable skill core |

### Project-local vs global — promotion criteria

A skill should live at the global home (`~/.claude/skills/` and/or
`~/.agents/skills/`) instead of a project's `.claude/skills/` when **any one**
of these holds:

1. **Data scope is broader than the host repo**: the skill reads or writes
   files that live outside the hosting leaf (e.g., a search skill that
   targets `xj-private-finance` + `xj-private-info` + `dropbox-inbox` +
   `private-config`). The project where it was first written is incidental.
2. **The skill is stack-generic**: it wraps a tool bound to a local account
   or external service (Gmail, WeChat, browser), not to a specific repo's
   data shape.
3. **Multiple leaves want to invoke it**: even if only one currently does,
   anticipating cross-leaf use justifies promotion.

Worked example: `rga` (ripgrep-all over PDFs/JSONL/SQLite/docx across the
whole personal-data footprint) and `singlefile-clipper` (reads SingleFile
captures from `~/Dropbox/dropbox-inbox/evidence/`) were promoted out of
`xj-private-finance/.claude/skills/` to `~/.claude/skills/` in 2026-05-25
because both fail criterion 1: their target data is not finance-bound.

Don't promote prematurely. A skill that **could** be reused but currently
serves one project is fine project-local. Promote when the second project
asks for it, or when the data scope makes the leaf-local placement
actively misleading.

When promoting, update both source-repo and brain CLAUDE.md cross-references
so future readers know where the skill lives. See `audit-workflow.md`
Phase 3 for the mechanics.

## Dual-Runtime Mirrors

When a project-local skill must be discoverable by both Claude Code and Codex,
prefer one canonical portable core plus runtime discovery adapters.

Recommended pattern:

- Choose one source-of-truth skill directory for the portable core.
- Expose the other runtime root with a symlink or generated adapter when the
  environment supports it.
- Document which path is canonical and which path is a discovery mirror in the
  repo's agent instructions or convention contract.
- Audit the skill through every discovery path that agents are expected to use.

Avoid hand-maintained duplicate skill copies unless runtime behavior genuinely
differs. If copies are unavoidable, add a sync or drift check.

## Trigger Writing

Codex relies primarily on `description`, so write it as trigger guidance:

- Good: `Use when creating, auditing, or refactoring Claude/Codex skills.`
- Bad: `Tool for skill workflows and references and validation.`

Rules:

- Describe when to use the skill.
- Prefer trigger phrases over capability lists.
- Avoid summarizing the workflow inside `description`.

## Structure Guidelines

Recommended layout:

```text
skill-name/
├── SKILL.md
├── references/
├── scripts/
├── assets/
└── agents/
```

Keep references one or two levels deep from `SKILL.md`. For long reference files, add a short table of contents.
