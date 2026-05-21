# CLI Contract: <tool-name>

Owner skill: `<skill-name>`

Purpose: <one sentence describing the deterministic operation this CLI owns>

## Safety Classes

| Command | Class | Requires dry-run | Requires approval | Notes |
| ------- | ----- | ---------------- | ----------------- | ----- |
| `show`  | read-only | no | no | Exact read by ID |
| `plan`  | read-only | no | no | Produces proposed changes |
| `apply` | mutating | yes | yes, `--yes` | Applies a plan |

## Commands

### `<tool> show <id> --json`

Read one object by stable ID.

Stable JSON fields:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "show",
  "result": {}
}
```

### `<tool> plan --input <path|-> --json`

Convert source input into a reviewable plan. Does not mutate state.

Stable JSON fields:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "plan",
  "plan_id": "PLAN-0001",
  "planned_changes": []
}
```

### `<tool> apply <plan-id> --dry-run --json`

Report what would change in the current environment. Does not mutate state.

### `<tool> apply <plan-id> --yes --json`

Apply a previously reviewed plan.

Stable JSON fields:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "apply",
  "changed": true,
  "changed_fields": [],
  "artifacts": []
}
```

## Error Codes

| Code | Meaning | Retryable |
| ---- | ------- | --------- |
| `usage` | Invalid flags or missing required argument | no |
| `not_found` | Target ID or input path was not found | no |
| `invalid_input` | Input parsed but failed validation | no |
| `conflict` | State changed since planning | maybe |
| `precondition_failed` | Required tool, auth, or file missing | maybe |
| `external_unavailable` | Remote dependency failed | yes |
| `internal` | Unexpected implementation failure | maybe |

## Exit Codes

- `0`: success, including no-op success
- `1`: runtime failure
- `2`: usage error
- `3`: conflict or precondition failure

## Verification

```bash
<tool> --help >/dev/null
<tool> show fixture --json | jq -e . >/dev/null
<tool> plan --input testdata/input.json --json | jq -e . >/dev/null
<tool> apply PLAN-0001 --dry-run --json | jq -e . >/dev/null
```

## Compatibility

Breaking changes:

- removing or renaming JSON fields
- changing enum strings
- changing safety class
- changing whether a command mutates state

Allowed changes:

- adding optional JSON fields
- adding read-only commands
- improving human-readable output
