# convention-engineering — Changelog

Semantic versioning. Major bumps land breaking trigger / schema / routing changes; minor bumps add capability without breaking adopters; patch bumps clarify wording or fix references.

## 2.2.0 — 2026-06-22

- Add the experimental `git-workflow` recipe for agent-operated commit, verify,
  push, bundle, and dirty-tree discipline.
- Add eval cases for local-only leaves, hook wiring drift, post-verify dirty
  generated output, upstream/ahead reporting, and legacy branch exceptions.
- Tighten the reference verifier's `pre_commit` check so a missing configured
  `core.hooksPath` cannot be masked by an unused `.githooks/pre-commit` file.

## 2.1.0 — 2026-05-06

- Fold the `convention-evaluator` sibling skill into this one. Its rubric and report template now live at `references/meta/evaluation-rubric.md` and `references/meta/evaluation-report-template.md`. The "fresh-context skeptical scoring" posture is preserved as a meta concern of the same skill rather than a separate publishing artifact. Two routing rows added to SKILL.md under Meta.
- Reason: the gap-vs-graded distinction is a report-shape difference, not a feature difference. Maintaining a separate skill (with its own SKILL.md, CHANGELOG, adapter mirrors across three harnesses) was overhead without distinct functionality.

## 2.0.0 — 2026-05-06

Breaking restructure. Adopters must remap reference paths (see `MIGRATION.md`).

- Top-level reference layout split into `concepts/`, `archetypes/`, `recipes/`, `meta/`. Old `core/`, `operations/`, `patterns/`, `templates/` directories removed.
- **Archetypes** (whole-repo shapes — `epic-wrapper`, `brain`) become self-contained directories with `pattern.md`, `bootstrap.md`, `migration.md`, `templates/`.
- **Recipes** (stackable conventions — `core-beliefs`, `agent-knowledge`, `docs-taxonomy`, `project-skill-placement`, `verify-script`, `work`, `wiki`) likewise become self-contained directories.
- Pattern docs trimmed: `epic-wrapper.md` 441 → ~250 lines; `brain.md` 481 → ~280 lines. Templates and migration content extracted from each pattern doc into sibling files.
- `verification-gates.md` and `verify-script-pattern.md` (175 lines combined) split: `concepts/verification.md` (short principle, ~40 lines) + `recipes/verify-script/pattern.md` (the bash recipe, ~80 lines).
- Format specs added: `references/archetypes/_format.md`, `references/recipes/_format.md`. Required vs optional sections explicit; "Per-Realm Gate Matrix" no longer mandatory across patterns that don't use realms.
- Status header convention introduced — every archetype and recipe carries `Status` and `Codified` lines under the title.
- `operations:` enum opened in the JSON Schema. Recommended values (`work`, `wiki`) now in the description; project-local recipes are permitted.
- `min_work_version` typed key marked deprecated in description (still functional). Tool-specific version constraints belong in `checks:` or pattern-specific blocks.
- `core_beliefs:` typed key added (introduced in `Unreleased` 1.x; folded into 2.0.0 release).
- `MIGRATION.md` ships at the skill root with the full 1.x → 2.0 path mapping and adapter-mirror sync note.

## 1.0.0 — 2026-05-02

First stable release after the W-0010 refactor. Breaking — no migration path from any earlier in-repo state.

- Drop the `.convention-engineering.json` machine contract entirely.
- Replace with the `.conventions.yaml` opt-in descriptor (recognised keys: `agent_docs`, `docs_root`, `taskfile`, `pre_commit`, `skill_roots`, `operations`, `min_work_version`, `checks`).
- Add JSON Schema at `schemas/conventions.schema.json`.
- Drop stack profiles (`go`, `typescript-react`, `python`, `research-corpus`) — convention is stack-agnostic.
- Drop the `contracts/` reference set, `supply-chain.md`, `architecture-contracts.md`, `error-messages-as-remediation.md`, `open-source-git-exclude-workflow.md`.
- Rewrite SKILL.md (151 → ~50 lines) and the audit + bootstrap workflows around the new descriptor.
- Add `operations:` key for `.work/` and `.wiki/` adoption.
- Add the dev-wrapper-repo bootstrapping use case to the description.
