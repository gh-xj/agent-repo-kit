# `work` CLI v0

`work` is a local-first, Linear-inspired work tracker for agents. It keeps the
canonical work graph in repo files, exposes a JSON-native CLI, and treats
triage as a first-class step before captured requests become committed work.

This document defines the v0 product contract. It is intentionally narrower
than a full project-management system.

## Product Principles

1. **Local-first Linear for agents.** The source of truth lives under `.work/`
   in the repository. The CLI should be useful without a hosted service,
   browser session, or account sync. Its primary operators are agents and
   terminal workflows.
2. **Inbox before commitment.** Raw requests, observations, and imported
   legacy notes enter `.work/inbox/`. They are not canonical work until triage
   accepts them.
3. **Views over canonical state.** A view is a saved query over `.work/items/`,
   not a second copy of work state. Reordering, filtering, and grouping belong
   in `.work/views.yaml`.
4. **Relations later.** v0 should avoid dependency graphs, related-issue
   edges, blocked-by trees, and cross-item relation semantics. Items may carry
   plain-text references, but relation-aware commands are outside v0.
5. **JSON-native CLI.** Every command that returns records must support stable
   machine-readable JSON, including a `schema_version`. Human output is a
   convenience view over the same data.
6. **Evidence later.** v0 should leave room for evidence fields and proof links,
   but evidence-gated closure is outside this command set.

## Command Surface

Global flags:

- `--store <path>` selects the work store. Default: `.work`.
- `--json` emits machine-readable JSON.
- `--no-color` disables colorized human output.
- `-v`, `--verbose` enables debug logs.

| Command | Purpose |
| --- | --- |
| `work init` | Create the `.work/` store layout, default config, and default views. It must not import `.tickets/` automatically. |
| `work inbox` | List inbox entries waiting for triage. |
| `work inbox add` | Capture a raw request into `.work/inbox/` with title, body/context, source metadata, and optional priority hints. |
| `work triage accept` | Promote an inbox entry into a canonical work item under `.work/items/`, preserving source metadata and marking the inbox entry accepted. |
| `work new` | Create a canonical work item directly when the work is already understood and does not need inbox triage. |
| `work view` | Render a named view from `.work/views.yaml` by querying canonical items. Default view: `ready`. |
| `work show` | Show one record by ID, including canonical work items and inbox entries. |

The command surface deliberately does not include migration commands,
dependency commands, or evidence/closure commands in v0.

Default v0 views:

- `ready` — accepted work that can be claimed.
- `active` — work currently in motion.
- `blocked` — work waiting on an unblock condition.
- `done` — completed work.

## Store Layout

```text
.work/
|-- config.yaml
|-- views.yaml
|-- .lock
|-- inbox/
`-- items/
    `-- W-0001/
        |-- item.yaml
        |-- events.jsonl
        `-- evidence/
```

- `.work/config.yaml` stores the work schema version, ID prefixes/allocation
  settings, default state names, and repo-local defaults.
- `.work/views.yaml` stores named view definitions. Views are queries and
  presentation settings over canonical state; they do not own item data.
- `.work/.lock` is a short-lived mutation lock. It prevents parallel agents
  from allocating the same IDs or publishing partial multi-file writes.
- `.work/inbox/` stores captured requests before triage. Inbox records may be
  incomplete, duplicated, stale, or exploratory.
- `.work/items/` stores canonical work items. These are the records agents
  plan against and mutate after triage.

## Canonical State

Canonical state lives in `.work/items/`.

Inbox entries are evidence of demand, not accepted work. `work triage accept`
is the boundary between "someone captured this" and "the repo is choosing to
track this as work."

Views are read models over canonical state. A view may filter by state, area,
label, priority, recency, or other item fields, but the result should be
recomputed from records instead of written back as board-specific state.

## `.tickets` Migration Stance

There is no automatic `.tickets` importer in v0. Legacy tickets should either:

- go through `work inbox add` followed by `work triage accept`, or
- be manually recreated as canonical work items with `work new`.

Migration should preserve useful legacy metadata such as `legacy_id` and
`legacy_path`, but should not blindly promote every old ticket into a new work
item.
