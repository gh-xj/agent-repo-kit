# Pattern: Epic Wrapper

A single sibling repo that owns the development process state for a
multi-repo product. The product's leaf repos (one per shipped artifact —
proto, daemon, server, cli, …) stay clean: code, install, release. The
epic wrapper holds the `.work/` tracker, design docs, deployment
artifacts, cross-repo verification glue, and a `repo/` symlink
namespace for each leaf.

Codified from the `loom-epic` bootstrap (2026-05-03). Replaces the older
`dev-wrapper` pattern; that pattern's single-leaf case is now the N=1
degenerate of this one.

## When To Use

Apply when:

- You ship a product as N≥1 sibling repos (proto + daemon + server, or
  api + worker + admin-ui, etc.) — anything where a coordinated change
  spans more than one repo.
- The leaves should stay minimal for adopters / future contributors.
- You need exactly one place to answer "do these versions work together?"
- You want to dogfood `.work/` for evolving the product itself.

The N=1 case (one leaf) also applies — if you maintain a single OSS
tool and want a maintainer-private wrapper, use this pattern with
`leaves: [<tool>]`. (This was previously the standalone `dev-wrapper`
pattern; merged here for consistency.)

Skip when:

- The product fits in a single repo (no multi-repo coordination needed).
- The repo is small enough that `docs/` + GitHub Issues is sufficient.
- You'd be the only maintainer and you don't already use `.work/` daily.

## Three Load-Bearing Decisions

These are the decisions that actually matter. Don't deviate without a
reason recorded in `CLAUDE.md`.

### 1. Sibling layout, `repo/` symlinks gitignored, bootstrapped from descriptor

Leaves and the wrapper are sibling directories on disk:

```
<workspace>/<product>/
├── <leaf-1>/             ← clean leaf (no .work/, no docs/)
├── <leaf-2>/
├── <leaf-3>/
└── <product>-epic/       ← THIS pattern
    └── repo/
        ├── <leaf-1>      → ../../<leaf-1>   (symlink, gitignored)
        ├── <leaf-2>      → ../../<leaf-2>   (symlink, gitignored)
        └── <leaf-3>      → ../../<leaf-3>   (symlink, gitignored)
```

The wrapper-local path is `repo/<leaf>` (root-relative shorthand:
`/repo/<leaf>`). The symlinks are **gitignored**, not committed. They
are recreated by `scripts/bootstrap.sh` after clone, which reads the
leaf list from `.conventions.yaml`. Reasons:

- Committed symlinks dangle if the cloner doesn't place sibling repos
  in the right relative location. Silent corruption is worse than an
  explicit bootstrap step.
- `repo/` makes symlinked source checkouts an explicit namespace
  instead of letting leaf names compete with wrapper-owned files.
- The leaf list belongs in `.conventions.yaml` anyway (the schema
  formalizes it). Once it's there, the bootstrap script is a trivial
  consumer.
- Keeps git history free of layout-mechanics churn (adding/removing a
  leaf is a one-line descriptor change, not a symlink commit).

The wrapper never imports, vendors, or submodules a leaf. Submodules
add real friction (clone steps, detached HEAD, push-from-inside) for
~zero value when both checkouts are already on disk.

### 2. Per-leaf wrappers are FORBIDDEN when an epic exists

The epic absorbs all per-leaf process state. Do not also create
`<leaf-1>-dev/` alongside `<product>-epic/`. One wrapper, one
`.work/`, one place to look.

If the product evolves so much that a single epic becomes unwieldy,
that is the signal to _split the product_, not to add per-leaf
wrappers underneath the epic.

### 3. `.work/` gitignored; cross-repo + per-leaf items live here

The wrapper's `.work/` holds **both** cross-repo features and per-leaf
items. Per-leaf items name the target in the title (`[<leaf-N>] M0 …`)
so views are scannable. Same store, same workflow, no per-leaf drift.

Gitignoring `.work/` keeps the convention identical to the brain and
dev-wrapper patterns: intake is local maintainer state, not durable
audit history. Auditing happens through committed artifacts (commits,
PRs, design docs).

## File Inventory

```
<product>-epic/
├── .conventions.yaml      # declared opt-ins, including epic.leaves
├── .gitignore             # excludes .work/, bin/, repo/<leaf> symlinks
├── .work/                 # `work --store .work init`; gitignored
├── CLAUDE.md              # agent contract (mirror to other active docs when needed)
├── README.md              # one-page overview + first-time bootstrap
├── Taskfile.yml           # verify, work, triage, build, up, e2e, bootstrap
├── docs/
│   ├── README.md
│   ├── requests/README.md
│   ├── planning/README.md
│   ├── plans/README.md
│   ├── implementation/README.md
│   └── taxonomy/README.md      # REQUIRED — verify.sh asserts this
├── scripts/
│   ├── verify.sh               # canonical from agent-repo-kit
│   └── bootstrap.sh            # recreates symlinks from .conventions.yaml
├── versions.yaml               # REQUIRED: pinned tag per leaf
├── compose.yaml                # recommended: composed runtime stack
├── e2e/                        # REQUIRED: black-box scenarios (.gitkeep ok)
├── repo/
│   └── .gitkeep                # symlink namespace anchor; links are ignored
└── (stack-specific workspace, e.g. go.work, pnpm-workspace.yaml)
```

Required count: 17 files + 1 stack-workspace if applicable (~18). The
two new files vs the old dev-wrapper inventory are
`scripts/bootstrap.sh` and `versions.yaml`; `repo/.gitkeep`, `e2e/`,
and the stack workspace are also lifted from optional to required.

## `.conventions.yaml` Extensions

The epic pattern adds one well-known top-level block:

```yaml
epic:
  leaves:
    - <leaf-1> # name only; ../<leaf> and repo/<leaf> are implied
    - <leaf-2>
    - <leaf-3>
  # optional: assertion that compose.yaml lists each service
  composed: true
```

`scripts/verify.sh` reads `epic.leaves` and asserts each `../<leaf>` is
a directory and the corresponding `repo/<leaf>` symlink resolves to
`../../<leaf>`. Missing leaves print
`verify: epic.leaves: ../<leaf> not found — run scripts/bootstrap.sh`.

## Templates

### `.conventions.yaml`

```yaml
# <product>-epic: dev wrapper + cross-repo orchestration for <product>.
# Wraps N sibling target repos. Read by the convention-engineering skill.
# Schema: ../agent-repo-kit/skills/convention-engineering/references/core/conventions-yaml.md.

agent_docs:
  - CLAUDE.md

docs_root: docs

taskfile: true

pre_commit: false

operations:
  - work

# min_work_version intentionally omitted: this repo IS the dev environment.
# Maintainer often runs unreleased work CLI builds; release-floor would
# gate legitimate dev work.

epic:
  leaves:
    - <leaf-1>
    - <leaf-2>
    - <leaf-3>

checks:
  - "CLAUDE.md exists; mirror to AGENTS.md by hand if both files are active in this repo."
  - "docs/{requests,planning,plans,implementation,taxonomy} directories exist."
  - "task verify exits 0 from a clean checkout (delegates to scripts/verify.sh)."
  - "scripts/verify.sh asserts declared opt-ins, including epic.leaves."
  - "repo/<leaf> symlinks exist for every epic.leaves entry and resolve to sibling checkouts."
  - ".work/ exists with config.yaml; task work -- view ready --json succeeds."
  - "Inbox is triaged regularly; entries older than 30 days without an action plan get closed."
  - "scripts/bootstrap.sh exists; recreates symlinks from epic.leaves."
  - "versions.yaml lists every entry under epic.leaves."
```

### `.gitignore`

```
.work/
.docs/         # private scratch parallel to .work/; never tracked
bin/
.DS_Store

# epic.leaves symlinks under repo/ — recreated by scripts/bootstrap.sh
# (List explicitly once leaves are known so unexpected repo/ files still surface.)
repo/<leaf-1>
repo/<leaf-2>
repo/<leaf-3>
```

`.docs/` is the private scratch convention: same posture as `.work/`
(local intake, never committed). The canonical `docs/` (no dot) is for
tracked design artifacts and is governed by `docs_root` in the
descriptor.

### `scripts/bootstrap.sh`

```bash
#!/usr/bin/env bash
# Recreate the sibling symlinks declared in epic.leaves under repo/.
# Idempotent. Skips a leaf cleanly if its sibling dir is missing.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

command -v yq >/dev/null 2>&1 || {
  echo "bootstrap: yq required (https://github.com/mikefarah/yq)" >&2
  exit 1
}

mapfile -t leaves < <(yq '.epic.leaves[]? // ""' .conventions.yaml | sed '/^$/d')

if [ "${#leaves[@]}" -eq 0 ]; then
  echo "bootstrap: no epic.leaves declared in .conventions.yaml" >&2
  exit 1
fi

mkdir -p repo

for leaf in "${leaves[@]}"; do
  if [ ! -d "../$leaf" ]; then
    echo "warn: ../$leaf not found — clone it as a sibling and re-run" >&2
    continue
  fi
  link="repo/$leaf"
  if [ -e "$link" ] && [ ! -L "$link" ]; then
    echo "skip: $link exists and is not a symlink — leaving it alone" >&2
    continue
  fi
  ln -sfn "../../$leaf" "$link"
  echo "linked: $link -> ../../$leaf"
done
```

### `Taskfile.yml`

```yaml
version: "3"

tasks:
  default:
    cmds:
      - task --list

  bootstrap:
    desc: Recreate repo/<leaf> symlinks from .conventions.yaml epic.leaves.
    cmds:
      - bash scripts/bootstrap.sh

  verify:
    desc: Run the canonical verification gate (declared opt-ins).
    cmds:
      - bash scripts/verify.sh

  work:
    desc: Run the work CLI against this repo's .work store.
    preconditions:
      - sh: command -v work >/dev/null 2>&1
        msg: "`work` not on PATH; install the `work` CLI before using these tasks"
    cmds:
      - work --store .work {{.CLI_ARGS}}

  triage:
    desc: List the inbox to drive the weekly triage pass.
    preconditions:
      - sh: command -v work >/dev/null 2>&1
        msg: "`work` not on PATH; install the `work` CLI before using these tasks"
    cmds:
      - work --store .work inbox
      - echo
      - echo "Inspect    -- task work -- show IN-NNNN"
      - echo "Accept     -- task work -- triage accept IN-NNNN --area <area> --priority P[0-3]"
      - echo "Direct     -- task work -- new \"title\" --area <area> --priority P[0-3]"

  # Stack-specific tasks (build / up / e2e) live below; shape varies by product.
```

### `CLAUDE.md` skeleton

```markdown
# CLAUDE.md — <product>-epic

Dev wrapper + cross-repo orchestration workspace for <product>. Holds
intake (`.work/`), design docs, deployment artifacts, the E2E harness,
and ops glue. Does **not** contain <product> source — that lives in
sibling checkouts exposed through the wrapper-local repo namespace:

`repo/<leaf-1>/`, `repo/<leaf-2>/`, `repo/<leaf-3>/`

After clone, run `task bootstrap` to recreate the `repo/<leaf>` symlinks.

## Architecture

- `.work/` — gitignored. Cross-repo + per-leaf work items. Per-leaf items
  name the target in the title (e.g. `[<leaf-2>] M0 — dial server …`).
- `docs/` — requests/planning/plans/implementation/taxonomy.
- `versions.yaml` — pinned commit/tag per leaf.
- `compose.yaml` — local composed stack for `task up`.
- `e2e/` — black-box scenarios that exercise the whole system.
- `repo/<leaf-N>` — symlinks to sibling `../<leaf-N>`; gitignored,
  recreated by `task bootstrap`.

## Non-Negotiable Rules

- `.work/` is gitignored; intake is local maintainer state.
- `repo/<leaf>` symlinks are gitignored; recreate via `task bootstrap`
  after clone.
- Per-leaf `-dev` wrappers are FORBIDDEN — this epic absorbs them.
- This repo MUST NOT contain <product> source; edit via `repo/` symlinks.
- `epic.leaves` and `versions.yaml` must list the same leaves; verify.sh
  enforces this.
```

## Execution Steps

1. **Create the directory.** Create `<workspace>/<product>/<product>-epic`
   under the same parent as the leaf repos, then `cd` into it.
2. **Write the file inventory** from the templates above. Substitute
   `<product>` and `<leaf-N>` throughout.
3. **Initialize the work store.** `work --store .work init`. If a `.work/`
   already exists from prior local state, `mv` it here instead of running
   `init`.
4. **Recreate `repo/<leaf>` symlinks for the first time.** `bash scripts/bootstrap.sh`.
   At least one leaf should already exist as a sibling directory; if not,
   the script will warn but still succeed.
5. **Smoke the gate.** `git init -b main && bash scripts/verify.sh`. Must
   print `verify: opt-ins ok`. The most common first-failure is a missing
   `docs/taxonomy/README.md` — it's required and easy to forget.
6. **Initial commit.**
   ```bash
   git add -A
   git commit -m "chore: bootstrap epic-wrapper for <product>"
   ```

## Migrating From Per-Leaf `-dev` Wrappers

If you have `<leaf-1>-dev/`, `<leaf-2>-dev/` etc. from the older
dev-wrapper pattern:

1. Create `<product>-epic/` per the steps above.
2. Move each `<leaf-N>-dev/.work/items/*.yaml` and matching
   `.work/spaces/<W-ID>/` into `<product>-epic/.work/`. Renumber IDs if
   they collide. Prefix titles with `[<leaf-N>]` if not already.
3. Move design docs from each `<leaf-N>-dev/docs/` into
   `<product>-epic/docs/`, prefixing filenames with `<leaf-N>_` to
   preserve provenance.
4. `git rm -rf <leaf-N>-dev` (after committing the migration to the
   epic). Or archive each `-dev` repo to `_archive/` if you need the
   history.

## Migrating Committed `.work/` to Gitignored

If `.work/` is currently tracked in git (common when someone scaffolded
without realizing), the gitignore alone doesn't untrack it:

```bash
git rm -r --cached .work/
echo ".work/" >> .gitignore
git commit -m "chore: untrack .work/ (now local intake only)"
```

## Per-Realm Gate Matrix

Not applicable — the epic pattern doesn't use realms. (This section
exists only in the brain pattern.)

## Gotchas Observed In Practice

- **`docs/taxonomy/` is required**, not optional. The reference
  `scripts/verify.sh` from `agent-repo-kit` asserts all five canonical
  subdirs (`requests/`, `planning/`, `plans/`, `implementation/`,
  `taxonomy/`). The bootstrap-workflow doc historically only mentioned
  four. Create all five.

- **Untracking `.work/` requires `--cached`**. After `.gitignore`-ing
  `.work/`, run `git rm -r --cached .work/` to actually stop tracking
  it. A gitignore edit alone leaves the existing files tracked.

- **`go-task` v3 mis-parses bare strings with colons inside `cmds:`.**
  An unquoted `echo "TODO: foo"` parses as a YAML mapping. Use `--`
  instead of `:` in echo strings, or single-quote the entire scalar.

- **`yq` must be the Go (mikefarah) version**, not the Python wrapper
  around `jq`. The verify and bootstrap scripts use `mikefarah/yq` flag
  syntax. Install via Homebrew or the GitHub releases page.

- **Symlinks resolve relative to the symlink's location**, not your
  cwd. `ln -s ../../<leaf> repo/<leaf>` inside `<product>-epic/`
  resolves to `<product>-epic/repo/../../<leaf>` = the sibling. This
  works regardless of whether you `cd` into the symlink before using it.

- **Workspace-file paths are relative to the workspace file**, not the
  symlink. A root `go.work`, `pnpm-workspace.yaml`, or equivalent should
  use either sibling paths such as `../<leaf>` or wrapper-local paths
  such as `./repo/<leaf>` deliberately; do not copy the symlink target
  string blindly.

## Anti-Patterns

- **Committing the symlinks.** Silent dangle on clone if leaves aren't
  placed correctly. Always `task bootstrap` instead.

- **Adding per-leaf `-dev` wrappers alongside the epic.** Pick one;
  the epic is canonical. Per-leaf wrappers fragment the work tracker
  and make cross-repo coordination invisible.

- **Adding `min_work_version`.** This repo IS the dev environment;
  pinning gates legitimate dev work.

- **Adding leaf source as a submodule.** Submodules add a clone step
  and detached-HEAD pain for zero benefit when both checkouts are on
  disk.

- **Hand-editing leaf symlinks.** Add/remove leaves via
  `.conventions.yaml epic.leaves`, then `task bootstrap`. Hand-edits
  under `repo/` drift from the descriptor and verify.sh will catch them
  next run.

- **Storing leaf-specific implementation in the epic.** If it's code
  that ships, it lives in a leaf. The epic is markdown + yaml + shell.

## Worked Example

`<product>-epic/` wrapping `<leaf-1>`, `<leaf-2>`, and `<leaf-3>` is the
canonical shape. Diff a concrete wrapper against this doc when refining.
