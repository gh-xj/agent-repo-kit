# Contract Schema

The repo's convention contract is the `.conventions.yaml` file at the repo
root. See `skills/convention-engineering/references/core/conventions-yaml.md`
for the descriptor's keys and reader semantics.

Until this skill is redesigned (tracked under the convention-evaluator
follow-up to the convention-engineering refactor), evaluators score against
the YAML descriptor as a whole rather than against individual fields:

- The descriptor exists at the repo root.
- Each declared opt-in (`agent_docs`, `docs_root`, `taskfile`, `pre_commit`,
  `skill_roots`, etc.) maps to one or more concrete artifacts that must
  exist or behaviors that must run.
- Each entry under `checks:` is interpreted by the evaluator and counts as
  a separate scored claim.

A field-by-field schema for evaluator dimensions is pending the rewrite.
