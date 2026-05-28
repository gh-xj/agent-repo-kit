# Work CLI Operator Workflow

Use this workflow to operate `.work/` without taking ownership of the work CLI
implementation or convention design.

`.work/` is local operational state and should be ignored by the repo root
`.gitignore`.

## Model

```text
.work/inbox/IN-0001.yaml --triage accept--> .work/items/W-0001.yaml
                                               |
                                               `-- optional .work/spaces/W-0001/
                                                    `-- attempts/*.yaml

.work/leases/W-0001.yaml                       # optional time-bounded claim
.work/types/<type>/policy.md                   # optional type policy
```

- Inbox entry: unaccepted demand signal. Use it for raw, duplicated,
  exploratory, or maybe-not-worth-tracking requests. Current records use
  `schema_version: 1`.
- Work item: accepted canonical record. Use it for planning, status, priority,
  area, labels, and source metadata. Current records use `schema_version: 1`.
- Work space: optional item-owned directory keyed by work ID. Use it for notes,
  research captures, plans, small evidence, and other supporting files.
- Work attempt: append-only action record under
  `.work/spaces/W-NNNN/attempts/*.yaml`. Use it to answer "what mutation ran,
  when, and with what compact inputs/results?"
- Work type: optional scaffold/template. Typed work items remain normal work
  items, but a type can scaffold `.work/spaces/<W-ID>/`.
- Work lease: optional coordination record. It says who currently claims a work
  item until a timestamp; it does not change lifecycle status.
- Type policy: optional agent-facing instructions attached to a work type.

## Store Discovery

Prefer the repo wrapper if it exists:

```bash
task work -- view ready
task work -- show W-0001
task work -- claim W-0001 --actor agent:codex:xj-mac --ttl 1h
task work -- migrate --dry-run
```

If there is no wrapper, use the binary directly:

```bash
work --store .work view ready
work --store .work show W-0001
work --store .work claim W-0001 --actor agent:codex:xj-mac --ttl 1h
work --store .work migrate --dry-run
```

If the repo is the `agent-repo-kit` checkout itself, the root wrapper runs the
source CLI:

```bash
task work -- view ready
```

## Version Diagnostics

When a command is missing, returns `unexpected argument`, or otherwise looks
older than the documented surface, check the installed binary before debugging
the store:

```bash
work version
work version --check
```

`work version --check` compares the local binary to the latest GitHub release.
Refresh Go-installed binaries with:

```bash
go install github.com/gh-xj/work-cli/cmd/work@latest
```

Do not suggest `work update`; the CLI intentionally leaves updates to `go
install` for now and package managers such as Homebrew later.

## Lifecycle

Inbox entries are pre-commit capture. Work items are committed work.

```text
.work/inbox/IN-0001.yaml --triage accept--> .work/items/W-0001.yaml
```

Use inbox when the request is raw, duplicated, exploratory, or may not deserve
tracking:

```bash
work inbox add "Investigate failing smoke run" \
  --body "Captured from user report" \
  --source "chat"
```

Accept only when the repo should track it:

```bash
work triage accept IN-0001 --area qa --priority P1
```

`triage accept` does **not** delete the inbox file. It rewrites
`.work/inbox/IN-NNNN.yaml` to `status: accepted, accepted_as: W-NNNN` and
creates the work item. The inbox file is a tombstone — leave it. Treat its
continued presence as the audit trail, not a leftover.

Create directly when the work is already understood:

```bash
work new "Document migration path" --area docs --priority P2
```

Successful `work new`, `work triage accept`, `work claim`, and `work done`
append an action record under `.work/spaces/W-NNNN/attempts/`. In JSON mode,
those commands also return the attempt path directly.

Use a work space when the work item needs files:

```text
.work/items/W-0001.yaml     # canonical status/title/metadata
.work/spaces/W-0001/        # item-owned notes and artifacts
```

Claim a work item before starting when multiple agents or humans may pick from
the same ready queue:

```bash
work claim W-0001 --actor agent:codex:xj-mac --ttl 1h
```

`work claim` writes `.work/leases/W-NNNN.yaml` with actor, optional session
provenance, `acquired_at`, and `expires_at`. A different actor cannot claim an
unexpired lease. The same actor can renew it. Expired lease files may remain on
disk; reads treat expiry as derived state.

Mark work complete with:

```bash
work done W-0001 --summary "Implemented and verified" --evidence "task ci"
```

That updates the work item to `done`, releases any active lease, optionally
writes `completion.md`, and appends a `done` attempt record.

## Migration

Use migration when a store was created before current record schema fields were
introduced:

```bash
work migrate --dry-run
work migrate
```

`work migrate --dry-run` reports records that would change. `work migrate`
backfills missing `schema_version: 1` on inbox entries and work items. It does
not migrate leases, spaces, type manifests, config, or policy files.

Migration is safe to rerun. Existing records without `schema_version` are read
as v1; records with a future unsupported `schema_version` should be treated as
a compatibility error rather than rewritten.

## Status Enum

`work` validates `--status` against a fixed set:

- `ready` — created/triaged, not yet started.
- `active` — currently being worked.
- `blocked` — waiting on an external trigger (use a label or description note
  to record the blocker).
- `done` — completed.
- `cancelled` — abandoned without completion.

Anything else is a parse error (`work: --status must be one of ...`). For
"parked / revisit later" semantics, use `blocked` plus a label like
`defer` — there is no `parked` status.

## Reading Work

Use views for scans:

```bash
work view ready
work view active
work view blocked
work view done
```

Use show for exact lookup:

```bash
work show W-0001
work show IN-0001
```

Use `--policy` to print the type policy for a typed work item:

```bash
work show W-0001 --policy
```

Use JSON for automation:

```bash
work view ready --json
work show W-0001 --json
work show W-0001 --policy --json
```

`work show W-NNNN --json` includes an active `lease` when one exists. `work
view ready --json` includes a top-level `leases` map for ready work items with
active leases.

## Typed Work Items

`work init` installs the built-in `research` type when absent. Repos may define
additional work types under `.work/types/<type>/`. The type is a scaffold
lookup key, not a native class in the core work model.

```bash
work new "Research agent inbox UX" --type research
work triage accept IN-0002 --type research
work show W-0001 --policy
```

The work item remains flat under `.work/items/W-NNNN.yaml`. Type-owned files
belong under `.work/spaces/W-NNNN/`. A non-typed work item may still use a
manually created space when the work needs item-owned files; keep the item YAML
as the source of truth.

Type policy lives at `.work/types/<type>/policy.md` by default, or at the path
declared by `policy:` in `.work/types/<type>/type.yaml`. Policy is read from
the type definition, not copied into each work space.

## Verification

Before claiming completion, run the repo's verification gate. If the repo has
a Taskfile wrapper, prefer:

```bash
task verify
```

The `work` CLI itself lives in [`gh-xj/work-cli`](https://github.com/gh-xj/work-cli).
For source-side QA against the CLI implementation, run the QA harness from
that repo (`task qa`), not from here.
