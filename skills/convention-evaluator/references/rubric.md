# Rubric

Three dimensions. All hard-fail. The point is skeptical external judgment of
how well a repo lives up to its declared `.conventions.yaml`, not checklist
self-approval.

## Dimensions

### `legibility`

Can an agent in fresh context navigate this repo and the conventions it
declares?

- Are `agent_docs` files present and structured per
  `convention-engineering/references/core/agent-knowledge.md`?
- Is the docs taxonomy unambiguous (one `docs_root`, the declared subfolders
  exist, files follow the date-prefixed naming)?
- Are mirrored contract files (`AGENTS.md` / `CLAUDE.md`) actually coherent,
  not divergent?
- When `.conventions.yaml` declares an opt-in, can the reader find the
  artifact it points at without grep?

### `enforceability`

Do the declared opt-ins actually constrain behaviour, or are they
aspirational?

- For each declared `agent_docs` / `docs_root` / `taskfile` / `pre_commit` /
  `skill_roots` / `operations` opt-in: the corresponding artifact exists,
  is non-trivial, and would fail loudly if removed.
- For each entry under `checks:`: the rule is testable from a fresh
  checkout, and the repo currently satisfies it.
- For each adopted operation (`work`, `wiki`): the directory exists, the
  agent contracts contain the pointer snippet, the op-specific health
  check actually runs.

### `verification`

Is there one canonical command that exercises the declared gates and
produces debuggable evidence?

- `task verify` (or the repo's equivalent) exists, exits 0 from a clean
  checkout, and exercises every declared opt-in.
- Failures emit per-step exit codes, log paths, and short tail excerpts —
  not one monolithic output blob.
- Smoke or regression coverage exists when the repo's risk warrants it
  (CLI-shaped repos, codegen-shaped repos).

## Score Scale

- `0`: absent or actively broken
- `1`: weak, mostly aspirational
- `2`: partial or inconsistently implemented
- `3`: acceptable and operational
- `4`: strong, coherent, and well-enforced

## Threshold Policy

All three dimensions are hard-fail. Default: each must score `>= 3`.

High-risk repos (see below): each must score `>= 4` on `enforceability`
and `verification`. `legibility >= 3` still passes.

## High-Risk Criteria

Treat the run as high-risk when **either** is true:

- the repo has three or more consumers or downstream dependents
- the repo acts as a shared template or policy source for other repos

If neither applies, default thresholds.

## Verdict

The evaluation passes when every dimension meets its threshold. Otherwise
it fails — there is no soft-fail tier. The report should always say
`PASS` or `FAIL` plus the per-dimension scores; status sits at the top
of the markdown.
