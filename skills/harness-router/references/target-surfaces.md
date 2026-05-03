# Target Surfaces

Choose the narrowest durable surface that matches the learning's authority,
scope, load policy, and enforcement need.

## Surface Guide

| Surface | Use For | Avoid When |
| --- | --- | --- |
| `AGENTS.md` | Mandatory repo-wide rules, verification gates, persistent project norms, pointers to deeper artifacts | The content is bulky, path-scoped, experimental, or only relevant to one harness |
| `CLAUDE.md` | Claude-specific entry point, imports, or adapter guidance | It would duplicate `AGENTS.md` without a harness reason |
| Related skill | Reusable workflow or judgment pattern that should trigger on demand | The learning is just project context or a one-off note |
| Used skill | Corrections to the skill's trigger, workflow, references, or guardrails discovered while using it | The correction belongs to a different owner skill |
| Skill reference | Bulky examples, rubrics, schemas, source inventories, or workflow details | The content must be always loaded |
| Protocol / contract | Command grammar, JSON/schema contract, adapter lifecycle, work lifecycle, generated-surface rule, permission boundary | It is only a human explanation or one-off workflow |
| Repo docs | Durable explanation, decision record, design rationale, research synthesis, migration note | The content is operational follow-up or private memory |
| `.work` item/space | Active follow-up, unresolved decision, research capture, implementation plan; read the target space's local rules before writing | The work is completed and should become stable docs instead |
| Memory | Personal preference, recurring local context, low-authority recall hint | Required behavior, shared project knowledge, or secret-bearing content |
| Hook/settings/CI/lint/eval | Deterministic enforcement, safety gate, regression prevention, measurable behavior | A human judgment call is still required |
| Structured local store | Inventories, traces, provenance logs, routing decisions, stale hashes, load evidence | No command reads/writes/validates the data yet |

## Default Routing Matrix

| Learning Type | Primary Destination | Verification |
| --- | --- | --- |
| Required repo rule | `AGENTS.md` | repo convention gate, docs parity checks |
| Harness-specific load behavior | harness adapter file | harness load check if available |
| Repeated workflow | skill or skill reference | skill trigger/audit check |
| Correction to used skill | used skill source | skill audit and reference existence checks |
| Interaction contract | protocol owner, schema, CLI docs, adapter rule, Taskfile, or convention docs | contract check, CLI smoke, or convention gate |
| Bulky durable knowledge | docs or research page | docs taxonomy/convention check |
| Active follow-up | `.work` work item/space | `work show` or repo wrapper |
| Personal preference | memory | manual review; never sole authority |
| Hard guarantee | hook, CI, lint, or eval | failing/passing gate |
| Untrusted external claim | raw capture first | source, date, and confidence review |
| Temporary state | session/workspace files | no promotion unless reused |

## Duplication Rule

Do not copy the same rule into multiple always-loaded surfaces. Prefer a single
canonical owner plus imports, pointers, generated adapters, or references.
