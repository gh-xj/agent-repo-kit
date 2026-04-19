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
- `dir:` — working directory for tasks in the included file. Relative to
  the including file. **If omitted, included tasks run in the _including_
  Taskfile's directory, not the included file's directory** (see guide.md
  §"Directory of included Taskfile" and `internal/fsext/fs.go` `ResolveDir`
  — an empty `dir:` joins to the entrypoint's own dir). Set `dir:` explicitly
  when the overlay needs its own working directory.
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

Precedence follows from how go-task builds the var map in
`compiler.go:107-143` (`Compiler.getVariables`): it calls `Set` in a fixed
sequence, and each `Set` overwrites any prior value for the same key
(`taskfile/ast/vars.go:62`, orderedmap `Set`). **Later writes win.**

For a task being invoked (included or not), the write order is:

1. **Environment** — OS env vars seeded into the result first
   (`compiler.go:48`, plus `TaskfileEnv` at `:107`).
2. **Root Taskfile `vars:`** — `TaskfileVars` of the entry Taskfile
   (`compiler.go:112`). Note: this is always the _entry_ Taskfile's
   top-level `vars:`, not the included file's.
3. **Includer's `includes.ns.vars:`** — `t.IncludeVars`, populated when
   the task was merged from an include (`compiler.go:118`,
   `taskfile/ast/tasks.go:178`).
4. **Included file's own `vars:`** — `t.IncludedTaskfileVars`, the
   top-level `vars:` block of the included Taskfile (`compiler.go:123`,
   `taskfile/ast/tasks.go:179`).
5. **Call-site vars** — `call.Vars` from `task: ns:foo` with a `vars:`
   block, or `task ns:foo FOO=x` on the CLI (`compiler.go:134`).
6. **Task-level `vars:`** — `t.Vars` defined on the task itself
   (`compiler.go:139`). **Highest precedence.**

Two consequences that often trip people up:

- **Task-level `vars:` beat call-site vars.** Hard-coding `FOO: baked-in`
  at the task level means `task foo FOO=whatever` at the CLI cannot
  override it. Use `FOO: '{{.FOO | default "baked-in"}}'` if you want an
  overridable default.
- **Included file's own `vars:` beat the includer's `includes.ns.vars:`.**
  If you ship an included Taskfile with `vars: {FOO: default}`, the
  includer cannot override it by passing `includes.ns.vars.FOO`. Same
  fix:

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

Two different keys, two different behaviours:

- **Top-level `dotenv:` in an included Taskfile** — hard error. Merging
  returns `ErrIncludedTaskfilesCantHaveDotenvs` (`taskfile/ast/taskfile.go:44`:
  _"Included Taskfiles can't have dotenv declarations. Please, move the dotenv
  declaration to the main Taskfile"_). Move it up to the entry Taskfile.
- **Task-level `dotenv:` on a task defined in an included Taskfile** —
  works. The per-task dotenv loader runs against `task.Dir` regardless of
  where the task was defined (`variables.go:158-174`; covered by
  `TestTaskDotenv*` at `task_test.go:1913-1978` and the `included-task-dotenv`
  case at `executor_test.go:1024-1026`, comment: _"Somehow dotenv is working
  here!"_). The path is resolved relative to the task's working directory
  (`new.Dir`), which for included tasks is `include.Dir` joined with the
  task's own `dir:` (`taskfile/ast/tasks.go:174`).

Caveat: dotenv values populate the task's `Env` (shell env for the command),
not the template `vars` map. `{{.FOO}}` inside a `cmd:` will not see a
dotenv-sourced `FOO`; `$FOO` inside the shell command will.

### Variable triangulation across includes

If the includer passes `vars: {FOO: x}` and the included file defaults `FOO`
to something else, the default wins (see precedence). Always use the
`{{.FOO | default "..."}}` pattern in the included file when values are
meant to be overridable.
