# Maintenance

This file is for maintainers of `harness-router`. Runtime guidance stays in
`SKILL.md` and `references/`; this file keeps the update discipline explicit.

## Purpose

Keep `harness-router` small, proposal-first, and burden-first. The skill routes
durable learnings; it does not silently mutate instructions, skills, memory,
checks, or repo conventions.

## Ownership

- `SKILL.md`: trigger, scope, first actions, operating loop, output shape.
- `references/externalization-model.md`: burdens, artifact classes, lifecycle
  vocabulary, and the promotion decision test.
- `references/routing-taxonomy.md`: authority, scope, durability, load policy,
  consumer, enforcement, sensitivity, provenance, and cost.
- `references/target-surfaces.md`: destination choices and avoid cases.
- `references/proposal-format.md`: human markdown proposal and optional
  structured block.
- `references/workflow.md`: full operating workflow.
- `MAINTENANCE.md`: maintainer checklist and changelog.

## Update Rules

- Keep `SKILL.md` as the router. Move examples, rubrics, schemas, and long
  explanations into references.
- Add to an existing reference before creating a new one.
- Add a durable concept only when it changes a routing decision, proposal
  shape, target surface, or validation step.
- Keep proposal rendering as readable markdown sections, not wide tables.
- Keep `Destination` first in human-facing recommendations.
- Keep `Externalized Burden` and `Artifact Class` close to the destination.
- Treat lifecycle terms as vocabulary unless a command reads or validates
  stored lifecycle state.
- Preserve the approval boundary: this skill proposes durable mutations; it
  does not apply them without user approval.

## Change Triggers

Update this skill when:

- user corrections reveal repeated routing mistakes
- proposal output is hard to review
- a new durable destination appears repeatedly
- protocol or contract changes are being misrouted as ordinary docs or skills
- skill audits find broken references or router bloat
- adapter sync reports drift after canonical skill edits

## Self-Evolution Loop

Use this loop when `harness-router` itself needs to improve. The skill may
propose its own evolution, but the same approval boundary still applies.

```text
evidence -> failure mode -> candidate learning -> owner surface -> proposal
-> approved patch -> validation -> changelog
```

### 1. Capture Evidence

Collect the smallest evidence needed:

- user correction
- bad or confusing proposal output
- repeated misrouting
- missing destination class
- broken reference or audit finding
- research result that changes the routing model
- verification output

If the evidence is bulky or still unresolved, stage it in a work space before
editing the skill.

### 2. Name The Failure Mode

Classify what failed:

- `trigger`: skill activated too often or not often enough
- `classification`: burden, artifact class, or routing dimensions were wrong
- `destination`: target surface choice was too broad, narrow, or stale
- `format`: proposal output was hard to review
- `governance`: approval, risk, provenance, or safety boundary was unclear
- `verification`: proposed change lacked a meaningful check
- `maintenance`: reference drift, adapter drift, or router bloat appeared

Do not patch the skill until the failure mode is clear.

### 3. Pick The Narrow Owner

Patch the smallest owner surface:

- trigger/scope issue -> `SKILL.md`
- burden or artifact vocabulary -> `references/externalization-model.md`
- routing dimensions -> `references/routing-taxonomy.md`
- destination choice -> `references/target-surfaces.md`
- proposal rendering -> `references/proposal-format.md`
- operating steps -> `references/workflow.md`
- maintainer process -> `MAINTENANCE.md`

Prefer editing an existing owner over adding a new file.

### 4. Propose Before Mutation

For non-trivial changes, write a harness enhancement proposal first. A useful
self-evolution proposal includes:

- destination
- externalized burden
- artifact class
- proposed change
- evidence
- why the owner surface fits
- confidence
- risks

Then get approval before patching instructions, references, checks, memory, or
repo conventions.

### 5. Patch And Validate

After approval, make the smallest patch that fixes the failure. Do not bundle
unrelated routing ideas into the same edit.

Validation should match the touched surface:

- any edit: skill audit
- `SKILL.md` edit: adapter sync and skill check
- repo behavior change: `task verify`
- CLI/check behavior change: `task -d cli ci`

### 6. Record The Lesson

Update the changelog when the change alters the skill's routing model,
proposal format, validation rules, or ownership boundaries. Do not add
changelog entries for typo-only edits.

## Self-Evolution Guardrails

- Do not make `harness-router` a silent self-mutating system.
- Do not promote raw session transcripts into always-loaded skill text.
- Do not add a new taxonomy term unless it changes a real routing decision.
- Do not let `SKILL.md` become the knowledge base.
- Do not let this skill decide how to build skill content; hand off to
  `skill-builder` once the destination is a skill.
- Do not store lifecycle state as data until a command reads or validates it.

## Validation

Before claiming completion:

```bash
task verify
```

When `SKILL.md` changes, hand-mirror the relevant fields into the matching
`adapters/<runtime>/harness-router/SKILL.md` files; there is no automated
sync today.

If CLI behavior changed, also run:

```bash
task -d cli ci
```

## Changelog

- `2026-05-02`: Added burden-first externalization model, protocol/contract
  routing, proposal evidence fields, and this maintenance file.
