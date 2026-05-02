# CLAUDE.md — agent-repo-kit

You are inside **agent-repo-kit**. This repo publishes a convention for
_other_ repos to adopt, and it also adopts that same convention on
itself — `task verify` runs here.

## Core belief

Before changing conventions, persistent state models, agent workflows, or
repo-wide architecture, read `docs/core-belief.md`. It is the philosophical
north star for this repo; this file is the operational map.

See `AGENTS.md` for the canonical entry-points map and the rules for
editing the kit-owned surfaces (`skills/`, `adapters/<harness>/`).
This file mirrors `AGENTS.md`; keep both in sync by hand.

The `work-cli-qa` project skill is a self-evolving release gate for this
repo's source `work` CLI. When a chaotic workflow, regression, or missing
scenario appears, enrich the skill playbook, encode deterministic checks in its
harness, run and evaluate the QA ledger, and refine until the new scenario
passes.

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Work** — local-first work tracker at `.work/`. The repo-local CLI is
  exposed through `task work -- ...`; local state lives in the ignored
  `.work/config.yaml` and `.work/items/*.yaml`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`, `task work -- triage accept IN-0001`,
  `task work -- view ready`, and `task work -- show W-0001`.
- **Conventions descriptor** — `.conventions.yaml` at the repo root declares
  which conventions this repo opts into. Read by the convention-engineering
  skill for bootstrap and audit.
