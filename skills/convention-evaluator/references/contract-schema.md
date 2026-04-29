# Contract Schema

`convention-evaluator` owns the meaning of the top-level convention contract fields. `convention-engineering` chooses repo-specific values, but it does not define schema semantics or thresholds.

## Required Top-Level Fields

- `contract_version`
- `mode`
- `profiles`
- `docs_root`
- `ownership_policy`
- `mirror_policy`
- `evaluation_inputs`
- `chunk_plan`

Optional deterministic fields remain part of the machine contract:

- `required_files`
- `taskfile_checks`
- `canonical_pointer_mode`
- `canonical_pointers`
- `content_checks`
- `git_exclude_checks`
- `invariant_contract`

## Versioning Rules

- `contract_version` is required.
- version `1` is the current supported major.
- additive fields are allowed within a major version.
- planner, checker, orchestrator, and evaluator fail closed on an unknown major version.
- defaults apply only where the schema explicitly defines them.

## Field Semantics

### `mode`

- `tracked`: repo-owned contract at `.convention-engineering.json`
- `overlay`: local-only contract at `.docs/convention-engineering.overlay.json`

### `docs_root`

Allowed values:

- `docs`
- `.docs`

### `ownership_policy`

Defines ownership boundaries for:

- portable agent-tooling authoring
- domain knowledge
- repo-local agent-tooling placement

This field is part of the evaluator’s `ownership_clarity` judgment, not just a routing hint.

### `mirror_policy`

Declares the repo’s mirrored-doc or canonical-doc policy. The evaluator uses this for drift analysis and document consistency checks.

### `evaluation_inputs`

Evaluator signal object. The minimum supported field is:

- `repo_risk`

`repo_risk` is evaluator input, not caller-owned threshold policy.

### `chunk_plan`

Chunk-aware convention runs must use:

- `enabled`
- `chunks[]`

Each chunk record is:

```json
{
  "id": "agent-legibility",
  "scope": "agent docs and pointer policy",
  "completion_criteria": ["CLAUDE.md and AGENTS.md policy documented"],
  "depends_on": []
}
```

`chunk_plan.chunks` is ordered and dependency-aware. Evaluator logic may use it to distinguish passed, deferred, and pending work.

## Ownership Rules

- `convention-evaluator` owns schema semantics, score scale, thresholds, and result status meanings.
- `convention-engineering` owns the contract values chosen for a repo run.
- A repo-local agent-tooling authoring convention (if the host runtime exposes one) may author artifacts referenced by the contract when repo-local tooling is justified, but it does not define the contract schema.

## Threshold Rule

`convention-engineering` may pass `evaluation_inputs.repo_risk`, and future evaluator-defined named modes may raise scrutiny. It may not lower evaluator thresholds through the contract.

## Operational Conventions

`convention-engineering` scaffolds operational conventions — currently `work` (`.work/`) and `wiki` (`.wiki/`) — by creating tracked local files plus a pointer in `CLAUDE.md` and `AGENTS.md`. The evaluator scores adoption through existing deterministic fields; there is no dedicated `operations` schema key.

Expected contract expressions when a repo claims adoption:

- `required_files`: e.g., `.work/config.yaml`, `.work/views.yaml`, `.wiki/RULES.md`, `.wiki/Taskfile.yml`.
- `taskfile_checks`: e.g., `work:check:`, `task -d .wiki lint` (or `task wiki:lint` when wired into the root Taskfile).
- `content_checks`: a `## Conventions` pointer snippet must appear in both `CLAUDE.md` and `AGENTS.md`. This is a `grep`-level presence check, not a parse.
- `invariant_contract`: e.g., `.work/events.jsonl` append-only when work history checks are enabled, `.wiki/raw/` immutability.

Missing or aspirational claims map to existing rubric dimensions (see `rubric.md`). Do not invent new top-level fields for operational conventions.
