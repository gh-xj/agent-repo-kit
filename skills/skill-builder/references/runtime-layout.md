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
