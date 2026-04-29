# `work` CLI v0

`work` is a local-first, Linear-inspired work tracker for agents. It keeps the
canonical work graph in repo files, exposes a JSON-native CLI, and treats
triage as a first-class step before captured requests become committed work.

This document defines the v0 product contract. It is intentionally narrower
than a full project-management system.

## Product Principles

1. **Local-first Linear for agents.** The source of truth lives under the
   repo-local `.work/` directory, which is ignored by git by default. The CLI
   should be useful without a hosted service, browser session, or account sync.
   Its primary operators are agents and terminal workflows.
2. **Inbox before commitment.** Raw requests, observations, and imported
   legacy notes enter `.work/inbox/`. They are not canonical work until triage
   accepts them.
3. **Views over canonical state.** Built-in views are derived from
   `.work/items/*.yaml`. v0 does not persist custom views.
4. **JSON-native CLI.** Every command that returns records must support stable
   machine-readable JSON, including a `schema_version`. Human output is a
   convenience view over the same data.
5. **No future-shaped persistence.** A persisted field or folder must have a
   current command that writes it, a current command that reads it, and a
   current invariant that validates it.

## Command Surface

Global flags:

- `--store <path>` selects the work store. Default: `.work`.
- `--json` emits machine-readable JSON.
- `--no-color` disables colorized human output.
- `-v`, `--verbose` enables debug logs.

| Command | Purpose |
| --- | --- |
| `work init` | Create the `.work/` store layout, default config, and built-in type presets. It initializes only the current store. |
| `work inbox` | List inbox entries waiting for triage. |
| `work inbox add` | Capture a raw request into `.work/inbox/` with title, body/context, source metadata, and optional priority hints. |
| `work triage accept` | Promote an inbox entry into a canonical work item under `.work/items/`, preserving source metadata and marking the inbox entry accepted. |
| `work new` | Create a canonical work item directly when the work is already understood and does not need inbox triage. |
| `work view` | Render a built-in named view by querying canonical items. Default view: `ready`. |
| `work show` | Show one record by ID, including canonical work items and inbox entries. |

The command surface deliberately does not include migration commands,
dependency commands, relation commands, or proof-gated closure commands in v0.

Typed work items are scaffold-backed, not native. `work new "Understand X" --type
research` and `work triage accept IN-0001 --type research` resolve
`.work/types/research/type.yaml`, create the generic item with
`type: research`, and scaffold `.work/spaces/W-NNNN/` from the work type.
Core `work` does not know what `research` means after scaffold creation.

`research` is the built-in v0 preset. `work init` installs it when absent and
leaves an existing `.work/types/research/` directory untouched.

Default v0 views:

- `ready` — accepted work that can be claimed.
- `active` — work currently in motion.
- `blocked` — work waiting on an unblock condition.
- `done` — completed work.

## Store Layout

```text
.work/
|-- config.yaml
|-- .lock
|-- inbox/
|   `-- IN-0001.yaml
|-- items/
|   `-- W-0001.yaml
`-- types/
    `-- research/
        |-- type.yaml
        `-- scaffold/
            |-- README.md
            |-- RULES.md
            |-- notes.md
            |-- findings.md
            |-- raw/.keep
            `-- pages/.keep
```

- `.work/config.yaml` stores the work schema version, ID prefixes/allocation
  settings, default state names, and repo-local defaults.
- The repo root `.gitignore` should include `/.work`; the ledger is local
  operational state, not release source.
- `.work/.lock` is a short-lived mutation lock. It prevents parallel agents
  from allocating the same IDs or publishing partial multi-file writes.
- `.work/inbox/` stores captured requests before triage. Inbox records may be
  incomplete, duplicated, stale, or exploratory.
- `.work/items/*.yaml` stores canonical work items. These are the records agents
  plan against and mutate after triage.
- `.work/types/research/` is the built-in research preset installed by
  `work init` if it does not already exist.
- `.work/types/<type>/type.yaml` declares an optional additional work type
  extension when a repo chooses to use typed work beyond the preset.
- `.work/spaces/<W-NNNN>/` stores type-owned files for a typed item and is
  created lazily. The path is derived from the work ID and is not persisted in
  the item record.

Minimal work type manifest:

```yaml
schema_version: 1
id: research
description: Research workspace
scaffold: scaffold
```

Required fields: `schema_version`, `id`. Optional fields: `description` and
`scaffold`. The built-in research scaffold records its human-readable version
in `scaffold/README.md`.

## Canonical State

Canonical state lives in `.work/items/*.yaml`.

Inbox entries are evidence of demand, not accepted work. `work triage accept`
is the boundary between "someone captured this" and "the repo is choosing to
track this as work."

Views are read models over canonical state. The v0 built-ins filter by status,
and each result is recomputed from records instead of written back as
board-specific state.

Typed item records stay flat:

```yaml
id: W-0001
title: Understand X
type: research
status: ready
```

Type-specific state belongs inside `.work/spaces/W-0001/`, not in nested
fields under `.work/items/W-0001.yaml`.
