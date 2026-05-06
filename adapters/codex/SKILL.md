---
name: convention-engineering
version: 2.1.0
description: "Use when designing or auditing repo conventions: `.conventions.yaml` descriptor, agent contract files (CLAUDE.md / AGENTS.md), docs taxonomy, verification gates, repo-local skill placement, `.work/` layout, `.wiki/` rules. Pick an archetype (whole-repo shape: epic-wrapper for multi-repo products, brain for personal/knowledge-base repos) and stack recipes (core-beliefs, docs-taxonomy, agent-knowledge, work, wiki, project-skill-placement, verify-script). Stack-agnostic. Skip for one-off product naming or domain architecture questions where no convention surface is being changed."
---

# Convention Engineering

Repo conventions for AI-agent-operated codebases. Stack-agnostic, descriptor-driven.

The pattern: a single `.conventions.yaml` at the repo root declares which conventions the repo opts into. An agent reads it, scaffolds or audits against it, and the same file feeds `task verify`.

## Three building blocks

- **Concepts** — principles and the descriptor reference. Always-on.
- **Archetypes** — whole-repo shapes; pick one. (`epic-wrapper`, `brain`.)
- **Recipes** — stackable conventions; layer many. (`core-beliefs`, `agent-knowledge`, `docs-taxonomy`, `project-skill-placement`, `verify-script`, `work`, `wiki`.)

A repo declares its archetype + recipes in `.conventions.yaml`, then runs `task verify` against the resulting opt-ins.

## Routing

### Concepts

| Question                                            | Reference                                          |
| --------------------------------------------------- | -------------------------------------------------- |
| What invariants does this skill optimise for?       | `references/concepts/agent-first-principles.md`    |
| What goes in `.conventions.yaml`?                   | `references/concepts/conventions-yaml.md`          |
| How should the verify gate work?                    | `references/concepts/verification.md`              |

### Archetypes

| Whole-repo shape                                    | Reference                                          |
| --------------------------------------------------- | -------------------------------------------------- |
| Multi-repo product with `repo/<leaf>` symlinks      | `references/archetypes/epic-wrapper/pattern.md`    |
| Personal/knowledge-base brain repo                  | `references/archetypes/brain/pattern.md`           |
| Adding a new archetype — format spec                | `references/archetypes/_format.md`                 |

### Recipes

| Stackable convention                                | Reference                                          |
| --------------------------------------------------- | -------------------------------------------------- |
| Core beliefs / invariants / anti-patterns trio      | `references/recipes/core-beliefs/pattern.md`       |
| `CLAUDE.md` / `AGENTS.md` structure                 | `references/recipes/agent-knowledge/pattern.md`    |
| Docs taxonomy (`docs/{requests,planning,plans,…}`)  | `references/recipes/docs-taxonomy/pattern.md`      |
| Repo-local skill placement                          | `references/recipes/project-skill-placement/pattern.md` |
| `scripts/verify.sh` shape                           | `references/recipes/verify-script/pattern.md`      |
| `.work/` tracker adoption                           | `references/recipes/work/pattern.md`               |
| `.wiki/` knowledge base adoption                    | `references/recipes/wiki/pattern.md`               |
| Adding a new recipe — format spec                   | `references/recipes/_format.md`                    |

### Meta (about the skill)

| Concern                                             | Reference                                          |
| --------------------------------------------------- | -------------------------------------------------- |
| Bootstrap a new repo (generic flow)                 | `references/meta/bootstrap-workflow.md`            |
| Audit an existing repo (generic flow)               | `references/meta/audit-workflow.md`                |
| Score a repo against its declared conventions       | `references/meta/evaluation-rubric.md`             |
| Evaluation report layout                            | `references/meta/evaluation-report-template.md`    |
| Adapter contract for harness mirroring              | `references/meta/adapter-contract.md`              |

## Quick Start

1. Read or create `.conventions.yaml` at the repo root.
2. For a new repo: `references/meta/bootstrap-workflow.md`.
3. For an existing repo: `references/meta/audit-workflow.md`.

## Boundaries

- This skill prescribes structure, not stack-specific tooling. Linters, type checkers, supply-chain policies are the repo's choice; declare them in the descriptor and run them via `task verify`.
- The audit produces a gap report ("X is missing"); the evaluation rubric produces a graded judgment ("legibility 3/4, here is why"). Both live in `references/meta/`.
- No auto-fixes without user approval.
- For path migrations from the 1.x layout, see `MIGRATION.md`.
