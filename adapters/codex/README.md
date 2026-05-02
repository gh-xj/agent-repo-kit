# adapters/codex/

Prefer the open skills CLI for Codex installs:

```sh
npx skills add gh-xj/agent-repo-kit -g -a codex --skill '*' -y
```

That installs the canonical skill directories under [`skills/`](../../skills)
using the runtime's supported layout. Restart Codex after install so it picks
up the new skills.

## Maintainer Development Flow

If you are actively editing this repo's `skills/` sources from a local clone,
prefer the maintainer symlink workflow:

```sh
task skills:link-dev
```

That keeps the repo checkout as the live source of truth by:

- linking every `skills/*/SKILL.md` repo skill into `~/.agents/skills/`
- linking Claude compatibility entries from `~/.claude/skills/` to the shared
  `~/.agents/skills/` roots
- leaving `~/.codex/skills/` untouched, so runtime-owned entries such as
  `.system` remain separate

Restart Codex after linking so it rescans the skill root. Manual maintainer
symlinks may not appear in `npx skills ls`; verify the installed paths
directly if you need to confirm the layout.

## Guardrails

- Keep adapter files thin. They point at the skill surfaces under `skills/`
  rather than duplicating content.
- Do not introduce Codex-specific skill authoring rules here — those belong
  in the portable `skills/skill-builder/` skill.
- When wiring a new skill into this adapter, keep `adapters/manifest.json`
  aligned with the canonical `skills/` set.
