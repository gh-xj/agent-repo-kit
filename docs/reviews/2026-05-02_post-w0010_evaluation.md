# Convention Evaluation — `agent-repo-kit` (2026-05-02)

**Status:** PASS
**High-risk:** yes — this repo publishes conventions for other repos to adopt; it acts as a shared template/policy source.
**Thresholds applied:** legibility >= 3, enforceability >= 4, verification >= 4.

## Scores

| Dimension      | Score (0-4) | Threshold | Verdict |
| -------------- | ----------- | --------- | ------- |
| legibility     | 3           | 3         | ✓       |
| enforceability | 4           | 4         | ✓       |
| verification   | 4           | 4         | ✓       |

## Findings

### legibility — 3

`AGENTS.md` and `CLAUDE.md` exist with the required sections (entry points,
editing rules, mirrored `## Conventions` block). The `## Conventions`
sections are byte-identical (`diff` empty between
`AGENTS.md:64-` and `CLAUDE.md:25-`).

`.conventions.yaml` is well-commented and points at the schema doc
explicitly (line 3 references
`skills/convention-engineering/references/core/conventions-yaml.md`).

The docs taxonomy is unambiguous: `docs/{requests,planning,plans,implementation,taxonomy}` all exist; `reviews/` was created in this evaluation pass.

Why not 4:

- `docs/` contains two undeclared subdirs (`specs/`, `human-requests/`) that
  do not appear in `.conventions.yaml` or `references/core/docs-taxonomy.md`. Either declare them or
  fold/rename to a declared folder.
- `core-belief.md` is referenced from `AGENTS.md:8` and `CLAUDE.md:8` but
  has no freshness contract — the file dates from before the W-0010
  refactor and may already be partly stale relative to the new
  `.conventions.yaml`-centric world.

### enforceability — 4

Every declared opt-in maps to a real artifact:

- `agent_docs: [AGENTS.md, CLAUDE.md]` — both present.
- `docs_root: docs` + canonical subdirs — present.
- `taskfile: true` — `Taskfile.yml:4` declares `verify`.
- `pre_commit: false` — matches reality (no `.githooks/`, no
  `core.hooksPath`).
- `skill_roots: [.claude/skills, .codex/skills]` — both directories exist.
- `operations: [work]` — `.work/config.yaml` exists; `task work -- view ready --json` succeeds.

The structural opt-ins are **machine-enforced** by `scripts/verify.sh`,
which runs in `task verify` (`Taskfile.yml:8`). The script reads
`.conventions.yaml` via `yq` and asserts each declared artifact exists.
This is the difference between a 3 and a 4 — opt-ins are not just
declared; they fail loudly if removed.

Free-form `checks:` entries remain agent-interpreted (intentional, per
the descriptor's design). `scripts/verify.sh` correctly stops at the
typed keys.

### verification — 4

`task verify` exists (`Taskfile.yml:4-8`), exits 0 from a clean checkout
(verified just now), and exercises every declared gate:

- `task -d cli ci` → lint + test + build + smoke against the work CLI.
- `bash scripts/verify.sh` → opt-in artifact assertions.

Both surfaces emit per-step output and exit codes; failures are debuggable
without re-running. The smoke task validates JSON parseability of `work`
CLI output via `jq -e .` (`cli/Taskfile.yml`), giving regression coverage
for the canonical CLI shape.

CI mirrors local: `.github/workflows/ci.yml` runs `task -d cli ci` on
push/PR.

## What would move this from N to N+1

- **legibility 3 → 4:** declare `specs/` and `human-requests/` in
  `.conventions.yaml` (or absorb their contents into `requests/` /
  `planning/`). Add a freshness contract to `docs/core-belief.md` or
  consolidate it into `AGENTS.md`.

`enforceability` and `verification` are at 4. Further movement requires
introducing a stronger enforcement model (e.g. machine-checked free-form
`checks:` rules), which is out of scope for this rubric.

## Notes

The descriptor's `operations:` key was added during this same epic
(W-0010 Phase 3.5). It is exercised in this evaluation: `work` is
declared, `.work/` exists, and the work CLI runs. No drift between the
schema doc and the actual descriptor.
