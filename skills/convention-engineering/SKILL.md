---
name: convention-engineering
description: "Use when designing or auditing repo conventions: `.conventions.yaml` descriptor, agent contract files (CLAUDE.md / AGENTS.md), docs taxonomy, verification gates, repo-local skill placement, `.work/` layout, `.wiki/` rules. Stack-agnostic. Skip for one-off product naming or domain architecture questions where no convention surface is being changed."
---

# Convention Engineering

Repo conventions for AI-agent-operated codebases. Stack-agnostic, descriptor-driven.

The pattern: a single `.conventions.yaml` at the repo root declares which conventions the repo opts into. An agent reads it, scaffolds or audits against it, and the same file feeds `task verify`.

## Routing

| Question                                            | Reference                                     |
| --------------------------------------------------- | --------------------------------------------- |
| What invariants does this skill optimise for?       | `references/core/agent-first-principles.md`   |
| What goes in `.conventions.yaml`?                   | `references/core/conventions-yaml.md`         |
| How should `CLAUDE.md` / `AGENTS.md` be structured? | `references/core/agent-knowledge.md`          |
| How to design the docs taxonomy?                    | `references/core/docs-taxonomy.md`            |
| Where do repo-local skills live?                    | `references/core/project-skill-placement.md`  |
| How should the verification gate work?              | `references/core/verification-gates.md`       |
| Bootstrap a new repo                                | `references/operations/bootstrap-workflow.md` |
| Audit an existing repo                              | `references/operations/audit-workflow.md`     |
| Adopt the work tracker (`.work/`)                   | `references/operations/work.md`               |
| Adopt the wiki (`.wiki/`)                           | `references/operations/wiki.md`               |

## Quick Start

1. Read or create `.conventions.yaml` at the repo root.
2. For a new repo: `references/operations/bootstrap-workflow.md`.
3. For an existing repo: `references/operations/audit-workflow.md`.

## Boundaries

- This skill prescribes structure, not stack-specific tooling. Linters, type checkers, supply-chain policies are the repo's choice; declare them in the descriptor and run them via `task verify`.
- The audit produces a gap report. No auto-fixes without user approval.
- `convention-evaluator` (skeptical scoring) is a separate sibling skill, not loaded automatically.
