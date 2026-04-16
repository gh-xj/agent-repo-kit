# .tickets — flat-file ticket tracker

Lightweight task tracker for `demo-repo`. Tickets are YAML-frontmatter
markdown files under `all/`. This repo keeps the tracker initialized as a
concrete `agent-repo-kit` example, while state still lives in the `status:`
frontmatter field.

## Per-repo edits allowed

- Edit `harness/taxonomy.yaml` (categories, phases, terms).
- Edit this README to add repo-specific notes.
- Everything else is universal — do NOT edit (so updates can flow from the
  convention-engineering template).

## Layout

```text
.tickets/
├── Taskfile.yml        # CLI: task -d .tickets <cmd>
├── README.md
├── audit-log.md        # append-only transition log
├── harness/
│   ├── schema.yaml            # frontmatter + transitions
│   ├── taxonomy.yaml          # categories, phases, terms
│   └── test-ticket-system.sh  # executable tests
└── all/
    └── T-NN-slug/
        └── ticket.md
```

## Commands

```bash
task -d .tickets init
task -d .tickets list
task -d .tickets new -- "refresh demo repo pointer block" --priority P1 --category contract --estimated 1h
task -d .tickets transition -- T-01 --to IN_PROGRESS
task -d .tickets transition -- T-01 --to DONE --note "merged to main"
task -d .tickets close -- T-06 T-08 --reason "superseded"
task -d .tickets test
```

## Rules

1. **Schema enforced at the CLI.** `priority` must be `P0..P3`, `category` must
   exist in `harness/taxonomy.yaml`, transitions must follow `harness/schema.yaml`.
2. **Terminal metadata is outcome-specific.** `DONE` writes `resolved:`;
   `CANCELLED` writes `cancel_reason:` and is recorded in both the ticket's
   Lifecycle Log and `audit-log.md`.
3. **IDs are monotonic and never reused.** Allocation is locked via `.lock`.
4. **Estimates first, actuals second.** Record `estimated` at creation; add
   `actual:` when known.
5. **Taxonomy self-evolves.** Extend `harness/taxonomy.yaml` when new work
   categories surface.
6. **Ticket bodies stay minimal.** Summary, Action Items, Lifecycle Log.
