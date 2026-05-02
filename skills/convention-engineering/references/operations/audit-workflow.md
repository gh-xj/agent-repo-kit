# Convention Audit Workflow

Check an existing repo against its declared conventions. Produce a gap
report — no auto-fixes without user approval.

## The Loop

1. Read `.conventions.yaml` at the repo root.
2. For each declared opt-in, verify the corresponding artifact / behavior.
3. For each entry under `checks:`, apply the rule.
4. Produce a gap report.

If `.conventions.yaml` does not exist, the repo has no declared conventions.
Either propose creating one (see bootstrap) or stop.

## What to Verify Per Opt-In

| Key in `.conventions.yaml` | Verify                                                                                      |
| -------------------------- | ------------------------------------------------------------------------------------------- |
| `agent_docs: [files]`      | Each listed file exists and contains the required sections (see `core/agent-knowledge.md`). |
| `docs_root: <path>`        | Directory exists; `requests/`, `planning/`, `plans/` exist under it.                        |
| `taskfile: true`           | Root `Taskfile.yml` exists; `task verify` is defined.                                       |
| `pre_commit: true`         | A pre-commit hook is installed and runs.                                                    |
| `skill_roots: [paths]`     | Each declared root exists.                                                                  |
| `checks: [...]`            | Apply each rule by reading code/docs/tooling.                                               |

## Conditional Operational Conventions

Audit each only if its directory exists. See
`references/operations/work.md` and `references/operations/wiki.md`.

If `.work/` exists:

- agent contract files contain the work pointer snippet
- `.work/config.yaml` exists
- `task work -- view ready --json` (or `work --store .work view ready --json`) succeeds

If `.wiki/` exists:

- agent contract files contain the wiki pointer snippet
- `.wiki/scripts/lint.sh` exists and `task -d .wiki lint` passes
- `.wiki/raw/` and `.wiki/pages/` exist

## Output Format

```
## Convention Audit: <repo>

### Present
- [x] CLAUDE.md with required sections
- [x] task verify defined and exits 0
...

### Missing
- [ ] AGENTS.md declared in .conventions.yaml but file absent
...

### Partial
- [~] task verify runs lint + tests but no format check
...

### Recommendation
Top fixes ordered by impact.
```

## Applying Fixes

After presenting the report, ask the user which gaps to address. Then:

1. For each gap, propose the minimal change that satisfies the declared
   convention.
2. If a fix would change `.conventions.yaml` itself (adding or removing an
   opt-in), surface that as a separate decision — the descriptor is
   user-owned.
3. Run `task verify` after each batch of fixes.
