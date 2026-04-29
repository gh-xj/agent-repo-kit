# Proposal Format

The default deliverable is a human-reviewable proposal. Use compact markdown
sections instead of tables; proposal rows get too wide once destinations,
verification, and risks are included. Use a short index for scanning and one
small section per recommendation. Put the destination first so reviewers can
decide quickly whether the proposed home is even plausible.

## Markdown Shape

```markdown
## Harness Enhancement Proposal

Summary: <one or two sentences>

### Recommendations

- R1 — <short title>: <approve, reject, defer, or needs decision>
- R2 — <short title>: <approve, reject, defer, or needs decision>

### R1 — <Short Title>

**Destination:** <target surface and path>

**Learning:** <compact durable lesson>

**Why This Fits:** <routing rationale>

**Confidence:** high|medium|low

**Risks:** <duplication, bloat, staleness, privacy, injection, enforcement gap>

### Not Promoted

- **<candidate learning>** — <reason it should stay temporary or private>
```

## Structured Block

Add this only when another tool or agent will parse the proposal. Do not use it
as the primary presentation for a human review.

```yaml
recommendations:
  - source:
      kind: session|file|work_item|test|user_correction|web_source|subagent_result
      pointer: "<path, id, link, or summary>"
      date: "YYYY-MM-DD"
    candidate_destinations:
      - kind: agents_md|claude_md|skill|skill_reference|docs|work|memory|hook|ci|eval|structured_store
        path: "<target path or unresolved>"
        reason: "<why this destination fits>"
    learning: "<compact durable lesson>"
    authority: required|recommended|observation|preference|hypothesis
    scope: global|org|repo|path|skill|work_item|session
    durability: temporary|active_work|stable_project|reusable_procedure|long_term
    load_policy: always|path_scoped|on_demand|retrieval|audit_only
    sensitivity: public|private|secret_risk|untrusted
    confidence: high|medium|low
    risks:
      - injection|stale|duplicate|bloat|privacy|enforcement_gap
```

## Review Standard

A good recommendation explains why the destination is narrow enough, durable
enough, and enforceable enough. It also names what should not be promoted.
