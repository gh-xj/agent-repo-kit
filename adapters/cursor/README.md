# adapters/cursor/

No packaged Cursor adapter is shipped yet. The section below is the manual
adoption recipe so a Cursor user can consume the repo-root skill surfaces
without waiting for a packaged adapter.

## Manual adoption

Cursor reads workspace-level docs out of `.cursor/rules/` (MDC files) or
out of per-file markdown at known paths. The kit ships hand-mirrored,
frontmatter-free copies under this directory — `skill-builder.md`,
`convention-engineering.md`, `convention-evaluator.md`, etc. — kept in sync
with the canonical `skills/<name>/SKILL.md` sources by hand.

```sh
# 1. Clone the kit.
git clone https://github.com/gh-xj/agent-repo-kit.git

# 2. Copy the mirrors you want into your own repo's Cursor config.
#    Example for a target repo at $TARGET:
mkdir -p "$TARGET/.cursor/rules"
cp agent-repo-kit/adapters/cursor/*.md "$TARGET/.cursor/rules/"
```

## Guardrails for a future packaged adapter

- Keep adapter files thin. They point at the skill surfaces under
  `skills/` rather than duplicating content.
- Cursor uses MDC for rule authoring; if a native MDC wrapper is added,
  keep metadata (globs, alwaysApply) out of the portable core.
