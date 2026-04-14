# AGENTS.md — agent-repo-kit

You are inside **agent-repo-kit**. This repo publishes a convention for
_other_ repos to adopt. It is not itself a downstream project.

## Entry points

- `contract/` — the harness-free content describing repo conventions
  (tickets, wiki, agent docs, verification gates, etc.). Canonical source.
- `evaluator/` — the harness-free scoring rubric used to grade a repo's
  adoption against the contract.
- `examples/demo-repo/` — a working repo that shows the conventions applied
  end to end; the CI exercises it.
- `adapters/<harness>/` — thin shims that expose `contract/` and
  `evaluator/` to a specific harness (Claude Code, Codex, Cursor).

## Rules for editing this repo

1. **Do not** add harness-specific frontmatter (e.g. Claude skill YAML) to
   files under `contract/` or `evaluator/`. That belongs in
   `adapters/claude-code/SKILL.md` and equivalents.
2. **Do not** reference "Claude", "Skill tool", "Codex", or absolute user
   paths like `/Users/...` or `~/.claude/` inside `contract/` or
   `evaluator/`. Those are harness specifics.
3. **Dual-write pointer blocks** — when adding a new convention, update
   both `examples/demo-repo/AGENTS.md` and `examples/demo-repo/CLAUDE.md`
   identically.
4. **Adapters re-export, they don't own.** An adapter file should be a
   short wrapper pointing at `contract/` or `evaluator/`.

## Testing

```bash
task -d examples/demo-repo/.tickets test   # expect 10/10
task -d examples/demo-repo/.wiki lint      # expect OK
```

CI runs both on every push and PR (see `.github/workflows/ci.yml`).
