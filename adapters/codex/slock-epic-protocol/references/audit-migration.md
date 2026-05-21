# Audit And Migration

Use this to find and remove protocol drift between Slock runtime folders and
repo-owned surfaces.

## Audit Inputs

Inspect:

- canonical descriptor map
- epic `repo/<leaf>` symlinks
- leaf `.conventions.yaml`
- leaf `AGENTS.md` / `CLAUDE.md`
- leaf `.work/types`
- leaf repo-local skill roots
- Slock agent dirs for mapped agents
- unmapped Slock agents that mention the product, old repo name, or retired
  roles

## Slock Directory Findings

For each mapped Slock agent, report:

```text
agent_id
mapped key
repo symlink target
MEMORY.md pointer-only? yes/no
notes files
.slock runtime files present? yes/no
stale active rules
migration target
risk
```

For unmapped agents, report only if they contain active-looking references to
the product. They may need a pointer to the canonical epic registry or a note
that their material is historical.

## Migration Targets

| Slock material | Repo target |
| --- | --- |
| Source lists, scan heuristics | Scout docs or Scout skill references |
| Brief quality bars, ingestion rules | Analysis skill references |
| Draft calibration, review packets, platform rules | Content Portal skill references |
| Channel policy, process evolution, strategy signals | Company operations or strategy docs |
| Work logs for active items | Owning repo `.work` item/workspace |
| Historical run logs | Leave local or archive as an explicitly historical repo record |
| Runtime sessions, tokens, raw exports | Never migrate |

Do not copy whole notes files blindly. Extract durable rules, remove stale
paths, and write them into the owning repo's current taxonomy.

## Repair Order

1. Move durable protocol into repo-owned docs/skills.
2. Update repo verification to check the new owner surface.
3. Regenerate Slock memory.
4. Run local Slock projection checks.
5. Leave historical notes in place unless the user asks for archival cleanup.

## Stale-Reference Sweep

Search active contract surfaces for old names:

- old monorepo path or repo name
- retired agent handles
- retired skill names
- obsolete output paths
- active references to Slock `notes/` as source of truth

Allow historical references only when the file clearly marks them as history.

## Risk Notes

The common failure is a Slock `MEMORY.md` or `notes/` file becoming a second
protocol store. Another common failure is adding a repo-local skill for one
runtime but forgetting the runtime actually used by the Slock agent.
