# attack-architecture Evals

Use these lightweight cases when changing trigger wording, the finding schema,
the lens set, or the principle-calibration rules.

## Trigger Cases

### Trigger Case - Architecture Attack

Prompt: "Attack the architecture of `src/billing` and tell me where it will hurt to change."

Expected: trigger

Reason: The user asks for an adversarial architecture review of an existing
subsystem.

### Trigger Case - Data Model Doubt

Prompt: "Is this data model right, or are we making illegal states too easy?"

Expected: trigger

Reason: The user asks for contract and data-model critique, which is an
architecture lens.

### Trigger Case - Bug Fix

Prompt: "This endpoint returns 500 after my last commit; find the bug and patch it."

Expected: no_trigger

Reason: This is a bug hunt on recent behavior, not an architecture attack.

### Trigger Case - Pre-Implementation Design

Prompt: "Help me design a new ingestion service before we write code."

Expected: no_trigger

Reason: The skill reviews existing design; it does not own greenfield
brainstorming.

## Output Cases

### Output Case - Principle-Calibrated Finding

Task: Run the skill on an existing module where one abstraction has one
implementation and no caller-provided variation.

Setup: Use normal depth with L1 selected.

Expected Artifact: Architecture attack report.

Success Criteria:
- The finding cites concrete `file:line` evidence.
- The finding explains the failure mode or change-cost pressure.
- `principle_tags` may include `KISS` or `YAGNI`, but the title and rationale
  do not rely on the acronym as the proof.

Evidence To Inspect:
- Generated report under `.docs/arch-attacks/` or the custom report path.
- Raw Phase 3 finding JSON, if preserved in the run notes.

## Regression Cases

### Regression Case - No Principle Laundering

Prior Failure: Architecture principle lists can tempt agents to emit vague
findings like "violates SOLID" or "not DRY" without evidence.

Expected Corrected Behavior: Principles appear only as calibration or
`principle_tags`; every finding still includes cited evidence and a concrete
failure mode.

Verification: Inspect the report for acronym-only findings and reject any item
whose evidence would not stand without the principle tag.

### Regression Case - Well-Architected Has Six Pillars

Prior Failure: Older summaries sometimes name five AWS Well-Architected pillars
and omit sustainability.

Expected Corrected Behavior: Cloud/platform calibration names operational
excellence, security, reliability, performance efficiency, cost optimization,
and sustainability, while keeping dedicated security review out of scope.

Verification: Inspect `SKILL.md` and `references/lens-prompts.md` after edits.
