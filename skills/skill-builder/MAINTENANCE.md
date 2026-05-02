# Maintenance

This file is for maintainers of `skill-builder`. Runtime guidance stays in
`SKILL.md` and `references/`; this file keeps the evolution loop explicit.

## Purpose

Keep `skill-builder` focused on one job: turning procedural expertise into a
portable, triggerable, evaluable skill package. It should not become the owner
of all harness evolution, memory, conventions, or eval infrastructure.

## Ownership

- `SKILL.md`: trigger, modes, first pass, non-negotiables, reference index.
- `references/skill-quality.md`: evidence basis, risk tier, expertise checks,
  eval-required triggers, and retirement signals.
- `references/skill-evals.md`: lightweight trigger, output, negative, and
  regression eval case shapes.
- `references/workflows.md`: create, update, audit, and migrate workflows.
- `references/runtime-layout.md`: runtime placement and portable core rules.
- `references/pattern-to-script.md`: when prose should become scripts.
- `references/repo-owned-clis.md`: when stable logic should become a CLI.
- `references/multi-skill-architecture.md`: skill dependency and composition
  safety.
- `references/maintenance.md`: runtime troubleshooting and health indicators.
- `MAINTENANCE.md`: maintainer process and changelog.

## Self-Evolution Loop

Use this loop when `skill-builder` itself needs to improve:

```text
evidence -> failure mode -> candidate skill change -> eval case
-> approved patch -> adapter sync -> verification -> changelog
```

The skill may propose its own evolution, but it must not silently rewrite
itself. For non-trivial changes, preserve the evidence and proposal in a work
item or research space before patching.

## Failure Modes

- `trigger`: the skill activates too often, too rarely, or for the wrong task.
- `packaging`: procedural expertise is stored in the wrong surface.
- `evidence`: the skill is built from weak examples, vibes, or stale research.
- `risk`: the skill makes a risky action easier without stronger review.
- `output`: the skill does not define a concrete deliverable or success signal.
- `eval`: trigger, negative, regression, or output cases are missing.
- `script_boundary`: deterministic logic remains fragile prose.
- `composition`: ownership conflicts with another skill, tool, doc, or check.
- `maintenance`: references drift, adapters go stale, or router bloat appears.

## Narrow Owner Map

- trigger wording, scope, mode routing -> `SKILL.md`
- evidence, risk, expertise, retirement -> `references/skill-quality.md`
- trigger/output/regression cases -> `references/skill-evals.md`
- create/update/audit/migrate steps -> `references/workflows.md`
- runtime placement -> `references/runtime-layout.md`
- script extraction -> `references/pattern-to-script.md`
- CLI promotion -> `references/repo-owned-clis.md`
- dependency or composition safety -> `references/multi-skill-architecture.md`
- troubleshooting and health indicators -> `references/maintenance.md`

Prefer the narrow owner over adding a new file.

## Validation

When `SKILL.md` changes, hand-mirror the relevant fields into the matching
`adapters/<runtime>/skill-builder/SKILL.md` files; there is no automated
sync today.

Before claiming completion:

```bash
task verify
```

If CLI behavior changed, also run:

```bash
task -d cli ci
```

## Changelog

- `2026-05-02`: Added evidence/risk-first skill packaging, lightweight skill
  eval guidance, and this self-evolution maintenance file.
