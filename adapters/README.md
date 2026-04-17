# adapters/

Adapters are thin wrappers that expose the harness-free content under
`convention-engineering/`, `convention-evaluator/`, and `skill-builder/` to a
specific agent harness. An adapter owns only runtime-specific loader glue;
the skill content stays in the repo-root surfaces.

**No adapter owns content.** All adapters point at the same source in
`convention-engineering/`, `convention-evaluator/`, and `skill-builder/`. If
you find yourself duplicating skill text into an adapter, stop and hoist it
back into the repo-root surfaces.

## Current adapters

- `claude-code/` — shipped. `install.sh --target claude-code` symlinks the
  three skill surfaces into `~/.claude/skills/`.
- `codex/` — placeholder docs only; no packaged adapter or install wiring.
- `cursor/` — placeholder docs only; no packaged adapter or install wiring.

Contributions welcome for new harnesses.
