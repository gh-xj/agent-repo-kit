---
name: skill-builder
description: Use when creating, refactoring, auditing, or migrating Claude/Codex skills, especially when trigger wording, portable structure, reference extraction, or runtime placement need design.
---

# Skill Builder

Design and maintain Claude/Codex skills as small, portable routers with clear triggers and minimal duplication.

## Use This For

- Creating a new skill.
- Refactoring an overgrown `SKILL.md`.
- Auditing trigger wording, layout, or duplication.
- Migrating capability into or out of a skill layer.
- Deciding when stable workflow logic belongs in `scripts/`, a skill-local CLI in `cli/`, or a repo-owned CLI under `tools/`.

## First Pass

1. Classify the request: `create`, `update`, `audit`, or `migrate`.
2. Run the confidence check: `confident`, `partial`, or `not ready`.
3. Choose scope: global or project-local, Claude or Codex or both.
4. Choose the operating surface:
   - skill prose
   - extracted reference
   - skill-local script
   - skill-local CLI in `cli/`
   - repo-owned CLI under `tools/`

## Router Rule

- Keep `SKILL.md` as the entrypoint and routing table, not the whole knowledge base.
- Prefer `SKILL.md` under 200 lines. Above 400 lines means refactor now.
- Keep the portable core to `name` + `description`.
- Put runtime-specific metadata beside the portable core, not inside it.
- Keep one source of truth for each fact or rule.
- Move deterministic, repeated logic into code once the pattern is verified.

## Mode Selection

| Mode      | Use When                                                   | Primary Output                                     |
| --------- | ---------------------------------------------------------- | -------------------------------------------------- |
| `create`  | No skill exists yet                                        | New `SKILL.md` plus any references/scripts/assets  |
| `update`  | Skill exists but needs new capability or cleanup           | Smaller or clearer skill with preserved behavior   |
| `audit`   | Triggering or structure feels wrong                        | Findings plus targeted refactor plan               |
| `migrate` | Capability is moving across docs, skills, scripts, or CLIs | Updated owner surface plus stale-reference cleanup |

See `references/workflows.md` for the concrete workflow for each mode.

## Confidence Gate

| Confidence  | Action                                  |
| ----------- | --------------------------------------- |
| `confident` | Build the full skill                    |
| `partial`   | Ship `v0.x` with `## Gaps`              |
| `not ready` | Capture notes only; do not over-specify |

If the domain is only partially understood, prefer the Q&A-driven path in `references/workflows.md`.

## Non-Negotiables

- The description must say when to use the skill, not summarize the workflow.
- Codex trigger quality depends heavily on `description`; make trigger phrases explicit.
- Do not duplicate the same rules across `SKILL.md` and references.
- Keep repo-specific operating rules out of the portable core when a local reference or wrapper skill is enough.
- If capability moves into a skill layer, remove stale docs, workflows, or Task targets in the same change.
- If repo behavior changes, name the verification gate explicitly.
- If a skill-local CLI is warranted, bootstrap it with your Go CLI scaffolder of choice and keep Task wrappers thin.

## Choosing Code Instead Of Prose

Use prose when judgment is required. Use code when the pattern is deterministic and repeated.

- Repeated logic inside a skill: see `references/pattern-to-script.md`
- Skill-local operations with one skill boundary and a small command surface: see `references/repo-owned-clis.md`
- Shared repo operations with policy or verification requirements: see `references/repo-owned-clis.md`

## CLI Surface

- `skill-builder init` scaffolds a router-grade `SKILL.md` and can optionally add a `cli/` surface.
- `skill-builder audit` checks frontmatter, router size, and referenced relative paths.
- When wired into a host repo's root `Taskfile.yml`, keep the wrappers thin:
  - `task skill-builder:init -- ...`
  - `task skill-builder:audit -- ...`
  - `task skill-builder:run -- ...`

## References

| File                                     | Use For                                                          |
| ---------------------------------------- | ---------------------------------------------------------------- |
| `references/runtime-layout.md`           | Runtime roots, portable core, loading model, placement decisions |
| `references/workflows.md`                | `create`, `update`, `audit`, and `migrate` workflows             |
| `references/repo-owned-clis.md`          | When stable logic should move into `cli/` or `tools/<name>/`     |
| `references/pattern-to-script.md`        | When stable logic should move into skill-local scripts           |
| `references/multi-skill-architecture.md` | Cross-skill systems and shared-core patterns                     |
| `references/superpowers-patterns.md`     | Hard gates, evidence-before-claims, and agent-drift controls     |
| `references/maintenance.md`              | Troubleshooting, parity checks, and health indicators            |

## Boundaries

- This skill designs and refactors skills; it does not make every repeated workflow a skill automatically.
- Do not keep repo-operating logic in the portable core when a repo-local reference or CLI is the correct owner.
- For destructive migrations or unclear ownership changes, confirm scope before rewriting multiple surfaces.
