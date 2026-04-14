# Operation: Tickets

Flat-file ticket tracker. One of the standard operational conventions that
convention-engineering scaffolds into a repo.

## Adopt this convention in 3 steps

1. **Copy the template and initialize.**

   ```bash
   cp -R ~/.claude/skills/convention-engineering/references/templates/tickets \
         <repo>/.tickets
   cd <repo>
   task -d .tickets init
   ```

2. **Wire discoverability into `CLAUDE.md` and `AGENTS.md`.**
   Add (or merge into an existing `## Conventions` section):

   ```markdown
   - **Tickets** — flat-file work tracker at `.tickets/`. Read `.tickets/README.md`
     for the verb surface and `.tickets/harness/schema.yaml` for the state
     machine. Daily commands:
     `task -d .tickets {new|list|transition|close|test}`.
   ```

3. **Customize and commit.**
   Edit `<repo>/.tickets/harness/taxonomy.yaml` to match the project's categories.
   Commit `.tickets/` **except** `.tickets/.lock/` (already gitignored). The
   rest of the template is universal and should not be edited per repo.

Manual copy-paste, by design — no automatic text surgery on agent contract files.

## When to adopt

Use `.tickets/` when the repo needs:

- Tracked transitions (OPEN → IN_PROGRESS → REVIEW → DONE) with an audit trail
- Schema-validated categories and priorities
- Per-ticket bodies (more than a one-line TODO can carry)

Skip it for casual TODOs (use `docs/tech-debt-tracker.md` or similar flat list).

## Verb surface

| Verb         | Usage                                                                           |
| ------------ | ------------------------------------------------------------------------------- |
| `init`       | `task -d .tickets init`                                                         |
| `list`       | `task -d .tickets list`                                                         |
| `new`        | `task -d .tickets new -- "title" --priority P1 --category bug [--estimated 2h]` |
| `transition` | `task -d .tickets transition -- T-04 --to DONE [--note "reason"]`               |
| `close`      | `task -d .tickets close -- T-06 T-08 --reason "superseded"`                     |
| `test`       | `task -d .tickets test`                                                         |

`create` is an alias for `new`.

## State machine

```
OPEN        → IN_PROGRESS | BLOCKED | CANCELLED
IN_PROGRESS → BLOCKED | REVIEW | DONE | CANCELLED
BLOCKED     → IN_PROGRESS | CANCELLED
REVIEW      → IN_PROGRESS | DONE | CANCELLED
DONE        → (terminal)
CANCELLED   → (terminal)
```

Authoritative source: `<repo>/.tickets/harness/schema.yaml`.

## Invariants enforced at the CLI

- `priority` ∈ {P0, P1, P2, P3}
- `category` must exist in `harness/taxonomy.yaml`
- Transitions must follow the schema graph
- IDs monotonic, never reused (allocation locked via `.lock/`)
- Titles and cancel reasons are YAML-escaped (newlines stripped, quotes escaped)

## Directory layout

```
.tickets/
├── Taskfile.yml
├── README.md
├── audit-log.md                # append-only, committed
├── harness/
│   ├── schema.yaml             # universal
│   ├── taxonomy.yaml           # project-specific
│   └── test-ticket-system.sh   # executable test harness
└── all/
    └── T-NN-slug/ticket.md     # one ticket = one dir
```

## Audit trail

Each `new` / `transition` / `close` appends a line to `.tickets/audit-log.md`.
Terminal states also record metadata in the ticket's own Lifecycle Log section
and in the frontmatter (`resolved:` for DONE, `cancel_reason:` for CANCELLED).

Commit `.tickets/` to make this audit trail survive across machines. `git log`
adds a second, coarser audit layer at commit granularity.

## Template location

`~/.claude/skills/convention-engineering/references/templates/tickets/`.
