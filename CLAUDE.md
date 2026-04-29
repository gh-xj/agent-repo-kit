# CLAUDE.md — agent-repo-kit

You are inside **agent-repo-kit**. This repo publishes a convention for
_other_ repos to adopt, and it also adopts that same convention on
itself — `ark check --repo-root .` and `task verify` both run here.

## Core belief

Before changing conventions, persistent state models, agent workflows, or
repo-wide architecture, read `docs/core-belief.md`. It is the philosophical
north star for this repo; this file is the operational map.

See `AGENTS.md` for the canonical entry-points map and the rules for
editing the kit-owned surfaces (`skills/`, `adapters/<harness>/`,
`examples/demo-repo/`). The `## Conventions` block below mirrors
AGENTS.md per the `.agent-repo-kit.json` mirror policy; editing it
requires regenerating with `ark skill sync` if it ever drifts.

<!-- agent-repo-kit:init:start -->

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
- **Verification** — run `task verify` from the repo root to execute the
  convention verification gate.
- **Tracked contract** — `.convention-engineering.json` is the
  machine-readable convention contract for this repo.

Conventions are scaffolded by `agent-repo-kit` under `.convention-engineering/`.

<!-- agent-repo-kit:init:end -->
