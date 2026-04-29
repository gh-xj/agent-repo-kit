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

- linking the six shipped repo skills into `~/.agents/skills/`
- linking Claude compatibility entries from `~/.claude/skills/` to the shared
  `~/.agents/skills/` roots
- leaving `~/.codex/skills/` untouched, so runtime-owned entries such as
  `.system` remain separate

Restart Codex after linking so it rescans the skill root. Manual maintainer
symlinks may not appear in `npx skills ls`; verify the installed paths
directly if you need to confirm the layout.

## Legacy Low-Level Flow

If you are working from a local checkout and explicitly want the repo's
low-level linker instead of `npx skills`, build `ark` and run:

```sh
git clone https://github.com/gh-xj/agent-repo-kit.git
cd agent-repo-kit
(cd cli && go build -o bin/ark ./cmd/ark)
./bin/ark adapters link --target codex --repo-root "$PWD"
```

That path is mainly for legacy compatibility workflows. The adapter files in
this directory are generated compatibility mirrors; they are not the primary
skill source.

## Guardrails

- Keep adapter files thin. They point at the skill surfaces under `skills/`
  rather than duplicating content.
- Do not introduce Codex-specific skill authoring rules here — those belong
  in the portable `skills/skill-builder/` skill.
- When wiring a new skill into this adapter, keep `adapters/manifest.json`
  aligned so `ark adapters link --target codex` stays functional.
