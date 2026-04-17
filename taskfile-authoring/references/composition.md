# Composition with `includes:`

Deep dive on go-task's `includes:` mechanism. Covers structure, path
resolution, variable precedence, and the `flatten: true` vs explicit `dir:`
decision.

## Full `includes:` Entry

```yaml
includes:
  conventions:
    taskfile: ./.convention-engineering/Taskfile.yml
    dir: ./.convention-engineering
    optional: false
    internal: false
    flatten: false
    aliases: [conv]
    excludes: [legacy-task]
    vars:
      PROJECT_NAME: my-repo
```

Keys:

- `taskfile:` — path to the Taskfile to include. Relative to the
  **including file**, not CWD.
- `dir:` — working directory for tasks in the included file. Also relative
  to the including file. Defaults to the directory of `taskfile:`.
- `optional: true` — include is skipped silently if the file is missing. V1
  lint only checks existence for **non-optional** includes (rule
  `includes-paths-resolvable`).
- `internal: true` — tasks from this include are hidden from `task --list`
  and cannot be invoked from the CLI (only as deps).
- `flatten: true` — included tasks become top-level. See decision rule below.
- `aliases: [...]` — additional namespace keys that refer to the same include.
- `excludes: [...]` — list of task names to drop from the include.
- `vars:` — values passed into the included file (but see precedence caveat).

## Path Resolution

Paths in `taskfile:` and `dir:` are resolved relative to the **including**
Taskfile, not the CWD at invocation time. This matters when:

- The user runs `task -t other/dir/Taskfile.yml foo` from an unrelated CWD.
- A convention-pack overlay ships its own Taskfile that gets included by
  whatever root Taskfile exists.

Inside an included task, three template variables pin down "where am I":

| Variable                | Resolves To                                              |
| ----------------------- | -------------------------------------------------------- |
| `{{.TASKFILE_DIR}}`     | Directory of the Taskfile that defines the current task. |
| `{{.ROOT_DIR}}`         | Directory of the top-level (entry) Taskfile.             |
| `{{.USER_WORKING_DIR}}` | CWD when the user invoked `task`.                        |

Rule: inside an included file, **always** reference paths with
`{{.TASKFILE_DIR}}/...`. Hard-coded `../` is an anti-pattern — it breaks
the moment the overlay is moved.

## Variable Precedence (Surprising)

When a task in an included file reads `{{.FOO}}`, go-task resolves `FOO` in
this order, highest to lowest:

1. **Call-site vars** — `task: ns:foo` with `vars:` block, or
   `task ns:foo FOO=x` on the CLI.
2. **Task-level `vars:`** — defined on the task itself.
3. **Included file's own `vars:`** — top-level `vars:` block in the
   included Taskfile. **This wins over the includer's `includes.ns.vars:`.**
4. **Includer's `includes.ns.vars:`** — values passed in from the
   `includes:` entry.
5. **Root `vars:`** — top-level `vars:` of the entry Taskfile.
6. **Environment** — OS env vars.

The surprise is #3 > #4: if you ship an included Taskfile with
`vars: {FOO: default}`, the includer cannot override it by passing
`includes.ns.vars.FOO`. The fix:

```yaml
# inside the included Taskfile
vars:
  FOO: '{{.FOO | default "fallback"}}'
```

Now `FOO` prefers an outer value and falls back to `"fallback"` only if
nothing else is set.

## Root Escape From Inside An Include

`task :build:web` (leading colon) invokes a task at the root Taskfile from
inside an included file. Useful for shared build orchestration. Without the
leading colon, `build:web` tries to resolve within the current namespace.

## `flatten: true` vs Explicit `dir:`

Decision rule:

- **`flatten: true`** — transparent overlay. The user should not know an
  include happened. Example: a scaffolded `.convention-engineering/` ships
  `verify`, `audit`, `bootstrap`. Callers type `task verify`, not
  `task conventions:verify`.
- **Explicit `dir:`** — genuine namespace. The include models a real
  sub-domain with its own lifecycle. Example: `tickets:test`, `wiki:lint`,
  `docs:build`. The namespace adds information.

Flatten without a plan is a minefield — see anti-patterns below.

## Anti-Patterns

### Hard-coded `../` in included-file cmds

```yaml
# BAD — inside .convention-engineering/Taskfile.yml
tasks:
  verify:
    cmds:
      - go run ../scripts/audit.go
```

Moving the overlay up or down a directory breaks the command silently. Use:

```yaml
# GOOD
tasks:
  verify:
    cmds:
      - go run {{.TASKFILE_DIR}}/scripts/audit.go
```

### `flatten: true` causing collisions

Six months in, someone adds a `test` task to the included file. The repo
already has a root `test` task. With `flatten: true`, one silently shadows
the other.

Fix options:

- Use `excludes: [test]` on the include.
- Rename the included task.
- Drop `flatten: true` and route through the namespace.

V1 lint catches this as rule `flatten-no-name-collision`.

### `dotenv:` in an included file

go-task only honors `dotenv:` at the **root** Taskfile. An included file
with `dotenv: ['.env']` produces no error and loads nothing. Move it up to
the root Taskfile, or load env explicitly in the task via `set -a; . .env`.

### Variable triangulation across includes

If the includer passes `vars: {FOO: x}` and the included file defaults `FOO`
to something else, the default wins (see precedence). Always use the
`{{.FOO | default "..."}}` pattern in the included file when values are
meant to be overridable.
