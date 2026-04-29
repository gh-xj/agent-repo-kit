# CLAUDE.md

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Work** — local-first work tracker at `.work/`. The repo-local CLI is
  exposed through `task work -- ...`; canonical state lives in
  `.work/config.yaml`, `.work/views.yaml`, and `.work/items/`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`, `task work -- triage accept IN-0001`,
  `task work -- view ready`, and `task work -- show W-0001`.
- **Wiki** — LLM-maintained knowledge base at `.wiki/`. Read `.wiki/RULES.md`
  for page types, frontmatter, and citation rules. Validate with
  `task wiki:lint` (or `task -d .wiki lint`).
- **Verification** — run `task verify` from the repo root to execute the full
  demo repo verification gate.
- **Tracked contract** — `.convention-engineering.json` is the machine-readable
  convention contract for this repo.

Conventions are owned by `agent-repo-kit`'s `convention-engineering/`
surface; templates live under that kit's
`convention-engineering/references/templates/` directory.
