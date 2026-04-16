# demo-repo

This is a demo of `agent-repo-kit`'s tracked bootstrap applied to a
real-shaped repo: a checked-in convention contract, a docs taxonomy, and
the optional `.tickets/` + `.wiki/` operational conventions. It is the
reference a human or agent can point at when asking "what does adoption
actually look like?"

## What's here

- `.convention-engineering.json` — the machine-readable convention contract
  for this tracked example repo.
- `docs/` — the tracked docs taxonomy for requests, planning, plans,
  implementation notes, and taxonomies.
- `.tickets/` — the flat-file work tracker. State machine in
  `harness/schema.yaml`, verb surface in `README.md`, Taskfile wired up.
- `.wiki/` — the LLM-maintained knowledge base. Page types, frontmatter,
  and citation rules live in `RULES.md`.
- `AGENTS.md` / `CLAUDE.md` — dual-written pointer blocks telling any
  agent where the conventions live. **Identical content in both files.**
- `Taskfile.yml` — includes the wiki and tickets taskfiles and exposes
  `task verify` for the full demo gate.

## Tracked bootstrap outline

1. **Create the tracked contract** — add `.convention-engineering.json`
   and the `docs/` taxonomy to your repo.
2. **Adopt operational conventions as needed** — copy `.tickets/` and
   `.wiki/` only if your repo needs tracked work items or source-backed
   knowledge pages.
3. **Mirror the pointer block** — paste the `## Conventions` block from
   `AGENTS.md` into both `AGENTS.md` and `CLAUDE.md` at your repo root.
4. **Customize and verify** — edit `.tickets/harness/taxonomy.yaml` for
   your project's categories and run `task verify`.

## Run the tests locally

```bash
task verify          # convention contract + tickets + wiki
task -d .tickets test # expect 10/10
task -d .wiki lint    # expect OK
```

Both run in CI on every push and PR — see `.github/workflows/ci.yml` at
the top of the `agent-repo-kit` repo.
