# Command Contract

Design commands from examples first. If the examples feel ambiguous, the code
will be ambiguous too.

## Command Grammar

Prefer:

```bash
tool search --query "text" --json
tool show ITEM-ID --json
tool plan --input source.json --json
tool apply PLAN-ID --dry-run --json
tool apply PLAN-ID --yes --json
```

Avoid:

```bash
tool do "figure it out"
tool update status=done +foo -bar
tool run --magic
tool apply latest
```

Rules:

- One target ID argument is better than implicit "current" state.
- Flags beat positional free-form strings for structured values.
- Use enums for closed sets.
- Use repeatable flags for list additions.
- Use explicit clear flags for removals: `--clear-labels`, `--clear-cache`.
- Use `--input <path|->` for structured input files and stdin.

## Output Modes

Plain text:

- concise
- stable enough for humans
- not a parsing contract

JSON:

- one complete JSON document
- best for reads and short mutations
- include `schema_version`

NDJSON:

- one JSON object per line
- best for streaming progress, logs, or large result sets
- every line must parse independently

Never emit ANSI color, tables, spinners, or progress bars in JSON or NDJSON
mode.

## JSON Envelope

Use a predictable envelope:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "apply",
  "changed": true,
  "result": {},
  "warnings": []
}
```

For mutations, include enough verification data:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "update",
  "changed": true,
  "changed_fields": ["status"],
  "before": {"status": "ready"},
  "after": {"status": "done"}
}
```

Keep `before` small. Do not dump huge input documents unless the caller asked
for them.

## Error Shape

Errors should be structured in machine mode:

```json
{
  "schema_version": "v1",
  "ok": false,
  "error": {
    "code": "not_found",
    "message": "item W-9999 was not found",
    "target": "W-9999",
    "retryable": false
  }
}
```

Use stable error codes:

- `usage`
- `not_found`
- `invalid_input`
- `conflict`
- `precondition_failed`
- `permission_denied`
- `external_unavailable`
- `internal`

## Exit Codes

Recommended baseline:

- `0` success, including no-op success
- `1` runtime failure
- `2` usage error
- `3` conflict or precondition failure

Do not invent a large exit-code taxonomy unless scripts already need it.

## Compatibility

After a command is used by a skill:

- adding JSON fields is okay
- removing or renaming JSON fields is breaking
- changing enum strings is breaking
- changing default target selection is risky
- changing plain text is lower risk but still avoid churn

If the contract must break, bump `schema_version` and update the parent skill in
the same change.
