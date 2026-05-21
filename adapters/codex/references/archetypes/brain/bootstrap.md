# Bootstrap: Brain Repo

Step-by-step procedure to scaffold a brain repo. Reads the [archetype description](./pattern.md).

## Steps

1. **Pre-flight.** Confirm `git`, `task`, `gitleaks` are installed. Inspect any directories the user wants to absorb — *especially* check for nested `.git` directories (other people's repos that should NOT be absorbed; see Anti-Patterns in `pattern.md`).

2. **Write `.conventions.yaml`** at the repo root using `templates/conventions.yaml.tmpl`. Realm names are the owner's choice; commit them deliberately.

3. **Scaffold the realm directories** with a README each. `library/` and `derived/` get one-line READMEs declaring the contract. Owner and captured realms get fuller READMEs.

4. **Write `CLAUDE.md`** from `templates/CLAUDE.md.tmpl` with the realms block, hard rules, layout overview, and pointers. Mirror to `AGENTS.md` if both are declared in `agent_docs:`.

5. **Write `Taskfile.yml`** with `task verify` as the canonical entry, `verify:secrets`, `verify:agent-docs-mirror`, and one subtask per declared realm/operation gate.

6. **Write `scripts/verify-*.sh`** — at minimum `verify-raw-immutability.sh` (template provided), `verify-source-readmes.sh`, and (if temporal) `verify-today-symlink.sh`. Make them soft-pass during bootstrap.

7. **Wire `.githooks/pre-commit`** with the cheap subset of verify gates (always: secrets + mirror + immutability + source-READMEs). Enable via `git config core.hooksPath .githooks`.

8. **Adopt operations as needed** — `.work/` for ingest/triage tracking, `.wiki/` for cross-cutting reference. Each adoption follows its own recipe doc and adds its own gate.

9. **Migrate or seed initial content.** For *legacy* personal-content directories (old journals, accumulated notes), prefer **preserve as-is in a `legacy-*/` subdirectory** over lossy schema conversion. The archetype's gates exempt `legacy-*/`. See [`migration.md`](./migration.md) for the full safe-migration procedure when absorbing live data.

10. **Smoke-test.** `task verify` must exit 0. Run `git status` and check for unintended additions; commit the scaffold; only then push to a private remote (with explicit user confirmation — pushing is a "visible to others" action).

## Common Gotchas During Bootstrap

- **Daily-log format mismatch on legacy migration.** Old journals are often monthly files with day headers; the new format is per-day files. Preserve old-format under `<owner-realm>/<daily>/legacy-monthly/` (or similar); start per-day going forward. The schema gate exempts `legacy-*/`.
- **Large binaries.** Brains accumulate PDFs, EPUBs, audio. Defer Git LFS until size pressure is real; document the LFS extension list in `.gitattributes` so flipping the switch later is a single command.
- **Empty `skill_roots:` are still verified.** If declared, every listed root must exist. Repos with no project-local skills should omit the key entirely rather than declare empty roots.
