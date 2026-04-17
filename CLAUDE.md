# CLAUDE.md

<!-- agent-repo-kit:init:start -->
## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Tickets** — flat-file work tracker at `.tickets/`. Read `.tickets/README.md`
  for the verb surface and `.tickets/harness/schema.yaml` for the state
  machine. Daily commands:
  `task -d .tickets {new|list|transition|close|test}`.
- **Wiki** — LLM-maintained knowledge base at `.wiki/`. Read `.wiki/RULES.md`
  for page types, frontmatter, and citation rules. Validate with
  `task wiki:lint` (or `task -d .wiki lint`).
- **Verification** — run `task verify` from the repo root to execute the
  convention verification gate.
- **Tracked contract** — `.convention-engineering.json` is the
  machine-readable convention contract for this repo.

Conventions are scaffolded by `agent-repo-kit` under `.convention-engineering/`.
<!-- agent-repo-kit:init:end -->
