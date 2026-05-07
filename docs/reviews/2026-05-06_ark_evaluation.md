# Convention Evaluation — `agent-repo-kit` (2026-05-06)

**Status:** FAIL
**High-risk:** yes — ARK is a shared template / policy source for other repos (work-cli-epic, work-cli, and presumably any other adopter), so the rubric's "shared template" criterion applies.
**Thresholds applied:** legibility ≥ 3, enforceability ≥ 4, verification ≥ 4.

Run from a fresh-context dogfood pass after the convention-engineering 2.1 fold of convention-evaluator. The agent that wrote 2.0/2.1 (this one) scoring it: not the rubric's intended posture, but documented for follow-up.

## Scores

| Dimension      | Score (0-4) | Threshold | Verdict |
| -------------- | ----------- | --------- | ------- |
| legibility     | 3           | 3         | ✓       |
| enforceability | 3           | 4         | ✗       |
| verification   | 3           | 4         | ✗       |

## Findings

### legibility — 3

`AGENTS.md` is well-structured per `references/recipes/agent-knowledge/pattern.md` (architecture pointer at line 13–46, editing rules at line 48–62, testing at line 64–71, conventions block at line 73–88). `docs/` has all five canonical taxonomy subdirs (requests/, planning/, plans/, implementation/, taxonomy/). `.conventions.yaml` opt-in artifacts are findable at predictable paths. The recently-added `docs/core-belief.md` is referenced from `AGENTS.md` line 10 and exists.

Two real legibility costs:

1. **Stale 1.x path references** in two places after the 2.0 refactor: `.conventions.yaml` line 4 still points at `skills/convention-engineering/references/core/conventions-yaml.md` (now `concepts/conventions-yaml.md`); `README.md` line 88 still points at `references/operations/bootstrap-workflow.md` (now `references/meta/bootstrap-workflow.md`). Both are dead links. The 2.0 refactor itself did not sweep these; they predate this evaluation.

2. **`AGENTS.md` ↔ `CLAUDE.md` relationship is "redirect", not "mirror".** `CLAUDE.md` (verified via `diff AGENTS.md CLAUDE.md`) replaces the entry-points and rules sections with `See AGENTS.md for ...; This file mirrors AGENTS.md; keep both in sync by hand.`, then keeps a Conventions block. The descriptor's check #1 (`.conventions.yaml` line 31) reads "AGENTS.md and CLAUDE.md cover entry points, editing rules, and the same conventions block." — a defensible reading is "redirect counts as covering"; a stricter reading is "cover means contain." Either way, an outside reader inheriting the kit could see this and conclude the mirror invariant is partially maintained.

### enforceability — 3

`scripts/verify.sh` cleanly asserts each declared typed opt-in: `agent_docs` files exist, `docs_root` taxonomy subdirs exist, `Taskfile.yml` has a `verify:` target, `skill_roots` if declared, `operations.work` (`work` binary on PATH + `.work/config.yaml` exists), and the `min_work_version` floor. `bash scripts/verify.sh` exits `verify: opt-ins ok` on this checkout. `task verify` adds the adapter drift gate (`bash scripts/sync-adapters.sh --check`) as a second step, so the previously agent-only check #5 is now mechanical.

Three enforceability gaps:

1. **`min_work_version: "0.1.0"` is silently bypassed by dev builds.** `work --version` returns `work dev` here. `verify.sh` line 72 takes `awk '{print $NF}'` of the version output — `dev`. The `sort -V | head -n1` comparison treats `0.1.0` as lexically less than `dev`, so the check passes. The maintainer's own checkout (running unreleased `work`) is the exact case the gate is designed to catch — and silently passes. The gate is an ornament for the maintainer; it only bites adopters who run a tagged release at all.

2. **Check #7 (deleted-CLI sweep) remains agent-only.** `.conventions.yaml` line 37 says "No tracked file references the deleted `ark` CLI or `.convention-engineering.json` (sweep gate)." A grep against the live worktree could be promoted to a verify.sh assertion in three lines; until then, the rule rots.

3. **`docs/core-belief.md` exists but is not declared as an opt-in.** ARK has a single `docs/core-belief.md` referenced from `AGENTS.md` line 10, but no `invariants.md` or `anti-patterns.md` companions, and no `core_beliefs:` block in `.conventions.yaml`. The file is load-bearing per `AGENTS.md` (`philosophical north star for this repo`) but its existence is not gate-enforced. Either adopt the full trio (the recipe ARK itself ships) or drop the philosophical-north-star claim until you do.

### verification — 3

`task verify` is the canonical entry. It composes two steps as separate cmd lines (`scripts/verify.sh` + `scripts/sync-adapters.sh --check`), so per-step exit codes are visible. CI runs it on every push and PR per `AGENTS.md` line 70. `bash scripts/verify.sh` produces `verify: <category>: <what>` lines on failure (debuggable, top-to-bottom readable) and a single `verify: opt-ins ok` on success.

Two verification gaps that matter at the high-risk threshold:

1. **No log paths or failure-tail format.** A failing step prints its category but not where to look for diagnostic output. For a one-shot bash gate with five short blocks this is borderline acceptable; for a high-risk shared-template repo where adopters' CI failures need to be self-explaining, the bar is higher.

2. **No machine-readable summary.** `summary.json` (or equivalent) for downstream consumers is absent. The convention-engineering rubric does not require it (it's "Optional") but it's the difference between 3 and 4 on this dimension at the high-risk threshold.

## What would move this from N to N+1

**legibility 3 → 4.** Three small edits:

- Fix `.conventions.yaml` line 4: `references/core/conventions-yaml.md` → `references/concepts/conventions-yaml.md`.
- Fix `README.md` line 88: `references/operations/bootstrap-workflow.md` → `references/meta/bootstrap-workflow.md`.
- Either add a check to `.conventions.yaml` declaring `CLAUDE.md` as a redirect-not-mirror sibling (so the relationship is auditable), or actually mirror them.

**enforceability 3 → 4.** Three edits:

- In `scripts/verify.sh` around line 70, reject any `cur` that does not match the semver pattern (the schema's `pattern` for `min_work_version`). A dev build should fail closed, not silently pass.
- Promote check #7 to a grep assertion: `grep -rn "ark\b\|\.convention-engineering\.json"` against tracked files, scoped to known types — three lines.
- Either declare `core_beliefs: true` (with `path: docs/`) and ship the trio, or remove the "core belief" reference from `AGENTS.md` until you do. The current state declares a load-bearing artifact whose existence is not gated.

**verification 3 → 4.** Two edits:

- On failure, emit a one-line "see <log path>" for each failing step. For shell-only gates that just means tee'ing the relevant block to a logfile under `.work/spaces/` or similar.
- Optionally emit `summary.json` with `{step, status, log}` per gate. Downstream CI dashboards can render it without parsing prose.

## Notes

The evaluation rubric was just landed (2026-05-06, version 2.1.0) and this is its first dogfood. The fact that it FAILs the very repo that authors it is a healthy signal — a rubric that PASSed everything would be useless. The three findings above are concrete and small (~10 lines of diff each); none would change the design of ARK, only tighten gates.

The rubric's "agent that wrote the conventions should not score them" rule was technically violated by this run. Re-running from a fresh context (different agent session, clean state) would carry more weight. Take this report as a starting list of gaps; treat its scores as the author's self-assessment until an independent run replaces it.
