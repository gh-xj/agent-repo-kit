---
name: convention-evaluator
description: Use when scoring repo conventions against a convention contract, validating handoff artifacts, or producing a skeptical convention evaluation report from isolated context.
---


# Convention Evaluator

Skeptical evaluation surface for convention work. This convention owns contract semantics, rubric thresholds, evidence expectations, and the machine-readable evaluation result.

## Scope

Owned:

- convention contract schema semantics and versioning
- rubric dimensions, score scale, and thresholds
- handoff, launcher receipt, evidence, and evaluation result contracts
- skeptical evaluation guidance from fresh context
- scoring repos that adopt operational conventions (work, wiki) via existing rubric dimensions — no dedicated schema fields or dimensions

Not owned:

- repo convention generation or bootstrap decisions
- contract values chosen for a specific repo run
- repo-local agent tooling or runtime metadata

## Routing Table

| Question                                                                             | Reference                           |
| ------------------------------------------------------------------------------------ | ----------------------------------- |
| What does each top-level contract field mean?                                        | `references/contract-schema.md`     |
| How are conventions scored and which dimensions hard-fail?                           | `references/rubric.md`              |
| What must be in `handoff.json`, `launch-receipt.json`, and `evaluation_result.json`? | `references/handoff-and-results.md` |

## Quick Start

1. Read `references/handoff-and-results.md` to validate the launcher boundary.
2. Read `references/contract-schema.md` for contract semantics.
3. Read `references/rubric.md` for thresholds and skeptical scoring.
4. Evaluate from fresh context only: repo state plus handoff artifacts.
5. Write both the human report and `evaluation_result.json`.

## Boundaries

- `launch-receipt.json` written by the launcher is the source of truth for isolation.
- `fresh_context` must be launcher-proven, not evaluator self-attested.
- `convention-engineering` may pass `evaluation_inputs.repo_risk`, but it may not lower evaluator thresholds.
- Authoring of agent tooling itself is out of scope; this convention only judges convention work.

## Cross-Convention Routing

| Domain                             | Convention               | When                                      |
| ---------------------------------- | ------------------------ | ----------------------------------------- |
| Convention planning and generation | `convention-engineering` | Repo bootstrap, patching, contract values |

Optional, agent-harness-specific: when running under an agent runtime that
exposes a `skill-builder` (or equivalent) convention for authoring agent
tooling, defer authoring questions there. This routing is not required for
the evaluator to function.
