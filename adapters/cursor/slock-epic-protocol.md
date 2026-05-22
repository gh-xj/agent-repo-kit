---
name: slock-epic-protocol
description: >-
  Use when designing, bootstrapping, auditing, repairing, or evolving a
  Slock-agent-to-epic-repo protocol: Slock agents mapped by
  slock_agent_registry to concrete leaf repos, generated pointer-only MEMORY.md
  files, generated repo symlinks, per-leaf tracked .work stores, repo-local
  skill mirrors, and cross-repo handoffs. Do not use for ordinary Slock
  chat/task commands or generic repo conventions without Slock agent mapping.
---

# Slock Epic Protocol

Maintain the bridge between Slock runtime agents and a federated epic repo. The
protocol is repo-owned; Slock is the wake-up, social, and task surface.

## Use This For

- Creating or changing `slock_agent_registry`.
- Mapping one Slock agent to one concrete leaf repo.
- Installing or checking generated Slock `repo` symlinks (always) and `epic`
  symlinks (when `epic_symlink_name` is set on the registry — see
  references/protocol.md).
- Keeping Slock `MEMORY.md` pointer-only.
- Migrating durable Slock notes into repo docs, skills, or `.work`.
- Adding repo-local Claude/Codex skill discovery for mapped agents.
- Designing cross-repo handoffs where target repo inbox/work is canonical.

## Do Not Use For

- Sending Slock messages, claiming Slock tasks, or reading Slock history for an
  ordinary work request.
- Generic epic-wrapper setup with no Slock agent map.
- Generic skill authoring. Use `skill-builder`.
- Generic `.work/` operation. Use `work-cli`.
- General repo convention bootstrap. Use `convention-engineering`.

## First Pass

1. Classify the task: `design`, `bootstrap`, `audit`, `repair`, or `migrate`.
2. Identify canonical surface:
   - single repo: repo descriptor
   - multi-repo product: epic descriptor
3. Inspect before editing:
   - descriptor map
   - leaf descriptors
   - Taskfile/check scripts
   - Slock agent dirs if local projection work is requested
   - repo-local skill roots
4. State the dependency direction: Slock points to repo; repo does not depend on
   Slock runtime state.
5. Apply the smallest repo-owned change and name the verification command.

## Core Invariants

- One mapped Slock agent owns exactly one leaf repo.
- The descriptor owns the map; Slock memory and symlinks are generated
  projections.
- `MEMORY.md` is pointer-only. Durable rules live in the repo.
- Default work model is `per_leaf_tracked`: every leaf owns its own `.work/`.
- Cross-repo refs are qualified: `<prefix>#IN-NNNN` or `<prefix>#W-NNNN`.
- Handoffs create target repo inbox/work state; Slock only notifies.
- Repo-local skills live under the leaf runtime roots actually used by agents.
- Runtime tokens, session logs, raw channel exports, and copied Slock memory are
  never canonical repo data.

## Workflow By Mode

### Design

Use `references/protocol.md` to choose surfaces, ownership boundaries, work
model, and runtime skill discovery. Produce a concrete target tree and registry
shape.

### Bootstrap

Use `references/bootstrap.md` to add descriptor fields, leaf domain claims,
Taskfile commands, target-owned install/check scripts, and generated Slock
projection rules.

### Audit

Use `references/audit-migration.md` to compare descriptor, leaf repos, Slock
agent dirs, memory files, notes, and skill roots. Lead with drift and risks.

### Repair

Repair repo-owned state first: descriptors, checks, skills, docs, and work
handoff records. Regenerate Slock `MEMORY.md` and symlinks only from the
descriptor.

### Migrate

Move durable rules from Slock notes/memory into the owning repo surface. Leave
historical logs local unless the user explicitly asks to archive them in a
repo-owned record.

## Approval Points

Ask before:

- deleting or rewriting historical Slock `notes/`
- replacing a non-symlink path inside a Slock agent dir
- introducing two-way sync between Slock and repo
- changing agent ownership, repo names, handles, or work prefixes
- committing local Slock runtime material

## Verification

Use `references/verification.md`. A good change names and runs the relevant
gate, usually:

```bash
task verify
task slock:check
```

If the task changes this skill's behavior, update `MAINTENANCE.md` in the same
change. For prior failures, add a regression case entry there; do not create a
separate `eval/` folder.

## References

| File | Use For |
| --- | --- |
| `references/protocol.md` | Ownership model, work model, Slock boundary, skill discovery |
| `references/bootstrap.md` | Descriptor, Taskfile, generated projection, and script scaffolding |
| `references/audit-migration.md` | Slock-dir scan, stale notes, unmapped agents, migration targets |
| `references/verification.md` | Mechanical gates, success checks, regression evidence |
| `MAINTENANCE.md` | Self-evolution charter and regression cases |
