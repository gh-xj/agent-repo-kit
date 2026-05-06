# Migration: Epic Wrapper

Use this when moving an existing layout into the epic-wrapper archetype.

## Migrating from per-leaf `-dev` wrappers

If you have `<leaf-1>-dev/`, `<leaf-2>-dev/` etc. from the older `dev-wrapper` archetype (now subsumed by epic-wrapper):

1. Create `<product>-epic/` per [`bootstrap.md`](./bootstrap.md).
2. Move each `<leaf-N>-dev/.work/items/*.yaml` and matching `.work/spaces/<W-ID>/` into `<product>-epic/.work/`. Renumber IDs if they collide. Prefix titles with `[<leaf-N>]` if not already.
3. Move design docs from each `<leaf-N>-dev/docs/` into `<product>-epic/docs/`, prefixing filenames with `<leaf-N>_` to preserve provenance.
4. `git rm -rf <leaf-N>-dev` after committing the migration to the epic. Or archive each `-dev` repo to `_archive/` if you need the history.

## Migrating committed `.work/` to gitignored

If `.work/` is currently tracked in git (common when someone scaffolded without realizing), the gitignore alone doesn't untrack it:

```bash
git rm -r --cached .work/
echo ".work/" >> .gitignore
git commit -m "chore: untrack .work/ (now local intake only)"
```
