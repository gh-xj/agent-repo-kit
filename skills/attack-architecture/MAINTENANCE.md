# attack-architecture Maintenance

## When This File Changes

- A real run changes the skill's triggers, phases, finding schema, lens set, or
  report contract.
- User feedback changes how the skill treats architectural principles or review
  evidence.
- A self-evolution audit item is intentionally skipped.

## Lessons

### 2026-05-17 - Principles Calibrate, Evidence Proves

**Trigger**: A user wanted DRY, KISS, YAGNI, SOLID, modern resilience, and
Well-Architected-style principles incorporated without losing the adversarial
architecture-review workflow.
**Change**: `SKILL.md`, `references/principle-calibration.md`,
`references/lens-prompts.md`, and `references/report-template.md` now treat
principles as calibration and optional finding tags, not as standalone verdicts.
**Expected effect**: Agents connect architecture smells to recognizable
principles while still requiring concrete file:line evidence and a failure mode.
**Falsifier**: Reports start producing acronym-only findings such as "violates
SOLID" or broad cloud-pillar checklists without cited code evidence.
