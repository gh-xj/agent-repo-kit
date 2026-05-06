# Verification

Verification has two layers: repo-owned invariants and local Slock projections.

## Repo-Owned Gate

Run the repo's canonical gate, usually:

```bash
task verify
```

It should check:

- agent docs exist and mirror when required
- docs taxonomy exists
- epic symlinks point at declared leaves
- registry fields are valid and unique
- each leaf descriptor agrees with the registry
- leaf `.work/` is tracked when using `per_leaf_tracked`
- declared skill roots exist
- domain-specific checks pass

## Local Slock Projection Gate

Run only when local Slock dirs are expected:

```bash
task slock:check
```

It should check:

- local agent dir exists for each mapped agent
- generated `repo` symlink exists
- symlink target matches descriptor
- generated symlink is locally gitignored when the Slock dir is a git repo
- `MEMORY.md` exactly matches generated pointer-only content

Exact match matters. If memory may accumulate active rules, it will drift into a
second protocol store.

## Skill Discovery Gate

For every leaf with repo-local skills:

- descriptor declares every runtime skill root used by mapped agents
- each declared root exists
- mirrors are symlinks or generated copies, not hand-maintained drift
- `SKILL.md` exists through every discovery path

If the project only supports one runtime, say that explicitly in the repo docs.

## Migration Success Check

A migration is done when:

- mapped Slock memory is pointer-only
- Slock notes no longer contain active rules that are missing from the repo
- unmapped agents with stale product context are marked historical or pointed
  at the canonical epic
- the target leaf can work from its own `.work`, docs, and skills without
  replaying Slock history

## Self-Evolution Check

When a real run exposes a protocol failure:

1. patch the owning reference or workflow
2. add a `MAINTENANCE.md` entry with trigger, change, verification, and
   falsifier
3. include a regression case in that entry
4. run the relevant repo gate

No `eval/` folder is used for this skill; maintenance entries are the durable
regression surface.
