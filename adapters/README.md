# adapters/

Adapters are thin wrappers that expose the harness-free content under
`contract/` and `evaluator/` to a specific agent harness. Each adapter
either symlinks or re-exports those pieces in the format the harness
expects (for example, Claude Code expects a `SKILL.md` with a YAML
frontmatter header; Codex expects a TOML agent definition).

**No adapter owns content.** All adapters point at the same source in
`contract/` and `evaluator/`. If you find yourself duplicating convention
text into an adapter, stop and hoist it back into `contract/`.

## Current adapters

- `claude-code/` — a Claude Code skill shim that makes `contract/` load as
  `convention-engineering` and `evaluator/` as `convention-evaluator`.
- `codex/` — placeholder.
- `cursor/` — placeholder.

Contributions welcome for new harnesses.
