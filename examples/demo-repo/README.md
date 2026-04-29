# demo-repo

This is a demo of `agent-repo-kit`'s tracked bootstrap applied to a
real-shaped repo: a checked-in convention contract, a docs taxonomy, and
the `.work/` operational convention. It is the
reference a human or agent can point at when asking "what does adoption
actually look like?"

## What's here

- `.convention-engineering.json` — the machine-readable convention contract
  for this tracked example repo.
- `docs/` — the tracked docs taxonomy for requests, planning, plans,
  implementation notes, and taxonomies.
- `.work/` — the local-first work tracker. Canonical state lives in
  `config.yaml` and `items/*.yaml`; built-in views need no extra files.
- `AGENTS.md` / `CLAUDE.md` — dual-written pointer blocks telling any
  agent where the conventions live. **Identical content in both files.**
- `Taskfile.yml` — exposes the `work` wrapper and `task verify` for the
  demo gate.

## Tracked bootstrap outline

1. **Create the tracked contract** — add `.convention-engineering.json`
   and the `docs/` taxonomy to your repo.
2. **Adopt operational conventions as needed** — initialize `.work/` if your
   repo needs tracked work items. Add `.wiki/` only when source-backed
   knowledge pages earn their surface area.
3. **Mirror the pointer block** — paste the `## Conventions` block from
   `AGENTS.md` into both `AGENTS.md` and `CLAUDE.md` at your repo root.
4. **Verify** — run `task verify`.

## Run the tests locally

```bash
task verify             # convention contract + work
task work -- view ready # inspect ready work
```

Both run in CI on every push and PR — see `.github/workflows/ci.yml` at
the top of the `agent-repo-kit` repo.
