# Convention Bootstrap Workflow

Scaffold convention files in a repo by writing `.conventions.yaml` first and
then creating each artifact it declares.

## Prerequisites

- Repo already has source code.
- You know which conventions apply (which can also be discovered by inspecting
  the code, then confirmed with the user).

## Steps

### 1. Draft `.conventions.yaml`

Create `/.conventions.yaml` at the repo root. Start with the minimum the repo
should adopt. Example:

```yaml
agent_docs:
  - CLAUDE.md
  - AGENTS.md
docs_root: docs
taskfile: true
pre_commit: true
skill_roots:
  - .claude/skills
checks:
  - "AGENTS.md mirrors CLAUDE.md (identical content)."
  - "task verify exits 0 from a clean checkout."
```

If the repo cannot commit the file (external open source, fork, etc.), keep
it locally and add `.conventions.yaml` to `.git/info/exclude`. Same file,
same shape, same readers.

### 2. Create Agent Contract Files

For each entry in `agent_docs:`, scaffold the file using the template in
`references/core/agent-knowledge.md`. Customise:

- repo name and purpose
- short architecture pointer
- exact build/test/lint commands (inspect Makefile / Taskfile / package.json)
- non-negotiable rules
- pointers to docs root and operational ops

If two contract files are declared, mirror their content.

### 3. Create Docs Taxonomy

Under the directory named in `docs_root:`, create:

- `requests/`
- `planning/`
- `plans/`
- `implementation/`

Add per-folder README files describing the filename contracts (see
`references/core/docs-taxonomy.md`).

### 4. Create Taskfile

If `taskfile: true`, create:

- root `Taskfile.yml` with `verify`, `fmt`, `lint`, `test` targets (or
  whatever the repo needs).
- optional `taskfiles/<area>.yml` sub-Taskfiles included from the root.

The skill is stack-agnostic. Pick the toolchain commands the repo already
uses; the convention is "one canonical entry point," not a specific linter.

### 5. Create Pre-Commit Hook

If `pre_commit: true`, create `.githooks/pre-commit` with the repo's chosen
fast checks (format, dep-tidy, optional lint --fix). Wire it up:

```bash
git config core.hooksPath .githooks
```

### 6. Adopt Operational Conventions (Optional)

Each is independent.

- **Work (`.work/`)** — see `references/operations/work.md`.
- **Wiki (`.wiki/`)** — see `references/operations/wiki.md`.

After adoption, paste each ops doc's pointer snippet into the agent contract
files.

### 7. Audit

Run the audit workflow (`references/operations/audit-workflow.md`) to verify
the scaffold matches the descriptor.

### 8. Commit

```bash
git add .conventions.yaml CLAUDE.md AGENTS.md docs/ Taskfile.yml .githooks/
git commit -m "chore: bootstrap conventions"
```

If `.work/` or `.wiki/` were adopted, include them. If `.conventions.yaml`
is local-only (open-source overlay), skip the commit and confirm the file
is in `.git/info/exclude`.

## Post-Bootstrap

1. Run `task verify` to confirm gates pass.
2. Update `.conventions.yaml` and the contract files with anything you
   discovered during bootstrap.
3. If a repo grows beyond what `.conventions.yaml` covers, extend the
   descriptor — unknown keys are allowed and treated as repo-local
   extensions.
