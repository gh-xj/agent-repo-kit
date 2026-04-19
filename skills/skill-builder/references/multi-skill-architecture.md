---
tags: [multi-skill, architecture, cross-repo, canonical-core]
---

# Multi-Skill Architecture

When two skills cover the same domain from different angles (e.g., build-time vs run-time, architecture vs investigation):

## When to Split vs Merge

| Signal | Action |
|--------|--------|
| Combined content > 500-line SKILL.md | Keep separate |
| Different user intents (build vs debug) | Keep separate |
| Shared concepts (pipeline, severity) | Extract canonical core |
| Single concern, just too long | Split into references, keep one skill |

## Canonical Core Pattern

Extract shared concepts into a `core/` directory in one skill (the "upstream"). The other skill back-references it.

```
upstream-skill/              (global, ~/.claude/skills/)
├── SKILL.md
└── references/
    ├── core/                # Shared canonical concepts
    │   ├── pipeline.md
    │   └── severity.md
    └── architecture/        # Skill-specific content

downstream-skill/            (repo-local, .claude/skills/)
├── SKILL.md                 # Declares dependency, has back-references
└── references/
    └── workflows/
```

## Dependency Rules

1. **One-way only** — downstream depends on upstream, never circular
2. **Declare explicitly** in downstream SKILL.md: "This skill depends on `upstream-skill` for X, Y, Z"
3. **Convention-based references** — Claude can read global skill files at runtime: `See upstream-skill > references/core/pipeline.md`
4. **Upstream never references downstream** — it can mention the downstream skill exists for routing purposes, but doesn't depend on it

## When to Extract to Core

A concept belongs in `core/` when:
- Referenced by both skills
- Has a single correct definition (not opinion/workflow)
- Changes to it should propagate to both skills automatically

## Source File Indexes

For architecture skills that describe code behavior, add a `## Source Files` table to each reference file mapping concepts to exact code locations:

```markdown
## Source Files

| File | Repo | Key Functions |
|------|------|---------------|
| `service/orchestrator.go:151` | my-service | `RunPipeline` — main orchestrator |
| `operator/analyzer.go:199` | my-service | `Analyze` — analysis entry point |
| `dal/kv/writer.go:48` | my-service | `Persist` — storage write |
```

**Why**: Bridges the gap between "the skill knows about it" and "the agent can find it in code." Without this, agents know the concept but waste tokens searching for the implementation.

**Maintenance**: Line numbers drift as code changes. Mitigate with self-healing instructions (see below). The function name is the stable anchor — line numbers are helpful hints.

## Versioned Detection Config

When skills reference external system configurations (rule engines, LLM prompts, policy definitions), store date-stamped snapshots in the repo:

```
detection-config/
├── dolphin-rules/2026-02-04/    # Rule engine export
│   ├── rules_ai.json
│   └── rules_malware.json
└── fornax-prompts/2026-02-05/   # LLM prompt export
    ├── screenshot_analysis.md
    └── dom_analysis.md
```

**Why**: Git diff between snapshots shows exactly what changed. Skills reference the snapshot path so agents can read the actual rule/prompt content during investigation.

**Convention**: New exports get a new date directory alongside old ones. Never overwrite — accumulate for diffing.

## Self-Healing Instructions

For skills that describe living systems, add maintenance instructions directly in the SKILL.md:

```markdown
## Knowledge Freshness

### Before trusting: verify critical claims
When this skill's knowledge drives a decision, verify against live code.

### After discovering: update this skill
If you find something that contradicts or extends this skill:
1. Read the relevant reference file
2. Update it with new information
3. Add a changelog entry

### Alert the user on drift
If you find a discrepancy: tell the user, then update the skill.
```

**Why**: Every investigation session becomes a self-healing opportunity. The skill improves with use rather than rotting.
