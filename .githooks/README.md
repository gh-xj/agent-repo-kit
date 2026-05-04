# `.githooks/`

Repo-local hook scripts. Not wired by default — you opt in.

## What's here

- `pre-commit` — fail the commit if adapter copies drift from canonical.
  Cheap (only triggers when staged files include `skills/**/SKILL.md` or
  `adapters/**/SKILL.md`). The check shells out to
  `scripts/sync-adapters.sh --check`.

## Wire it up

```bash
task setup:hooks
```

That runs `git config core.hooksPath .githooks` for this clone. Bypass a
single commit with `git commit --no-verify`.

## Caveats

- **Corporate / global hooks path.** If you have a global hook path
  (e.g. corp commit-policy hooks at `~/.../commit_hook/`), setting
  `core.hooksPath` here overrides it for this repo. If you need both,
  chain: write a `.githooks/pre-commit` that calls the corp hook first,
  then runs the adapter drift check.
- **CI redundancy.** `task verify` (used in CI) also runs the drift
  check. The hook is for fast local feedback; it does not replace CI.
- **Descriptor honesty.** `.conventions.yaml` declares `pre_commit:
false` because the hook is opt-in. Flip the descriptor to `true` only
  in repos that wire `core.hooksPath` as part of bootstrap.
