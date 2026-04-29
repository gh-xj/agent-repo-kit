# CLAUDE.md

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Work** — local-first work tracker at `.work/`. The repo-local CLI is
  exposed through `task work -- ...`; canonical state lives in
  `.work/config.yaml` and `.work/items/*.yaml`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`, `task work -- triage accept IN-0001`,
  `task work -- view ready`, and `task work -- show W-0001`.
- **Verification** — run `task verify` from the repo root to execute the full
  demo repo verification gate.
- **Tracked contract** — `.convention-engineering.json` is the machine-readable
  convention contract for this repo.

Conventions are owned by `agent-repo-kit`'s `convention-engineering/`
surface; templates live under that kit's
`convention-engineering/references/templates/` directory.
