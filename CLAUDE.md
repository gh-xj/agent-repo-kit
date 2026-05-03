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

The `work` CLI now lives in its own repo: https://github.com/gh-xj/work-cli.
The `work-cli-qa` skill (the source-side release gate) moved with it. ARK
keeps the `work-cli` skill — the operator-facing playbook — and the adapter
mirrors. ARK's `task verify` installs the external `work` binary (or
hard-fails) and asserts `min_work_version` from `.conventions.yaml`.

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Work** — local-first work tracker at `.work/`. The CLI is the external
  [`work-cli`](https://github.com/gh-xj/work-cli); install with
  `go install github.com/gh-xj/work-cli/cmd/work@latest` (or via the release
  tarball). Drive it through `task work -- ...`. Local state lives in the
  ignored `.work/config.yaml` and `.work/items/*.yaml`. Daily commands:
  `task work -- inbox`, `task work -- inbox add "title"`,
  `task work -- triage accept IN-0001`, `task work -- view ready`, and
  `task work -- show W-0001`.
- **Conventions descriptor** — `.conventions.yaml` at the repo root declares
  which conventions this repo opts into, including `min_work_version` for
  the external `work` CLI. Read by the convention-engineering skill for
  bootstrap and audit; enforced by `scripts/verify.sh`.
