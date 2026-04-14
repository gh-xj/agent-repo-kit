# Bootstrap Workflow

Use this workflow to scaffold convention files for a new repo.

If the repo is an external/open-source project and you do not want to commit local convention files, use `references/operations/open-source-git-exclude-workflow.md` instead.

## Prerequisites

- Repo must already have source code (not scaffolding from scratch)
- You need to know which stack profiles apply

## Steps

### Step 1: Determine Profiles

Ask which stacks the repo uses:

- **Go**: Backend services, CLI tools, libraries
- **TypeScript/React**: Frontend SPAs, fullstack apps
- **Python**: Scripts, data pipelines, ML services, FaaS functions

Multiple profiles compose. A Go backend + React frontend uses both.

### Step 2: Generate Convention Contract

Start from `references/convention-config.example.json` and write the contract
artifact before running audit:

- tracked mode: `.convention-engineering.json`
- local-overlay mode: `.docs/convention-engineering.overlay.json`

Customize at least:

- `contract_version`
- `mode`
- `profiles`
- `docs_root`
- `ownership_policy`
- `mirror_policy`
- `evaluation_inputs`
- `chunk_plan`
- `required_files`
- `taskfile_checks`

This contract is the shared machine artifact for both bootstrap and audit.

### Step 3: Establish Docs Taxonomy

Read `references/core/docs-taxonomy.md`, then create taxonomy folders from the
contract you just wrote:

- tracked mode uses the contract's `docs_root: "docs"`
- local-overlay mode uses the contract's `docs_root: ".docs"`

Create taxonomy folders under the contract-selected root:

- `requests/`
- `planning/`
- `plans/`
- `implementation/`
- `taxonomy/`

### Step 4: Generate CLAUDE.md

Read `references/core/agent-knowledge.md` for the 5-section template. Customize with:

- Repo name and purpose
- Actual architecture layers (inspect the code)
- Actual commands (check for Makefile, Taskfile, package.json scripts)
- Stack-specific rules from the relevant profiles
- The tracked or overlay contract path you just created

### Step 5: Generate Linter Config

Read the relevant profile for linter config templates:

- Go: `.golangci.yml` from `profiles/go.md`
- TypeScript: `eslint.config.js` from `profiles/typescript-react.md`
- Python: `[tool.ruff]` in `pyproject.toml` from `profiles/python.md`

Customize layer boundary rules to match actual package structure.

### Step 6: Generate Taskfile.yml

Read the relevant profile for standard task names. Create:

- Root `Taskfile.yml` with includes
- `taskfiles/core.yml` with build, test, lint, fmt, verify

### Step 7: Generate Pre-Commit Hook

Read the relevant profile for hook template. Create `.githooks/pre-commit`.
Configure git to use it: `git config core.hooksPath .githooks`

### Step 8: Scaffold Operational Conventions

Decide whether to adopt the repo-scoped operational conventions. Each is
independent — adopt either, both, or neither.

- **Tickets (`.tickets/`)** — adopt if the repo needs tracked work items with
  audit trail, state transitions, and schema-validated categories. Skip for
  casual TODOs. If adopting, follow `references/operations/tickets.md` (3-step
  adoption).
- **Wiki (`.wiki/`)** — adopt if the repo needs an LLM-maintained knowledge
  base with source summaries, notes, and wikilink lint. Skip if a plain `docs/`
  tree is enough. If adopting, follow `references/operations/wiki.md` (3-step
  adoption).

Confirm: for each adopted op, the `CLAUDE.md` / `AGENTS.md` pointer snippet
from the ops doc is pasted into the repo's agent contract files.

### Step 9: Audit

Run the audit workflow (`references/operations/audit-workflow.md`) to verify the scaffold is complete.

### Step 10: Commit

```bash
git add CLAUDE.md .convention-engineering.json .golangci.yml Taskfile.yml taskfiles/ .githooks/
git commit -m "chore: bootstrap convention files"
```

If `.tickets/` or `.wiki/` were adopted, include them in the commit
(`git add .tickets/ .wiki/`).

For open-source/local-only setups, skip commit and keep `.docs/convention-engineering.overlay.json` plus other local-only files under `.git/info/exclude`.

## Post-Bootstrap

After scaffolding:

1. Run `task verify` to confirm all gates pass
2. Update CLAUDE.md with any repo-specific rules discovered during setup
3. Consider whether the repo needs a cross-repo skill (3+ consumers criteria)
