# adapters/

Adapters are thin wrappers that expose the harness-free skills under
`skills/` to a specific agent harness. An adapter owns only runtime-specific
loader glue; skill content stays under `skills/`.

**No adapter owns content.** All adapters point at the same sources under
`skills/`. If you find yourself duplicating skill text into an adapter, stop
and hoist it back into `skills/<skill-name>/`.

## Current adapters

- `claude-code/` — shipped. `install.sh --target claude-code` symlinks the
  skill surfaces under `skills/` into `~/.claude/skills/`.
- `codex/` — shipped. `install.sh --target codex` symlinks the skill surfaces
  under `skills/` into `~/.codex/skills/`.
- `cursor/` — placeholder docs only; no packaged adapter or install wiring.

## Manifest

`manifest.json` is the single source of truth for which skills get
symlinked into which harness. Consumed by `ark adapters link` and
`ark adapters list-links`. Add a new `harnesses[].links[]` entry there
when wiring a new skill into an existing harness.

Contributions welcome for new harnesses.
