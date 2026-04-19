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

`manifest.json` declares the harnesses and their skill roots. The set of
skills to link is auto-derived: every immediate subdirectory of `skills/`
that contains a `SKILL.md` gets a symlink into each harness's skill root.

Adding a new skill requires zero manifest edits — drop it into
`skills/<name>/` with a `SKILL.md` and `ark adapters link` (and the
installer) pick it up. Adding a new harness means adding an entry to
`harnesses[]`.

Contributions welcome for new harnesses.
