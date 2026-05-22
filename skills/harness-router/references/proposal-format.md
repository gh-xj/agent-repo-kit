# Proposal Format

The default deliverable is a human-reviewable proposal in a **hybrid layout**:
a narrow index table for scanning, then one section per recommendation for the
detail. Tables are used where each row has the same short fields (the index,
enum quick references, the Not Promoted list); sections are used where any
cell would expand into a path, paragraph, or rationale (the per-recommendation
detail). Destination always comes first.

## Markdown Shape

```markdown
## Harness Enhancement Proposal

Summary: <one or two sentences>

### Recommendations

| #  | Title         | Decision                                  | Destination (kind) |
|----|---------------|-------------------------------------------|--------------------|
| R1 | <short title> | approve / reject / defer / needs decision | skill_reference    |
| R2 | <short title> | ...                                       | docs               |

### R1 — <Short Title>

**Destination:** <target surface and path>

**Externalized Burden:** <continuity | procedure | interaction | governance | observability | planning | evaluation>

**Artifact Class:** <instruction | skill | skill_reference | protocol | docs | work | memory | check | structured_store>

**Proposed Change:** <compact durable action or update>

**Evidence:** <source, date, work item, file, test, or user correction>

**Why This Fits:** <routing rationale>

**Confidence:** high | medium | low

**Risks:** <duplication | bloat | staleness | privacy | injection | enforcement gap>

### Not Promoted

| Candidate Learning | Reason it stays temporary or private |
|--------------------|--------------------------------------|
| <learning>         | <reason>                             |
```

Use the index table even for a single recommendation; it forces destination-kind
up front and keeps multi-item proposals scannable.

## When to Use a Table vs. a Section

| Surface                 | Form    | Why |
|-------------------------|---------|-----|
| Recommendations index   | table   | Same short fields per row; reviewer scans Title + Decision + kind. |
| Per-R<N> detail block   | section | Destination paths, Evidence prose, Why-This-Fits paragraphs overflow cells. |
| Not Promoted            | table   | Two columns, short reason text. |
| Enum / field reference  | table   | Repeated label/value records. |

A row with multi-line cells, code, paths longer than ~40 chars, or a paragraph
is the signal to leave the table and go back to a section.

## Field Quick Reference

| Field                 | Required | Values / Notes |
|-----------------------|----------|----------------|
| Summary               | yes      | One or two sentences. |
| Recommendations index | yes      | Narrow table; one row per R<N>. |
| Destination           | yes      | Target surface + path; **first** in each per-R<N> block. |
| Externalized Burden   | yes      | continuity / procedure / interaction / governance / observability / planning / evaluation |
| Artifact Class        | yes      | instruction / skill / skill_reference / protocol / docs / work / memory / check / structured_store |
| Proposed Change       | yes      | Action verb form; not a description. |
| Evidence              | yes      | Source, date, file, work item, test, or user correction. |
| Why This Fits         | yes      | Routing rationale tying destination to scope, durability, load policy, enforcement. |
| Confidence            | yes      | high / medium / low |
| Risks                 | yes      | duplication / bloat / staleness / privacy / injection / enforcement gap |
| Not Promoted          | yes      | Separate section; explicit list of candidates that stayed temporary/private. |

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
It also names what should not be promoted. The index table must show Decision
and Destination kind so a reviewer can triage without reading every block.
Keep `learning` in the structured block for machine parsing.
