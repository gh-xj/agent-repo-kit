# Bootstrap: Epic Wrapper

Step-by-step procedure to scaffold an epic-wrapper repo. Reads the [archetype description](./pattern.md).

## Steps

1. **Create the directory.** `<workspace>/<product>/<product>-epic` under the same parent as the leaf repos. `cd` into it.

2. **Write the file inventory** from `templates/`. Substitute `<product>` and `<leaf-N>` throughout:

   ```bash
   cp templates/conventions.yaml.tmpl       .conventions.yaml
   cp templates/gitignore.tmpl              .gitignore
   mkdir -p scripts
   cp templates/bootstrap.sh                scripts/bootstrap.sh
   chmod +x scripts/bootstrap.sh
   cp templates/Taskfile.yml.tmpl           Taskfile.yml
   cp templates/CLAUDE.md.tmpl              CLAUDE.md
   ```

   Then hand-edit (or `sed -i`) the substitutions for `<product>` and `<leaf-N>`.

3. **Create the docs taxonomy.** Five canonical subdirs are required by `scripts/verify.sh`:

   ```bash
   mkdir -p docs/{requests,planning,plans,implementation,taxonomy}
   for d in requests planning plans implementation taxonomy; do
     touch "docs/$d/README.md"
   done
   ```

4. **Initialize the work store.**

   ```bash
   work --store .work init
   ```

   If a `.work/` already exists from prior local state, `mv` it here instead of running `init`.

5. **Recreate `repo/<leaf>` symlinks for the first time.**

   ```bash
   bash scripts/bootstrap.sh
   ```

   At least one leaf should already exist as a sibling directory; if not, the script will warn but still succeed.

6. **Smoke the gate.**

   ```bash
   git init -b main && bash scripts/verify.sh
   ```

   Must print `verify: opt-ins ok`. The most common first-failure is a missing `docs/taxonomy/README.md` — it's required and easy to forget.

7. **Initial commit.**

   ```bash
   git add -A
   git commit -m "chore: bootstrap epic-wrapper for <product>"
   ```

## Common Gotchas During Bootstrap

- **`docs/taxonomy/` is required.** Create all five canonical subdirs even if you don't have content for them yet.
- **Untracking a previously-committed `.work/`** requires `git rm -r --cached .work/`. A `.gitignore` edit alone leaves the existing files tracked. See [`migration.md`](./migration.md).
- **`go-task` v3 mis-parses bare strings with colons inside `cmds:`.** An unquoted `echo "TODO: foo"` parses as a YAML mapping. Use `--` instead of `:` in echo strings, or single-quote the entire scalar.
- **`yq` must be the Go (mikefarah) version**, not the Python wrapper. Install via Homebrew or the GitHub releases page.
- **Workspace-file paths are relative to the workspace file**, not the symlink. A root `go.work`, `pnpm-workspace.yaml`, or equivalent should use either sibling paths such as `../<leaf>` or wrapper-local paths such as `./repo/<leaf>` deliberately; do not copy the symlink target string blindly.
