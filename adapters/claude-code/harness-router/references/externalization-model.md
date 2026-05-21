# Externalization Model

Use this model before choosing a destination. A learning should move out of
session context only when a durable artifact will reduce a future cognitive
burden more than it adds maintenance burden.

## Burdens

- `continuity`: remember state, history, decisions, evidence, and resumable
  context across sessions.
- `procedure`: repeat a workflow or judgment pattern correctly.
- `interaction`: invoke tools, commands, schemas, adapters, or protocols
  correctly.
- `governance`: control authority, permissions, review, rollback, privacy, or
  safety boundaries.
- `observability`: preserve traces, outcomes, failures, and provenance.
- `planning`: decompose goals, sequence work, and track active follow-up.
- `evaluation`: decide whether behavior still works through tests, checks,
  evals, or benchmark fixtures.

If no burden is clear, do not promote the learning.

## Artifact Classes

Pick the class before the exact path:

- `instruction`: always-loaded repo or harness guidance.
- `skill`: on-demand procedural expertise.
- `skill_reference`: bulky or conditional skill detail.
- `protocol`: command grammar, schema, lifecycle, adapter contract, permission
  boundary, or generated-surface contract.
- `docs`: durable explanation, decision record, design rationale, or research
  synthesis.
- `work`: active item, item-owned workspace, notes, evidence, or plan.
- `memory`: low-authority personal or local recall hint.
- `check`: hook, lint, CI, test, eval, or grader that enforces behavior.
- `structured_store`: machine-readable trace, inventory, proposal, or routing
  record that a current command reads or validates.

Do not create a new artifact class unless it changes a routing decision.

## Lifecycle

Use the smallest lifecycle needed:

```text
captured -> staged -> proposed -> approved -> implemented -> verified
rejected
deprecated
```

- `captured`: raw source or session evidence exists.
- `staged`: active thinking lives in a work space, research page, or notes.
- `proposed`: a human-reviewable durable change exists.
- `approved`: the user accepted the change.
- `implemented`: the target artifact was edited.
- `verified`: the relevant check, test, or review passed.
- `rejected`: the learning should not be promoted.
- `deprecated`: an older artifact should be removed or replaced.

Do not make lifecycle state a stored field unless a command reads or validates
it. In ordinary proposals, use it as decision vocabulary.

## Decision Test

Promote only when the artifact makes future work more:

- stable
- inspectable
- reusable
- governable

than keeping the learning in chat, local notes, or the current work space.

