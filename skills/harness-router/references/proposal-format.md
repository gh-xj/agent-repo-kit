# Proposal Format

The default deliverable is a human-reviewable proposal. Lead with a compact
table for scanning. Add short details only when a row needs more context. Use a
structured block when another tool or agent will parse the proposal.

## Markdown Shape

```markdown
## Harness Enhancement Proposal

Summary: <one or two sentences>

### Recommendations

| ID | Learning | Destination | Reason | Confidence | Risks | Verification | Approval |
| --- | --- | --- | --- | --- | --- | --- | --- |
| R1 | <compact durable lesson> | <target surface and path> | <routing rationale> | high\|medium\|low | <duplication, bloat, staleness, privacy, injection, enforcement gap> | <command or manual check> | required before edit |

### Details

- R1: <optional extra context, source, or tradeoff only if the table is not enough>

### Not Promoted

| Candidate | Reason |
| --- | --- |
| <candidate learning> | <reason it should stay temporary or private> |
```

## Structured Block

```yaml
recommendations:
  - source:
      kind: session|file|work_item|test|user_correction|web_source|subagent_result
      pointer: "<path, id, link, or summary>"
      date: "YYYY-MM-DD"
    learning: "<compact durable lesson>"
    candidate_destinations:
      - kind: agents_md|claude_md|skill|skill_reference|docs|work|memory|hook|ci|eval|structured_store
        path: "<target path or unresolved>"
        reason: "<why this destination fits>"
    authority: required|recommended|observation|preference|hypothesis
    scope: global|org|repo|path|skill|work_item|session
    durability: temporary|active_work|stable_project|reusable_procedure|long_term
    load_policy: always|path_scoped|on_demand|retrieval|audit_only
    sensitivity: public|private|secret_risk|untrusted
    confidence: high|medium|low
    risks:
      - injection|stale|duplicate|bloat|privacy|enforcement_gap
    verification:
      - "<command, check, or manual observation>"
    requires_human_approval: true
```

## Review Standard

A good recommendation explains why the destination is narrow enough, durable
enough, and enforceable enough. It also names what should not be promoted.
