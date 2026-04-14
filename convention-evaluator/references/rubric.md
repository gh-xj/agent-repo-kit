# Rubric

The evaluator grades convention work across five dimensions. The point is skeptical external judgment, not checklist self-approval.

New material added to `convention-engineering` (new core principles, artifact
categories, teaching docs) maps onto these five dimensions via the existing
`required_files`, `content_checks`, `taskfile_checks`, and `invariant_contract`
contract fields. Do not add new dimensions or new top-level schema keys per
feature — the evaluator's job is observable behaviour, not belief audit.

## Dimensions

### `legibility`

Questions:

- Are the repo’s convention docs present and navigable?
- Are commands exact enough to run?
- Can an agent find the right convention surface quickly?

### `enforceability`

Questions:

- Do lint, boundary, hook, or invariant surfaces actually constrain behavior?
- Are required files and deterministic checks real rather than aspirational?

### `verification`

Questions:

- Is there one canonical verification surface?
- Does it produce debuggable evidence?
- Are smoke or regression gates present when the repo needs them?

### `drift_resistance`

Questions:

- Are mirrored or canonical docs kept coherent?
- Is the docs root unambiguous?
- Are freshness or update contracts present where drift is likely?

### `ownership_clarity`

Questions:

- Are repo conventions separated from domain knowledge?
- Are generic agent-tooling authoring concerns routed away from repo-local conventions (to a dedicated authoring convention when the host runtime provides one)?
- Are repo-local agent tools used only when local ownership is justified?

## Score Scale

- `0`: absent or actively broken
- `1`: weak and mostly aspirational
- `2`: partial or inconsistently implemented
- `3`: acceptable and operational
- `4`: strong, coherent, and well-enforced

## Threshold Policy

Hard-fail dimensions:

- `enforceability`
- `verification`
- `ownership_clarity`

Soft-fail by default:

- `legibility`
- `drift_resistance`

Default thresholds:

- hard-fail dimensions must score `>= 3`
- soft-fail dimensions must score `>= 2`

High-risk repos raise the soft-fail threshold to `>= 3`.

## High-Risk Criteria

Treat `evaluation_inputs.repo_risk` as high-risk when one or more of these are true:

- the repo has three or more consumers or downstream dependents
- cross-repo changes are common
- deployment, codegen, or harness behavior depends heavily on convention
- the run includes repo-structure or code-layout refactors
- the repo acts as a shared template or policy source

## Failure Interpretation

- checker failures that materially reduce a hard-fail dimension produce `semantic_failed`
- soft-dimension failures remain soft unless `repo_risk` raises the threshold
- unreadable or invalid handoff artifacts produce `infrastructure_failed`

## Operational Conventions (tickets, wiki)

These adopt via manual template copy plus a `## Conventions` pointer snippet in `CLAUDE.md` and `AGENTS.md`. Score them through the existing dimensions — do not add new ones:

- missing `.tickets/` or `.wiki/` directory or required files claimed by the contract → `enforceability`
- pointer snippet absent from `CLAUDE.md` or `AGENTS.md` (grep-level check, not parse) → `legibility`
- task invocations claimed but not runnable (`task -d .tickets test`, `task -d .wiki lint`, `task wiki:lint`) → `verification`
- append-only or immutable surfaces drifted (`.tickets/audit-log.md`, `.wiki/raw/`) → `drift_resistance`
- repo-local agent tooling duplicating what template ownership already covers → `ownership_clarity`

The evaluator judges adoption; it does not generate or repair the scaffolding.
