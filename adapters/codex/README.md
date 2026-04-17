# adapters/codex/

No packaged Codex adapter is shipped yet — `install.sh` does not accept
`--target codex`. The section below is the manual-adoption recipe so a
Codex user can consume the repo-root skill surfaces without waiting for
a packaged adapter.

## Manual adoption

Codex discovers skills under `~/.agents/skills/<skill>/SKILL.md`. The
pattern is identical to the Claude Code adapter under
`../claude-code/`, minus the YAML frontmatter block (Codex reads the
skill name from the directory and the description from the body).

```sh
# 1. Clone the kit and build ark (skip build if you already have it).
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
(cd cli && go build -o bin/ark .)

# 2. Symlink each repo-root skill surface into ~/.agents/skills/.
mkdir -p ~/.agents/skills
for name in convention-engineering convention-evaluator skill-builder; do
  ln -sf "$PWD/$name" "$HOME/.agents/skills/$name"
done

# 3. Restart your Codex session to pick them up.
```

The manifest at `../../.agent-repo-kit.json` already declares Codex
targets under `adapters/codex/` — running `ark skill sync` regenerates
those files. If you want frontmatter-free copies that do not rely on
the symlinked directory layout, render them explicitly and copy the
`adapters/codex/<skill>/SKILL.md` files into place.

## Guardrails for a future packaged adapter

- Keep adapter files thin. They point at the repo-root
  `convention-engineering/`, `convention-evaluator/`, and
  `skill-builder/` surfaces rather than duplicating content.
- Do not introduce Codex-specific skill authoring rules here — those
  belong in the portable `skill-builder/` skill.
- When you wire a new target into `install.sh`, mirror the three
  symlinks that the `claude-code` target creates.
