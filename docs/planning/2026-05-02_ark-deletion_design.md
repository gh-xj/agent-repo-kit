# Ark CLI Deletion / Refactor — Intent

Status: planned, not yet executed.
Tracking: `.work/items/W-0010.yaml` (epic: convention-engineering refactor).

## Why

The `ark` CLI was built to enforce a `.convention-engineering.json` contract,
scaffold convention files, and orchestrate evaluator handoffs. The convention-
engineering skill is being refactored to drop that contract entirely (replaced
by a small `.conventions.yaml` opt-in descriptor) and to defer the evaluator
integration. Once those decisions land, `ark` is left holding only:

- a checker for a config file that no longer exists,
- a scaffolder for a doc taxonomy that the agent can stamp out from a few-line
  YAML descriptor,
- an orchestrator for a deferred evaluator pipeline,
- a tasklint for stack profiles that are being deleted.

In short: every justification for `ark` is being removed in this epic.

## What gets deleted

Confirmed for deletion:

- `cli/cmd/ark/` — binary entry.
- `cli/internal/arkcli/` — command tree.
- `cli/internal/contract/` — `.convention-engineering.json` checker.
- `cli/internal/scaffold/` — `ark init`.
- `cli/internal/orchestrator/` — `ark orchestrate`.
- `cli/internal/tasklint/` — Taskfile linter scoped to deleted profile rules.

Conditional / pending evaluation (do not delete in this epic without checking
callers from `cli/cmd/work` and other entry points first):

- `cli/internal/evaluator/` — convention-evaluator runner. Tied to deferred
  Q14. Recommended: delete now (dead code rots), rebuild fresh later if/when
  the evaluator skill returns. Decision pending until ark cleanup runs.
- `cli/internal/skillbuilder/`, `cli/internal/skillsync/`,
  `cli/internal/upgrade/` — must be inspected for reuse by `cli/cmd/work` or
  other surfaces. Delete only if exclusively `ark`-owned.

Out of scope:

- `cli/cmd/work/` and `cli/internal/work*/` — the work CLI stays.
- `scripts/link-dev-skills.sh`, `scripts/tag-release.sh` — repo-level shell
  scripts, unrelated to `ark`.

## What replaces it

Nothing binary. The replacement pieces live in the convention-engineering
skill itself:

- A `.conventions.yaml` schema sketch (in
  `skills/convention-engineering/references/core/conventions-yaml.md`).
- A `task verify` target the repo defines, which reads `.conventions.yaml`
  and verifies declared items.
- Bootstrap and audit workflows reduced to "create / verify the YAML
  descriptor and the files it declares."

## Sequencing

1. Land the convention-engineering skill cleanup first (W-0010 phases 1–3).
   The skill stops referencing `ark`.
2. After the skill is clean, do the `ark` deletion as a separate, isolated
   change set:
   - Inspect `cli/internal/{evaluator,skillbuilder,skillsync,upgrade}` for
     callers outside `cmd/ark`.
   - Delete the confirmed packages above.
   - Resolve any compile errors in `cli/cmd/work` and other entry points.
   - Update `cli/Taskfile.yml`, `install.sh`, and root README/CLAUDE.md/
     AGENTS.md to drop `ark` references.
3. Commit `ark` deletion as a single change set so it is easy to revert if a
   downstream consumer surfaces.

## Risk / Rollback

- The repo currently advertises `ark` in `CLAUDE.md`, `README.md`, and
  `install.sh`. Any external automation invoking `ark` will break.
- Rollback: revert the deletion commit; the skill cleanup commit can stay.

## Open Questions Deferred Past This Epic

- Should the YAML descriptor ever be machine-checked by _some_ runner, or is
  "agent reads the YAML and runs the checks" sufficient indefinitely?
- If/when `convention-evaluator` returns, does it ride on top of the YAML
  descriptor without re-introducing a separate contract layer?
