# Task Gates Contract

## Required Outcomes

- A single canonical merge gate exists (`verify` or repo equivalent).
- `smoke` and `regress` are atomic: run and validate artifacts in one call.
- Local and CI use the same canonical gate command.
- The canonical gate emits machine-readable summary artifacts and per-step logs for debugging.
- The canonical gate fails fast on missing runtime dependencies required by gated steps (for example `bun` for frontend test/build gates).

## Recommended Topology

| Task | Purpose |
|---|---|
| `check:rule-tests` | static rule coverage contract |
| `smoke` | deterministic fast signal + contract validation |
| `regress` | fixture/scenario replay + contract validation |
| `verify` | canonical aggregate gate |

## Verify Artifact Contract (Harness Pattern)

- Run outputs are written under a timestamped directory, for example: `artifacts/harness/verify/<timestamp>/`.
- A run-level summary file exists: `summary.json`.
- Per-step logs exist for every gate stage (pass/fail/skip), one log file per step.
- Summary includes: overall status, step statuses, durations, log paths, and failure metadata.

## Checker Guidance

- Resolve `includes` chains when searching task tokens.
- Allow token checks against multiple taskfiles.
- Fail with explicit missing token names, not generic parse errors.
- Verify canonical gate references a deterministic artifact root and summary path contract.
