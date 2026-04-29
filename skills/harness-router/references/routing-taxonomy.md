# Routing Taxonomy

Use these dimensions to classify each possible learning before choosing a
destination.

## Dimensions

| Dimension | Values | Question |
| --- | --- | --- |
| Authority | `required`, `recommended`, `observation`, `preference`, `hypothesis` | How binding is this learning? |
| Scope | `global`, `org`, `repo`, `path`, `skill`, `work_item`, `session` | Where should it apply? |
| Durability | `temporary`, `active_work`, `stable_project`, `reusable_procedure`, `long_term` | How long should it matter? |
| Load policy | `always`, `path_scoped`, `on_demand`, `retrieval`, `audit_only` | When should an agent see it? |
| Consumer | `human`, `agent`, `harness`, `ci`, `hook`, `eval_runner`, `future_researcher` | Who or what needs it? |
| Enforcement | `guidance`, `hook`, `test`, `lint`, `policy`, `manual_review` | Is prose enough? |
| Sensitivity | `public`, `private`, `secret_risk`, `untrusted` | What handling risk exists? |
| Provenance | source, date, confidence, reviewer status | Can a future agent verify it? |
| Cost | bytes, tools, latency, maintenance burden | What does loading or maintaining it cost? |

## Promotion Criteria

Promote a session learning only when at least one is true:

- It corrects a repeated agent failure.
- It describes a stable project rule or invariant.
- It changes a reusable workflow.
- It prevents a concrete future bug, safety issue, or verification miss.
- It captures a decision that future agents need to continue work.

Keep it out of durable surfaces when:

- It is a one-off preference with no expected reuse.
- The source is untrusted and not corroborated.
- It duplicates an existing rule or reference.
- The target surface would become broader than the learning's scope.
- The learning should become a test, hook, or eval instead of prose.

## Confidence

Use `high` when the source is authoritative, repeated, and has a clear target.
Use `medium` when the learning is useful but still needs human review or target
selection.
Use `low` when it is speculative, weakly sourced, or potentially transient.

