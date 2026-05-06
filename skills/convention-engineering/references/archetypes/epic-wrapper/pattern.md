# Archetype: Epic Wrapper

Status: stable
Codified: 2026-05-03

A single sibling repo that owns the development process state for a multi-repo product. The product's leaf repos (one per shipped artifact — proto, daemon, server, cli, …) stay clean: code, install, release. The epic wrapper holds the `.work/` tracker, design docs, deployment artifacts, cross-repo verification glue, and a `repo/` symlink namespace for each leaf.

Replaces the older `dev-wrapper` archetype; that archetype's single-leaf case is now the N=1 degenerate of this one.

## When To Use

Apply when:

- You ship a product as N≥1 sibling repos (proto + daemon + server, or api + worker + admin-ui, etc.) — anything where a coordinated change spans more than one repo.
- The leaves should stay minimal for adopters / future contributors.
- You need exactly one place to answer "do these versions work together?"
- You want to dogfood `.work/` for evolving the product itself.

The N=1 case (one leaf) also applies — if you maintain a single OSS tool and want a maintainer-private wrapper, use this archetype with `leaves: [<tool>]`.

Skip when:

- The product fits in a single repo (no multi-repo coordination needed).
- The repo is small enough that `docs/` + GitHub Issues is sufficient.
- You'd be the only maintainer and you don't already use `.work/` daily.

## Three Load-Bearing Decisions

### 1. Sibling layout, `repo/` symlinks gitignored, bootstrapped from descriptor

Leaves and the wrapper are sibling directories on disk:

```
<workspace>/<product>/
├── <leaf-1>/             ← clean leaf (no .work/, no docs/)
├── <leaf-2>/
├── <leaf-3>/
└── <product>-epic/       ← THIS archetype
    └── repo/
        ├── <leaf-1>      → ../../<leaf-1>   (symlink, gitignored)
        ├── <leaf-2>      → ../../<leaf-2>   (symlink, gitignored)
        └── <leaf-3>      → ../../<leaf-3>   (symlink, gitignored)
```

The wrapper-local path is `repo/<leaf>` (root-relative shorthand: `/repo/<leaf>`). The symlinks are **gitignored**, not committed. They are recreated by `scripts/bootstrap.sh` after clone, which reads the leaf list from `.conventions.yaml`. Reasons:

- Committed symlinks dangle if the cloner doesn't place sibling repos in the right relative location. Silent corruption is worse than an explicit bootstrap step.
- `repo/` makes symlinked source checkouts an explicit namespace instead of letting leaf names compete with wrapper-owned files.
- The leaf list belongs in `.conventions.yaml` anyway (the schema formalizes it). Once it's there, the bootstrap script is a trivial consumer.
- Keeps git history free of layout-mechanics churn.

The wrapper never imports, vendors, or submodules a leaf. Submodules add real friction (clone steps, detached HEAD, push-from-inside) for ~zero value when both checkouts are already on disk.

### 2. Per-leaf wrappers are FORBIDDEN when an epic exists

The epic absorbs all per-leaf process state. Do not also create `<leaf-1>-dev/` alongside `<product>-epic/`. One wrapper, one `.work/`, one place to look.

If the product evolves so much that a single epic becomes unwieldy, that is the signal to *split the product*, not to add per-leaf wrappers underneath the epic.

### 3. `.work/` gitignored; cross-repo + per-leaf items live here

The wrapper's `.work/` holds **both** cross-repo features and per-leaf items. Per-leaf items name the target in the title (`[<leaf-N>] M0 …`) so views are scannable. Same store, same workflow, no per-leaf drift.

Gitignoring `.work/` keeps the convention identical to the brain and core-beliefs patterns: intake is local maintainer state, not durable audit history. Auditing happens through committed artifacts (commits, PRs, design docs).

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

Required count: 17 files + 1 stack-workspace if applicable.

## `.conventions.yaml` Extensions

The archetype adds one well-known top-level block:

```yaml
epic:
  leaves:
    - <leaf-1> # name only; ../<leaf> and repo/<leaf> are implied
    - <leaf-2>
    - <leaf-3>
  composed: true # optional: assert compose.yaml lists each service
```

`scripts/verify.sh` reads `epic.leaves` and asserts each `../<leaf>` is a directory and the corresponding `repo/<leaf>` symlink resolves to `../../<leaf>`. Missing leaves print `verify: epic.leaves: ../<leaf> not found — run scripts/bootstrap.sh`.

## Templates

Concrete scaffold files live in [`templates/`](./templates/):

- `conventions.yaml.tmpl` — descriptor with `epic.leaves` block
- `gitignore.tmpl` — repo `.gitignore` (excludes `.work/`, `.docs/`, `bin/`, `repo/<leaf>` symlinks)
- `bootstrap.sh` — recreates `repo/<leaf>` symlinks from `.conventions.yaml`
- `Taskfile.yml.tmpl` — root Taskfile with `verify`, `work`, `triage`, `bootstrap` targets
- `CLAUDE.md.tmpl` — agent contract skeleton

## Bootstrap

See [`bootstrap.md`](./bootstrap.md) for the step-by-step procedure.

## Migration

See [`migration.md`](./migration.md) if you're moving from per-leaf `-dev` wrappers or have a previously-committed `.work/` that needs to be untracked.

## Worked Example

`gh-xj/work-cli-epic` is a real instance (initial commit 2026-05-03; single leaf — `leaves: [work-cli]`). Diff a concrete wrapper against this doc when refining.

## Anti-Patterns

- **Committing the symlinks.** Silent dangle on clone if leaves aren't placed correctly. Always `task bootstrap` instead.
- **Adding per-leaf `-dev` wrappers alongside the epic.** Pick one; the epic is canonical. Per-leaf wrappers fragment the work tracker and make cross-repo coordination invisible.
- **Adding `min_work_version`.** This repo IS the dev environment; pinning gates legitimate dev work.
- **Adding leaf source as a submodule.** Submodules add a clone step and detached-HEAD pain for zero benefit when both checkouts are on disk.
- **Hand-editing leaf symlinks.** Add/remove leaves via `.conventions.yaml epic.leaves`, then `task bootstrap`. Hand-edits under `repo/` drift from the descriptor and verify.sh will catch them.
- **Storing leaf-specific implementation in the epic.** If it's code that ships, it lives in a leaf. The epic is markdown + yaml + shell.

## Gotchas

- **`docs/taxonomy/` is required**, not optional. The reference `scripts/verify.sh` asserts all five canonical subdirs (`requests/`, `planning/`, `plans/`, `implementation/`, `taxonomy/`).
- **`yq` must be the Go (mikefarah) version**, not the Python wrapper. The verify and bootstrap scripts use `mikefarah/yq` flag syntax. Install via Homebrew or the GitHub releases page.
- **Symlinks resolve relative to the symlink's location**, not your cwd. `ln -s ../../<leaf> repo/<leaf>` inside `<product>-epic/` resolves to `<product>-epic/repo/../../<leaf>` = the sibling. This works regardless of whether you `cd` into the symlink before using it.
