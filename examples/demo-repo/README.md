# demo-repo

This is a demo of `agent-repo-kit`'s tickets + wiki conventions applied to
a real-shaped repo. It is the reference a human or agent can point at when
asking "what does adoption actually look like?"

## What's here

- `.tickets/` — the flat-file work tracker. State machine in
  `harness/schema.yaml`, verb surface in `README.md`, Taskfile wired up.
- `.wiki/` — the LLM-maintained knowledge base. Page types, frontmatter,
  and citation rules live in `RULES.md`.
- `AGENTS.md` / `CLAUDE.md` — dual-written pointer blocks telling any
  agent where the conventions live. **Identical content in both files.**
- `Taskfile.yml` — includes the wiki taskfile so `task wiki:lint` works
  from the repo root.

## Three-step adoption (from scratch)

1. **Copy** `.tickets/` and `.wiki/` into your repo.
2. **Add pointer** — paste the `## Conventions` block from `AGENTS.md`
   into both `AGENTS.md` and `CLAUDE.md` at your repo root.
3. **Customize** — edit `.tickets/harness/taxonomy.yaml` for your
   project's ticket categories, and `.wiki/RULES.md` for your knowledge
   taxonomy.

## Run the tests locally

```bash
task -d .tickets test     # expect 10/10
task -d .wiki lint        # expect OK
```

Both run in CI on every push and PR — see `.github/workflows/ci.yml` at
the top of the `agent-repo-kit` repo.
