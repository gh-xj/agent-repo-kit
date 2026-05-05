# `.conventions.yaml`

A repo-root YAML file that declares which conventions the repo opts into and
which checks an agent should run. Replaces the previous JSON contract.

A machine-readable JSON Schema lives at
`skills/convention-engineering/schemas/conventions.schema.json`. Add this
line at the top of the descriptor for editor autocomplete and inline
validation:

```yaml
# yaml-language-server: $schema=skills/convention-engineering/schemas/conventions.schema.json
```

## Principle

Minimal opt-ins + free-form checks. The schema is open: an agent reads the
file, applies the declared conventions, and runs the listed checks. The skill
does not police the shape beyond a small set of recognised keys.

## Location

`/.conventions.yaml` at the repo root. Single file. No overlay variants.

If the repo cannot commit it (e.g. external open source), append
`.conventions.yaml` to `.git/info/exclude` and keep it local — same file, same
shape.

## Recognised Keys (all optional)

```yaml
agent_docs:
  - CLAUDE.md
  - AGENTS.md # files that must exist as agent contracts

docs_root: docs # or .docs

taskfile: true # repo declares a canonical Taskfile

pre_commit: true # repo runs a pre-commit hook

skill_roots: # repo-local agent-skill discovery roots
  - .claude/skills
  - .agents/skills

operations: # adopted operational conventions
  - work # `.work/` tracker — see references/operations/work.md
  - wiki # `.wiki/` knowledge base — see references/operations/wiki.md

checks: # free-form list of agent-readable rules
  - "Every doc under {docs_root}/requests/ has a frontmatter `id`."
  - "AGENTS.md and CLAUDE.md have identical content if both declared."
  - "task verify exits 0 from a clean checkout."
```

Any key not in this list is allowed; the agent treats unknown keys as
repo-specific extension and respects them when present. Unknown keys must
not crash a reader.

## Reader Contract

An agent or `task verify` consuming this file:

1. Loads YAML; tolerates missing keys (treat as opt-out).
2. For each declared opt-in, verifies the corresponding artifact exists or
   the corresponding behavior runs:
   - `agent_docs` → each listed file exists.
   - `docs_root` → directory exists; `requests/`, `planning/`, `plans/`
     subfolders exist.
   - `taskfile` → root `Taskfile.yml` exists; `task verify` is defined.
   - `pre_commit` → `.githooks/pre-commit` exists and is executable, or
     `core.hooksPath` points to one.
   - `skill_roots` → each listed root exists if declared.
   - `operations` → for each adopted op, the corresponding directory exists
     (`.work/`, `.wiki/`), the agent contract files contain the op's pointer
     snippet, and the op-specific health check runs (e.g. `task work -- view
ready --json` succeeds when `work` is adopted).
3. For each entry under `checks:`, applies the rule. The rule format is
   prose; the agent interprets it. This is intentional — checks evolve per
   repo without schema lock-in.
4. Reports gaps. Does not auto-fix without user approval.

## Pattern Extensions

Some keys come from specific repo patterns rather than the core conventions.
Unknown keys remain allowed — this section just canonizes names so they don't
drift across pattern docs:

- `realms:` and `ingest_sources:` — see `references/patterns/brain.md` (the
  brain pattern formalizes a four-realm split: owner-authored, agent-captured
  immutable, external-authored, derived-regenerable).
- `epic:` — see `references/patterns/epic-wrapper.md`. Used by an epic
  wrapper repo to declare its sibling leaves. Shape:

  ```yaml
  epic:
    leaves:
      - <leaf-1> # name only; ../<leaf> and repo/<leaf> are implied
      - <leaf-2>
    composed: true # optional: assert compose.yaml lists each leaf
  ```

  `scripts/verify.sh` reads `epic.leaves` and asserts each `../<leaf>` is a
  directory and `repo/<leaf>` is a symlink that resolves to `../../<leaf>`.
  `scripts/bootstrap.sh` reads the same list to recreate the symlinks after
  clone.

A pattern doc is the right place to declare a new well-known key. Adding one
here without a pattern doc to back it is premature.

## Bootstrap

Create `.conventions.yaml` first; everything else flows from it. The file is
the single source of truth for "what conventions does this repo follow."

## Why YAML, Not JSON

Comments, multiline strings, and the free-form `checks:` list read better in
YAML. The reader is an agent, not a strict schema validator.
