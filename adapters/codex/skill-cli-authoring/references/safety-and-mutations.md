# Safety And Mutations

Every command should declare its safety class before implementation.

## Safety Classes

| Class       | Meaning                                          | Examples                         |
| ----------- | ------------------------------------------------ | -------------------------------- |
| read-only   | Does not modify disk, network, or external state | `search`, `show`, `validate`     |
| additive    | Adds records; does not remove or overwrite       | `capture`, `append-note`         |
| idempotent  | Repeating same args has no further effect        | `ensure`, `sync --check`         |
| mutating    | Updates existing state                           | `update`, `apply`                |
| destructive | Deletes, overwrites, publishes, or revokes       | `delete`, `reset`, `publish`     |
| open-world  | Talks to external services or broad filesystem   | `fetch`, `upload`, `crawl`       |

Use these classes in docs, tests, and future MCP annotations.

## Write Command Rules

Mutating commands should:

- require explicit target IDs or paths
- reject empty updates
- return no-op success when nothing changed
- include `changed` and `changed_fields` in JSON
- write atomically when touching files
- avoid read commands that write cleanup as a side effect

Destructive commands should also:

- support `--dry-run`
- require `--yes` or an equivalent explicit confirmation flag
- describe affected targets before applying
- avoid broad globs unless paired with `--dry-run`

## Dry-Run Contract

`--dry-run` must not mutate state. It should return the same plan shape as the
real command:

```json
{
  "schema_version": "v1",
  "ok": true,
  "dry_run": true,
  "would_change": true,
  "planned_changes": [
    {"action": "write", "path": "out/report.json"}
  ]
}
```

Avoid dry-runs that only print prose. Agents need machine-checkable planned
changes.

## Plan Then Apply

For consequential mutations, prefer a two-step shape:

```bash
tool plan --input source.json --json
tool apply PLAN-ID --dry-run --json
tool apply PLAN-ID --yes --json
```

This is not ceremony for its own sake. It separates three different jobs:

- `plan` converts messy input into a stable proposed change.
- `apply --dry-run` proves what would happen in the current environment.
- `apply --yes` performs the already-described change.

That separation helps agents because the model can inspect a compact plan
instead of re-reading the full source input, can ask for approval using concrete
targets, and can retry the apply step without re-planning from changed context.

Challenge the pattern when the operation is already a tiny idempotent field
update:

```bash
tool update ITEM-ID --status done --json
```

Do not force plan/apply onto every command. Use it when there is broad file
mutation, destructive behavior, external publishing, generated output, or
multi-step interpretation where review before execution matters.

## Idempotency

Prefer idempotent verbs where possible:

- `ensure-config` over `create-config`
- `sync --check` and `sync` over `copy`
- `update --status done` over `transition-done`

If a command is not idempotent, state why in the command docs and tests.

## External Services

When a skill-local CLI talks to an external service:

- make authentication checks explicit
- keep read commands usable for diagnostics
- classify network errors separately from invalid input
- include request IDs or remote IDs in JSON when available
- never print secrets in stdout, stderr, logs, fixtures, or test failures

## Audit Output

For important writes, record enough evidence for review:

- target IDs
- changed fields
- generated file paths
- source inputs
- dry-run plan ID when applicable
- external URLs or request IDs when safe

Audit output is not a substitute for version control or tests. It is the minimum
state the agent needs to explain what happened.
