# Operation: Work

Local-first work tracker for agent-operated repos. `.work/` is the standard
operational convention for captured work, triage, canonical work items, saved
views derived from canonical state, and machine-readable CLI output.

## Adopt

Preferred path:

```bash
ark init --repo-root <repo> --ops work
```

Manual path:

```bash
work --store <repo>/.work init
```

Then add the pointer snippet to `AGENTS.md` and `CLAUDE.md`:

```md
- **Work** - local-first work tracker at `.work/`. The repo-local CLI is
  exposed through `task work -- ...`; local state lives in the ignored
  `.work/config.yaml` and `.work/items/*.yaml`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`, `task work -- triage accept IN-0001`,
  `task work -- view ready`, and `task work -- show W-0001`.
```

Add `/.work` to the repo root `.gitignore`. The work ledger is operational
state, not release source.

## Use When

Use `.work/` when the repo needs:

- captured incoming work before commitment
- human or agent triage before a request becomes canonical
- simple durable statuses: `ready`, `active`, `blocked`, `done`, `cancelled`
- built-in views over canonical state
- JSON-native output for agents and shell scripts
- local-first storage that can be inspected without committing operational churn

For casual one-off TODOs, inline comments or a short checklist may still be
enough. Do not create a tracker if there is no lifecycle.

## Commands

| Verb | Command |
| --- | --- |
| Init | `work init` |
| Inbox | `work inbox` |
| Capture | `work inbox add "title" --body "context" --source "source"` |
| Accept | `work triage accept IN-0001 --area cli --priority P1` |
| Direct create | `work new "title" --area docs --priority P2` |
| Typed create | `work new "title" --type research` |
| View | `work view ready` |
| Show | `work show W-0001` |

If the repo wires the root Taskfile wrapper, use `task work -- <args>`:

```bash
task work -- inbox
task work -- view ready
task work -- show W-0001
```

## Layout

```text
.work/
├── .gitignore
├── config.yaml
├── inbox/
│   └── IN-0001.yaml
├── items/
│   └── W-0001.yaml
└── types/
    └── research/
        ├── type.yaml
        └── scaffold/
            ├── README.md
            ├── RULES.md
            ├── notes.md
            ├── findings.md
            ├── raw/.keep
            └── pages/.keep
```

`.work/.gitignore` excludes transient lock and temporary publish paths inside
the ignored work store. The canonical local state is `config.yaml` and flat
YAML records under `inbox/` and `items/`. `work init` installs the built-in
`research` type if absent and does not overwrite an existing
`.work/types/research/` directory.

## Work Types

`type:` is a scaffold lookup key, not a native work item class. Core `work`
owns IDs, statuses, inbox, items, views, lifecycle, and JSON output. Work types
own optional workspace files.

Built-in research work type:

```text
.work/types/research/
├── type.yaml
└── scaffold/
    ├── README.md
    ├── RULES.md
    ├── notes.md
    ├── findings.md
    ├── raw/.keep
    └── pages/.keep
```

```yaml
schema_version: 1
id: research
description: Research workspace
scaffold: scaffold
```

Creating `work new "Understand X" --type research` writes the generic item
with `type: research` and copies the scaffold to `.work/spaces/W-NNNN/`.
`work show` and `work view` continue to work if the work type is later removed.
Type-specific state belongs in the item space, not in nested item YAML.

## Verification

At minimum, `task verify` should check:

- `work --store .work init` succeeds
- `work --store .work view ready --json` succeeds

The scaffolded `work:check` task follows this shape.
