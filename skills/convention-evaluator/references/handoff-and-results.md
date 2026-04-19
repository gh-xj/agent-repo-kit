# Handoff And Results

The evaluator runs only from persisted artifacts plus the live repo. No planner or generator scratch context is allowed beyond what the repo and handoff files record.

## Launcher Receipt

`launch-receipt.json` is launcher-owned. It proves the evaluator was launched in fresh context.

Required fields:

- `parent_invocation_id`
- `evaluator_invocation_id`
- `launch_mode`
- `fresh_context`
- `fork_context` when the runtime exposes it
- `handoff_manifest_id`
- `launched_at`

Rules:

- `fresh_context` must be `true`
- `fork_context` must be `false` when present
- missing or invalid launcher metadata is an evaluator infrastructure failure

## Handoff Manifest

`handoff.json` is the bridge from generation to evaluation.

Required fields:

- `manifest_id`
- `contract_path`
- `brief_path`
- `generated_artifact_paths`
- `raw_evidence_bundle_path`
- `launch_receipt_path`
- `chunk_state`
- `requested_scope`

`chunk_state` records should carry:

- `id`
- `status`: `pending`, `in_progress`, `passed`, `rework`, or `deferred`
- `hard_fail_dimensions`
- `soft_fail_dimensions`

Recommended location:

- `<docs_root>/reviews/YYYY-MM-DD_<topic>_handoff.json`

## Raw Evidence Bundle

The raw evidence bundle is evaluator input, not scored output.

Required structure:

- directory beside the evaluation report
- `manifest.json` with stable relative paths
- one record per collected command or tool run with:
  - `name`
  - `command`
  - `cwd`
  - `started_at`
  - `finished_at`
  - `exit_code`
  - `tool_name`
  - `tool_version`
  - `stdout_path`
  - `stderr_path`

The bundle should also include checker output, verify logs, and file references for changed convention surfaces.

## Evaluation Result

`evaluation_result.json` is the machine-readable completion interface.

Required fields:

- `scope`: `chunk` or `final`
- `status`: `passed`, `semantic_failed`, or `infrastructure_failed`
- `chunk_id` when scope is `chunk`
- `hard_fail_dimensions`
- `soft_fail_dimensions`
- `report_path`
- `checker_json_path`
- `launch_receipt_path`
- `generated_at`

Recommended location:

- `<docs_root>/reviews/YYYY-MM-DD_<topic>_evaluation_result.json`

## Scope Rules

- chunk scope evaluates the current chunk and regression-checks already passed chunks
- chunk scope ignores future `pending` chunks
- final scope fails when unresolved deferred findings still affect a hard-fail dimension

## Report Expectations

The markdown evaluation report must:

- show all five dimension scores
- show the thresholds actually applied
- group failed checks by dimension
- include `chunk_id` when scope is `chunk`
- explain whether status is `passed`, `semantic_failed`, or `infrastructure_failed`
