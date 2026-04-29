---
name: convention-engineering
description: "Use when auditing or bootstrapping repo conventions: agent docs, docs taxonomy, verification gates, layer boundaries, dependency hygiene, repo-local skill placement, work tracker (.work/), or optional wiki (.wiki/). Triggers: init work, new work item, work inbox, work triage, work status, work view, init wiki, ingest source, wiki lint, knowledge base, scaffold repo conventions."
---

<!-- agent-repo-kit:skill-sync — do not edit; regenerate with `ark skill sync` -->

# Convention Engineering

Unified repo convention knowledge with multi-stack profiles. Audit existing repos, bootstrap new ones, maintain agent-first legibility.

Pillars: context engineering + architectural constraints + garbage collection.

## Scope

Owned:

- Universal repo convention patterns (agent docs, architecture contracts, verification gates, supply chain).
- Repo-local skill placement policy for AI agent runtimes.
- Stack-specific profiles (Go, TypeScript/React, Python, Research Corpus) that compose additively.
- Audit workflow (check existing repo against conventions, produce gap report).
- Bootstrap workflow (scaffold convention files for new repo).
- Open-source/local-overlay workflow (apply conventions with a local-only git exclude list and `.docs/` without touching tracked files).
- Verification gate contracts.

Not owned:

- Domain-specific architecture knowledge (stays in domain-specific references).
- Build system internals (repo-specific).
- Product behavior decisions.

## Routing Table

### Core Patterns

| Question                                                             | Reference                                          |
| -------------------------------------------------------------------- | -------------------------------------------------- |
| What invariants does this skill optimize for? (north star)           | `references/core/agent-first-principles.md`        |
| How to structure agent docs (CLAUDE.md, AGENTS.md, ARCHITECTURE.md)? | `references/core/agent-knowledge.md`               |
| How to design docs taxonomy for RFI/design/plan?                     | `references/core/docs-taxonomy.md`                 |
| How should repo-local Claude/Codex skills be placed?                 | `references/core/project-skill-placement.md`       |
| How to enforce layer boundaries?                                     | `references/core/architecture-contracts.md`        |
| How to design verification gates?                                    | `references/core/verification-gates.md`            |
| How should lint/CLI error messages be written?                       | `references/core/error-messages-as-remediation.md` |
| Dependency hygiene for my stack?                                     | `references/core/supply-chain.md`                  |

### Stack Profiles

| Question                           | Reference                                 |
| ---------------------------------- | ----------------------------------------- |
| Go repo conventions?               | `references/profiles/go.md`               |
| TypeScript/React repo conventions? | `references/profiles/typescript-react.md` |
| Python repo conventions?           | `references/profiles/python.md`           |
| Research corpus repo conventions?  | `references/profiles/research-corpus.md`  |

Profiles compose: a Go+React repo uses both `references/profiles/go.md` and `references/profiles/typescript-react.md`. A research repo with Go CLI tooling uses both `references/profiles/research-corpus.md` and `references/profiles/go.md`.

### Operations

| Question                                                                     | Reference                                                   |
| ---------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Audit a repo against conventions                                             | `references/operations/audit-workflow.md`                   |
| Bootstrap a new repo                                                         | `references/operations/bootstrap-workflow.md`               |
| Apply conventions to an open-source repo using local `git exclude` + `.docs` | `references/operations/open-source-git-exclude-workflow.md` |
| Scaffold or operate a work tracker (`.work/`)                                | `references/operations/work.md`                             |
| Scaffold or operate an LLM-maintained wiki (`.wiki/`)                        | `references/operations/wiki.md`                             |

### Companion Surfaces

| Question                                                                       | Surface                |
| ------------------------------------------------------------------------------ | ---------------------- |
| How do I score convention quality or interpret contract semantics skeptically? | `convention-evaluator` |

#### Claude Code adapter (harness-specific)

When loaded as a Claude Code skill, this convention pairs with a separate
skill-authoring surface for `SKILL.md` scaffolding, references layout, and
runtime metadata. If using Claude Code, see the `skill-builder` skill for that
work. In other runtimes, hand off to whatever skill-authoring tool that
runtime provides.

### Verification Contracts

| Question                          | Reference                                   |
| --------------------------------- | ------------------------------------------- |
| Gate/task topology                | `references/contracts/task-gates.md`        |
| Docs governance checks            | `references/contracts/docs-governance.md`   |
| Replay scenario gates             | `references/contracts/replay-regression.md` |
| Failure classification and triage | `references/contracts/failure-taxonomy.md`  |

### Configuration

| Question                               | Reference                                   |
| -------------------------------------- | ------------------------------------------- |
| Config schema for the contract checker | `references/config.schema.json`             |
| Example config file                    | `references/convention-config.example.json` |

## Quick Start

1. Read the routing table above, pick the reference for your question.
2. For repo audit: `references/operations/audit-workflow.md`.
3. For new repo bootstrap: `references/operations/bootstrap-workflow.md`.
4. For docs taxonomy decisions (RFI/design/plan): `references/core/docs-taxonomy.md`.
5. For repo-local skill placement: `references/core/project-skill-placement.md`.
6. For open-source/local-only setup (`git exclude` + `.docs`): `references/operations/open-source-git-exclude-workflow.md`.
7. Run the contract checker (assumes `ark` is on `PATH` after `install.sh`; set `ARK_BINARY` to override):

```bash
ark check --repo-root . --json
```

With repo config:

```bash
ark check --repo-root . --config .convention-engineering.json --json
```

To orchestrate evaluation artifacts and launch the evaluator:

```bash
ark orchestrate --repo-root . --topic convention-run
```

## Boundaries

- Audit produces gap reports, not auto-fixes (user must approve changes).
- Verification gate contracts are carried forward from a predecessor pattern as-is.
- Profiles are descriptive (patterns to follow), not prescriptive (exact configs to copy).

## Changelog

| Date       | Change                                                                                                                                                                                                                                                                                | File                                                                                                                                                                                                                                                                       |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-02-27 | v1.0: Created as replacement for a predecessor verification-harness pattern. Absorbed contracts, added core patterns, stack profiles, operations.                                                                                                                                     | all                                                                                                                                                                                                                                                                        |
| 2026-02-27 | v1.1: Added explicit canonical verify artifact/debug contract and dependency-preflight guidance (`bun`/frontend gates).                                                                                                                                                               | references/contracts/task-gates.md, references/core/verification-gates.md, references/contracts/failure-taxonomy.md                                                                                                                                                        |
| 2026-02-28 | v1.2: Added open-source `git exclude` + `.docs` workflow and checker support for `.git/info/exclude` pattern contracts.                                                                                                                                                               | SKILL.md, scripts/main.go, scripts/main_test.go, references/operations/open-source-git-exclude-workflow.md, references/config.schema.json, references/convention-config.example.json, references/operations/audit-workflow.md, references/operations/bootstrap-workflow.md |
| 2026-03-04 | v1.3: Added canonical docs taxonomy standard for request/design/plan lifecycle and integrated it into routing, audit, and bootstrap guidance.                                                                                                                                         | SKILL.md, references/core/docs-taxonomy.md, references/contracts/docs-governance.md, references/operations/audit-workflow.md, references/operations/bootstrap-workflow.md, README.md                                                                                       |
| 2026-03-11 | v1.4: Added research corpus profile covering lifecycle-layer structure, raw capture contracts, frontmatter validation, semantic gate patterns, run metadata, and INDEX freshness.                                                                                                     | SKILL.md, references/profiles/research-corpus.md                                                                                                                                                                                                                           |
| 2026-03-11 | v1.6: Quality pass — added TOCs to 6 files >100 lines, linked config schema/example from SKILL.md routing table, deduplicated depguard diagram in Go profile, added section index to main.go, fixed `required_taskfile_tasks` → `taskfile_checks` in research-corpus config template. | SKILL.md, references/profiles/research-corpus.md, references/profiles/go.md, references/profiles/python.md, references/operations/audit-workflow.md, references/core/architecture-contracts.md, references/core/verification-gates.md, scripts/main.go                     |
| 2026-04-03 | v1.8: Narrowed routing to repo conventions and repo-local skill placement policy; handed off skill authoring to a separate skill-authoring surface and skeptical scoring to the planned `convention-evaluator` sibling surface.                                                       | SKILL.md, references/core/project-skill-placement.md                                                                                                                                                                                                                       |
| 2026-04-03 | v1.9: Added the `convention-evaluator` sibling router covering contract semantics, evaluator-owned thresholds, and isolated evaluation artifacts.                                                                                                                                     | SKILL.md, ../convention-evaluator/SKILL.md                                                                                                                                                                                                                                 |
| 2026-04-03 | v1.10: Added orchestrated evaluation flow with convention briefs, handoff/receipt artifacts, raw evidence bundle capture, and package-directory CLI invocation for the multi-file script surface.                                                                                     | SKILL.md, README.md, scripts/main.go, scripts/orchestrator.go, scripts/orchestrator_test.go, scripts/main_test.go                                                                                                                                                          |
