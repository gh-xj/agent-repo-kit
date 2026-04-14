# convention-evaluator

Skeptical convention scoring from isolated context.

## Read Order

1. `SKILL.md` (convention doc and routing table)
2. `references/handoff-and-results.md`
3. `references/contract-schema.md`
4. `references/rubric.md`

## What This Convention Owns

- contract schema semantics and versioning
- evaluator-owned thresholds and score scale
- handoff, launcher, evidence, and result artifact contracts
- skeptical report structure for convention audits

## Required Inputs

- repo root
- handoff manifest path

The evaluator reads only the live repo plus the files referenced by the handoff manifest. `fresh_context` is proven by `launch-receipt.json`, not by evaluator narration.

## Required Artifacts

- `handoff.json`
- `launch-receipt.json`
- raw evidence bundle directory with `manifest.json`
- markdown evaluation report
- `evaluation_result.json`

## Result Statuses

- `passed`
- `semantic_failed`
- `infrastructure_failed`

## Threshold Model

See `references/rubric.md` for the full policy. The hard-fail dimensions are:

- `enforceability`
- `verification`
- `ownership_clarity`

Soft-fail dimensions:

- `legibility`
- `drift_resistance`

`evaluation_inputs.repo_risk` may raise soft-fail thresholds, but the contract does not lower evaluator policy.
