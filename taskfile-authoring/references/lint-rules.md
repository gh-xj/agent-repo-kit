# `ark taskfile lint` V1 Rule Catalog

Ten structural rules. Each rule ID below matches the identifier used by the
sibling `cli/internal/tasklint` package, so `ark taskfile lint` output and
this doc stay in sync.

Every rule: one-line statement, rationale, a minimal trigger, the fix, and
an upstream doc link where relevant.

---

## 1. `version-required`

**Statement:** the top-level `version:` key must be present.

**Rationale:** go-task changes schema between major versions. Without an
explicit declaration, tools (including `task` itself) cannot reason about
which schema applies. A missing `version:` is always a bug.

**Trigger:**

```yaml
tasks:
  build:
    cmds:
      - go build .
```

**Fix:** add `version: '3'` at the top of the file.

**Upstream:** https://taskfile.dev/reference/schema/

---

## 2. `version-is-three`

**Statement:** `version:` must be `'3'` (string) or `3` (number).

**Rationale:** V1 lint only understands go-task v3 schema. Older schemas (v1,
v2) have divergent semantics (e.g. different `deps:` handling). Flagging
non-3 versions prevents confusing false positives from other rules.

**Trigger:**

```yaml
version: "2"
tasks:
  build: { cmds: [go build .] }
```

**Fix:** upgrade to go-task v3 and change to `version: '3'`.

**Upstream:** https://taskfile.dev/reference/schema/

---

## 3. `unknown-top-level-keys`

**Statement:** top-level keys must come from the v3 schema allowlist
(`version`, `vars`, `env`, `dotenv`, `includes`, `tasks`, `output`, `method`,
`silent`, `run`, `interval`, `set`, `shopt`).

**Rationale:** typos like `task:` (singular) or `varz:` silently produce a
Taskfile that does not behave as written. Catching them structurally is
cheap.

**Trigger:**

```yaml
version: "3"
task: # typo: should be `tasks:`
  build: { cmds: [go build .] }
```

**Fix:** correct the key (`task:` → `tasks:`), or remove it if unintended.

**Upstream:** https://taskfile.dev/reference/schema/

---

## 4. `unknown-task-keys`

**Statement:** per-task keys must come from the v3 task schema allowlist
(`cmds`, `cmd`, `desc`, `summary`, `aliases`, `sources`, `generates`,
`status`, `preconditions`, `method`, `prefix`, `silent`, `interactive`,
`internal`, `ignore_error`, `run`, `platforms`, `deps`, `label`, `vars`,
`env`, `dotenv`, `dir`, `set`, `shopt`, `requires`).

**Rationale:** the most common variant is `cmd:` mistyped as `command:` or
`deps:` mistyped as `depends:`, which silently skips the value.

**Trigger:**

```yaml
version: "3"
tasks:
  build:
    command: go build . # should be `cmd:` or `cmds:`
```

**Fix:** correct the key spelling.

**Upstream:** https://taskfile.dev/reference/schema/#task

---

## 5. `cmd-and-cmds-mutex`

**Statement:** a task may use `cmd:` or `cmds:`, never both.

**Rationale:** go-task accepts both keys, but which one actually executes
depends on load order. Mixing them is ambiguous and almost certainly a
refactor left half-finished.

**Trigger:**

```yaml
version: "3"
tasks:
  build:
    cmd: go build .
    cmds:
      - go test ./...
```

**Fix:** pick one. For a single command prefer `cmd:`; for a list prefer
`cmds:`.

**Upstream:** https://taskfile.dev/reference/schema/#task

---

## 6. `includes-paths-resolvable`

**Statement:** every non-`optional` include's `taskfile:` path must resolve
to an existing file relative to the including Taskfile.

**Rationale:** broken includes cause confusing "task not found" errors at
invocation time. Catching them during lint surfaces the real issue (a moved
or deleted file) immediately.

**Trigger:**

```yaml
version: "3"
includes:
  conventions:
    taskfile: ./.convention-engineering/Taskfile.yml # file does not exist
```

**Fix:** create the file, correct the path, or mark the include
`optional: true` if the miss is intentional.

**Upstream:** https://taskfile.dev/usage/#including-other-taskfiles

---

## 7. `flatten-no-name-collision`

**Statement:** when `flatten: true` is set on an include, none of the
included task names may collide with root task names or with other flattened
names.

**Rationale:** flatten silently shadows — `task test` might run the root's
`test` or the include's, depending on merge order. Contributors cannot tell
without reading all involved files.

**Trigger:**

```yaml
version: "3"
includes:
  conventions:
    taskfile: ./.conv/Taskfile.yml
    flatten: true
tasks:
  test: # also defined in .conv/Taskfile.yml
    cmds: [go test ./...]
```

**Fix:** add `excludes: [test]` to the include, rename one of the tasks, or
drop `flatten: true`.

**Upstream:** https://taskfile.dev/usage/#including-other-taskfiles

---

## 8. `method-valid-enum`

**Statement:** `method:` values are limited to `checksum`, `timestamp`, or
`none`.

**Rationale:** typos like `method: check` (intending `checksum`) are silently
invalid and trigger default behavior. The error surfaces as "task always
runs" which is hard to attribute.

**Trigger:**

```yaml
version: "3"
tasks:
  build:
    method: check
    sources: ["**/*.go"]
    cmds: [go build .]
```

**Fix:** change to `checksum`, `timestamp`, or `none`.

**Upstream:** https://taskfile.dev/usage/#prevent-unnecessary-work

---

## 9. `fingerprint-dir-gitignored`

**Statement:** `.task/` must be present in `.gitignore` whenever the
Taskfile declares any task with `sources:`, `generates:`, or `method:`.

**Rationale:** `.task/` stores the fingerprint database. Committing it
poisons other contributors' caches: go-task will see a checksum from
someone else's tree and skip tasks that should run.

**Trigger:** a Taskfile with `sources:` declared, and a `.gitignore` that
does not contain `.task/`.

**Fix:** add `.task/` to `.gitignore`.

**Upstream:** https://taskfile.dev/usage/#prevent-unnecessary-work

---

## 10. `dotenv-files-gitignored`

**Statement:** every path in a root-level `dotenv:` declaration must be
git-ignored (or untracked).

**Rationale:** `dotenv:` is typically used for secrets or machine-specific
config. Committing the referenced file leaks credentials or pins local
state into the repo.

**Trigger:**

```yaml
version: "3"
dotenv: [".env"]
# and .env is checked in (not in .gitignore)
```

**Fix:** add the file to `.gitignore`, remove it from the index
(`git rm --cached .env`), and commit the change.

**Upstream:** https://taskfile.dev/reference/schema/#dotenv

---

## DEFERRED to v2

Rules considered and cut from V1, with one-sentence "why not now" notes.
When usage data shows one of these is worth the false-positive budget,
promote it.

- **`task-has-desc`** — every task should have a `desc:`. Breaks for
  internal helpers, and `internal: true` is an acceptable workaround only
  in some cases. Needs a real-world sweep to calibrate.
- **`sources-generates-pairing`** — require `generates:` whenever `sources:`
  is declared. Not universal: many legitimate patterns use `status:`-only
  or `method: none` with sources.
- **`canonical-ci-verify-aggregator`** — require a `ci` task with
  `deps: [lint, test, build]`. Too opinionated for a general lint; a single
  repo's shape rather than a schema rule.
- **`go-canonical-tasks`** — require `fmt`, `fmt:check`, `lint`, `test`,
  `build` in any Taskfile that references `go`. Stack detection is fragile;
  advisory guidance lives in `stack-go-cli.md`.
- **`uv-canonical-tasks`** — analogous to the Go rule. Same fragility.
- **`no-bare-rm-rf`** — flag `rm -rf` in `cmds:`. Hits legitimate clean
  tasks too often; better enforced via code review.
- **`destructive-task-has-prompt`** — require an interactive confirmation
  before destructive operations. Stylistic, not structural.
- **Remote-taskfile validation** — fetch and validate remote `includes:`.
  Upstream remote includes are experimental and not deterministically
  resolvable (version pinning is still being specified).
- **`--strict` layer** — a tighter set that treats today's advisory rules
  as errors. Premature to ship without real opt-in/opt-out usage data.
- **Per-repo config (`.ark-taskfile.json`)** — rule suppression and
  per-repo overrides. Premature; add when we see genuine suppression
  needs in practice rather than guessing at the schema up front.
