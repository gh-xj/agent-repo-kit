# adapters/cursor/

No packaged Cursor adapter is shipped yet — `install.sh` does not accept
`--target cursor`. The section below is the manual-adoption recipe so a
Cursor user can consume the repo-root skill surfaces without waiting for
a packaged adapter.

## Manual adoption

Cursor reads workspace-level docs out of `.cursor/rules/` (MDC files) or
out of per-file markdown at known paths. The kit ships generated,
frontmatter-free mirrors under this directory — `skill-builder.md`,
`convention-engineering.md`, and `convention-evaluator.md` — which are
synced from the repo-root canonical sources by `ark skill sync`.

```sh
# 1. Clone the kit and build ark.
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
(cd cli && go build -o bin/ark ./cmd/ark)

# 2. Regenerate the Cursor mirrors from canonical (normally already in sync).
./cli/bin/ark skill sync --repo-root .

# 3. Copy or symlink the three mirrors into your own repo's Cursor config.
#    Example for a target repo at $TARGET:
mkdir -p "$TARGET/.cursor/rules"
cp adapters/cursor/*.md "$TARGET/.cursor/rules/"
```

The `.agent-repo-kit.json` manifest already declares Cursor targets —
re-running `ark skill sync` in the kit will always regenerate these
files. `ark skill check` acts as a drift lint.

## Guardrails for a future packaged adapter

- Keep adapter files thin. They point at the skill surfaces under
  `skills/` rather than duplicating content.
- Cursor uses MDC for rule authoring; if a native MDC wrapper is added,
  keep metadata (globs, alwaysApply) out of the portable core.
- When you wire `--target cursor` into `install.sh`, emit a copy step
  (not a symlink) since Cursor rule files live inside the target repo,
  not a user-home path.
