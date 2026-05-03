---
name: convention-evaluator
description: "Use when scoring how well a repo lives up to its declared `.conventions.yaml` from a fresh, skeptical context. Produces a graded markdown report (PASS/FAIL across legibility, enforceability, verification) at `docs/reviews/YYYY-MM-DD_<topic>_evaluation.md`. For gap-finding and bootstrap, use convention-engineering instead."
---

# Convention Evaluator

Skeptical scoring of a repo's adoption of its `.conventions.yaml`. Different
job from `convention-engineering`:

- `convention-engineering` audit → gap report ("X is missing").
- `convention-evaluator` → graded judgment ("legibility 3/4, here is why").

Run from fresh context. The agent that wrote the conventions should not
score them.

## When To Use

- A repo has declared `.conventions.yaml` and you want a skeptical read of
  whether it lives up to the declaration.
- A convention refactor just landed and you want a calibrated PASS/FAIL.
- Periodic drift check on a repo that you know well — open a fresh session
  for the score.

Skip when:

- The repo has no `.conventions.yaml`. Use `convention-engineering` to
  bootstrap one first.
- You only need a gap list (use `convention-engineering` audit).

## Inputs

- The live repo.
- `.conventions.yaml` at its root.

That is it. No launcher receipts, no handoff manifests, no evidence
bundles — the rubric runs on what is in the repo.

## Routing

| Question                                        | Reference                                                    |
| ----------------------------------------------- | ------------------------------------------------------------ |
| What dimensions and thresholds apply?           | `references/rubric.md`                                       |
| What should the report look like?               | `references/report-template.md`                              |
| What does `.conventions.yaml` mean for scoring? | `convention-engineering/references/core/conventions-yaml.md` |

## Quick Start

1. Open a fresh agent session in the repo.
2. Read `.conventions.yaml` at the repo root.
3. Read `references/rubric.md` for the three dimensions and thresholds.
4. Score each dimension 0-4. Cite a path or command output for every claim.
5. Decide PASS / FAIL.
6. Write the report at `docs/reviews/YYYY-MM-DD_<topic>_evaluation.md`
   following `references/report-template.md`.

## Boundaries

- The evaluator does not generate, repair, or refactor convention surfaces.
  It judges.
- No machine-readable output. The markdown report is the interface.
- Operational conventions (`.work/`, `.wiki/`) are scored through the
  three dimensions, not separate ones.
