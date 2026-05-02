# adapters/

Adapters are thin wrappers that expose the harness-free skills under
`skills/` to a specific agent harness. An adapter owns only runtime-specific
loader glue; skill content stays under `skills/`.

End-user installs should prefer the open skills CLI:

```bash
npx skills add gh-xj/agent-repo-kit -g -a claude-code -a codex --skill '*' -y
```

Maintainers working from a local checkout should prefer the repo-local symlink
workflow instead:

```bash
task skills:link-dev
```

That keeps `skills/` as the live source of truth while you edit. A local-path
`npx skills add /path/to/repo ...` install copies files and will not
live-update when the checkout changes.

**No adapter owns content.** All adapters point at the same sources under
`skills/`. If you find yourself duplicating skill text into an adapter, stop
and hoist it back into `skills/<skill-name>/`.

## Current adapters

- `claude-code/` — shipped. End-user install path:
  `npx skills add gh-xj/agent-repo-kit -g -a claude-code --skill '*' -y`.
- `codex/` — shipped. End-user install path:
  `npx skills add gh-xj/agent-repo-kit -g -a codex --skill '*' -y`.
- `cursor/` — placeholder docs only; no packaged adapter or install wiring.

## Manifest

`manifest.json` declares the harnesses and their skill roots. The set of
skills to link is auto-derived: every immediate subdirectory of `skills/`
that contains a `SKILL.md` is exposed to each harness's skill root.

Adding a new skill requires zero manifest edits — drop it into
`skills/<name>/` with a `SKILL.md` and `npx skills` discovery picks it up.
Adding a new harness means adding an entry to `harnesses[]`.

Contributions welcome for new harnesses.
