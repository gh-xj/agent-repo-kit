# Build Cache: `preconditions`, `status`, `sources`

go-task gives you three orthogonal mechanisms for "should this task run?" Each
answers a different question and mixing them up is the most common source of
spurious rebuilds or skipped work.

## Decision Flow

When you sit down to write a task, decide in this order:

1. **`preconditions:`** — fail loudly if the environment is wrong. Tool
   missing, required file absent, wrong branch. Runs before the task body; if
   any precondition command exits non-zero, the task errors out (it does
   **not** skip).
2. **`status:`** — skip if the work is already done. You write the check. All
   status commands must exit 0 for go-task to consider the task up-to-date.
   There is no automatic source tracking here.
3. **`sources:` + `generates:` + `method:`** — automatic fingerprint of input
   files. go-task records a hash of `sources:` content into `.task/`, and
   re-runs if the hash changes.

When both `status:` and `sources:` are set, **both must agree to skip**. Either
one signaling "out of date" forces a re-run.

## `method:` Variants

- **`method: checksum`** (default): SHA-256 over `sources:` content. Safe in
  CI. Slightly slower than timestamp on huge trees, but correct regardless of
  how files were fetched. Use this for source code.
- **`method: timestamp`**: compares mtimes of `sources:` against `generates:`.
  Fragile — `git clone` and many CI caches reset mtimes, so the task always
  runs. **Never use in CI.** Fine for local-only throwaway tasks.
- **`method: none`**: disable fingerprinting entirely. The task always runs
  unless a `status:` check skips it. Combine with an explicit `status:` for
  full control.

## Pitfalls

- **`generates:` alone does nothing** under `method: checksum`. The fingerprint
  reads `sources:`. `generates:` only affects `method: timestamp` (comparing
  output mtimes against inputs).
- **`.task/` must be in `.gitignore`**. This is where checksums are recorded.
  Committing it poisons every contributor's cache. Lint rule
  `fingerprint-dir-gitignored` enforces this.
- **Recursive globs need `**`**. `sources: ['src/*.go']`does not walk
subdirectories. Use`sources: ['src/**/*.go']` for recursion.
- **Dynamic `sh:` vars do not participate in the fingerprint.** A var defined
  as `FOO: {sh: git rev-parse HEAD}` changes every commit, but go-task does
  not mix var values into the checksum. If a task's output depends on a var,
  encode that into `status:` explicitly.

## Per-Use-Case Recipes

### Source-code compile

Classic sources/generates/checksum. CI-safe, fast on cache hit.

```yaml
build:
  desc: Compile the binary
  sources: ["**/*.go", "go.mod", "go.sum"]
  generates: ["bin/myapp"]
  method: checksum
  cmds:
    - go build -o bin/myapp .
```

### Download-if-missing

No input files to fingerprint. Use `status:` to skip when the artifact exists.

```yaml
fetch-tool:
  desc: Download protoc if missing
  status:
    - test -f bin/protoc
  cmds:
    - curl -sSfL https://... -o bin/protoc
    - chmod +x bin/protoc
```

### Docker image up-to-date

Input is external (the Dockerfile + registry state). Use `status:` against the
local image store, and disable automatic fingerprinting.

```yaml
image:
  desc: Build app container
  method: none
  status:
    - docker image inspect myapp:dev >/dev/null 2>&1
  cmds:
    - docker build -t myapp:dev .
```

### CI-safe always-checksum

For any task that touches source code and might run in CI, default to
`method: checksum`. Timestamps on CI are unreliable; the checksum cost is
negligible.

```yaml
build:
  sources: ["**/*.go", "go.mod", "go.sum"]
  generates: ["bin/app"]
  method: checksum # explicit, even though it's the default
  cmds:
    - go build -o bin/app .
```

## Quick Reference

| Need                        | Mechanism                         |
| --------------------------- | --------------------------------- |
| Tool must exist             | `preconditions:`                  |
| Output file already present | `status:`                         |
| Input files unchanged       | `sources:` + `method: checksum`   |
| Mix: file + env state       | `status:` + `method: none`        |
| Local-only, speed critical  | `method: timestamp` (never in CI) |
