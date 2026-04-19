# Failure Taxonomy

## Triage Protocol

1. Confirm dependency preflight first (`task setup:check` or equivalent) so missing tools are not misclassified as regressions.
2. Reproduce with canonical gate and same inputs.
3. If failure signature indicates command exited without explicit test failures, rerun once deterministically.
4. Read run summary and step logs before changing code.
5. Classify and route fix.
6. Re-run gate and preserve closure evidence.

## Artifact-First Debug Order

1. Canonical summary (`artifacts/harness/verify/<timestamp>/summary.json` or repo equivalent)
2. Failed step log from summary log path
3. Previous successful run summary for diff

## Classification Rules

| Class | Meaning | Primary Action |
|---|---|---|
| `new_product_logic` | Intentional behavior change | Update contract + evidence intentionally |
| `fixture_or_time_drift` | Input nondeterminism | Repair deterministic fixtures/time windows |
| `infra_or_flake` | Environment instability | Isolate and track with owner+expiry |
| `true_regression` | Unintentional defect | Fix root cause and keep gate red until fixed |

## Guardrails

- Do not classify as `true_regression` before deterministic rerun + artifact read.
- Do not bypass gates without explicit owner and expiry.
- Missing prerequisite binaries (for example `bun`, `task`, `wire`) are environment failures until proven otherwise.
