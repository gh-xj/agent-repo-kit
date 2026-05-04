# convention-evaluator — Changelog

Semantic versioning. Major bumps change the rubric, threshold model, or
report contract; minor bumps refine guidance or add evidence patterns;
patch bumps clarify wording.

## 1.0.0 — 2026-05-02

First stable release. Breaking against any earlier in-repo state.

- Drop the launcher-receipt + handoff-manifest + machine-readable
  evaluation_result.json machinery (artifacts of the deleted ark
  orchestrator).
- Trim the rubric from 5 dimensions with hard/soft tiers to **3
  hard-fail dimensions**: legibility, enforceability, verification.
- Trim high-risk criteria from 5 to 2 (3+ consumers OR shared template).
- Inputs reduced to the live repo and its `.conventions.yaml`.
- Output is a single markdown report at
  `docs/reviews/YYYY-MM-DD_<topic>_evaluation.md`.
- Adds `references/report-template.md`.
