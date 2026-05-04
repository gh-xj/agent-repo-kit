# convention-engineering — Changelog

Semantic versioning. Major bumps land breaking trigger / schema / routing
changes; minor bumps add capability without breaking adopters; patch bumps
clarify wording or fix references.

## 1.0.0 — 2026-05-02

First stable release after the W-0010 refactor. Breaking — no migration
path from any earlier in-repo state.

- Drop the `.convention-engineering.json` machine contract entirely.
- Replace with the `.conventions.yaml` opt-in descriptor (recognised
  keys: `agent_docs`, `docs_root`, `taskfile`, `pre_commit`,
  `skill_roots`, `operations`, `min_work_version`, `checks`).
- Add JSON Schema at `schemas/conventions.schema.json`.
- Drop stack profiles (`go`, `typescript-react`, `python`,
  `research-corpus`) — convention is stack-agnostic.
- Drop the `contracts/` reference set, `supply-chain.md`,
  `architecture-contracts.md`, `error-messages-as-remediation.md`,
  `open-source-git-exclude-workflow.md`.
- Rewrite SKILL.md (151 → ~50 lines) and the audit + bootstrap
  workflows around the new descriptor.
- Add `operations:` key for `.work/` and `.wiki/` adoption.
- Add the dev-wrapper-repo bootstrapping use case to the description.
