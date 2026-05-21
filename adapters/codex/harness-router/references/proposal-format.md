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

**Externalized Burden:** <continuity, procedure, interaction, governance, observability, planning, or evaluation>

**Artifact Class:** <instruction, skill, skill_reference, protocol, docs, work, memory, check, or structured_store>

**Proposed Change:** <compact durable action or update>

**Evidence:** <source, date, work item, file, test, or user correction>

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
      - kind: agents_md|claude_md|skill|skill_reference|protocol|docs|work|memory|hook|ci|eval|structured_store
        path: "<target path or unresolved>"
        reason: "<why this destination fits>"
    learning: "<compact durable lesson>"
    externalized_burden: continuity|procedure|interaction|governance|observability|planning|evaluation
    artifact_class: instruction|skill|skill_reference|protocol|docs|work|memory|check|structured_store
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

A good human-facing recommendation starts with the destination, names the
burden and artifact class, states the proposed change as an action, and explains
why that destination is narrow enough, durable enough, and enforceable enough.
It also names what should not be promoted. Keep `learning` in the structured
block for machine parsing.
