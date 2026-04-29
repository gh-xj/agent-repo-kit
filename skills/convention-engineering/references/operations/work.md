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
  exposed through `task work -- ...`; canonical state lives in
  `.work/config.yaml` and `.work/items/*.yaml`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`, `task work -- triage accept IN-0001`,
  `task work -- view ready`, and `task work -- show W-0001`.
```

## Use When

Use `.work/` when the repo needs:

- captured incoming work before commitment
- human or agent triage before a request becomes canonical
- simple durable statuses: `ready`, `active`, `blocked`, `done`, `cancelled`
- built-in views over canonical state
- JSON-native output for agents and shell scripts
- local-first storage that can be inspected and versioned

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
└── items/
    └── W-0001.yaml
```

`.work/.gitignore` excludes transient lock and temporary publish paths. The
canonical state is `config.yaml` and flat YAML records under `inbox/` and
`items/`.

## Verification

At minimum, `task verify` should check:

- `.work/config.yaml` exists
- `work --store .work view ready --json` succeeds when the `work` binary is
  available

The scaffolded `work:check` task follows this shape.

## Migration From `.tickets`

`.work/` replaces the legacy `.tickets/` tracker. Do not keep both active in a
repo. If legacy tickets exist, move only still-relevant items through inbox and
triage, then remove `.tickets/` from the active tree.
