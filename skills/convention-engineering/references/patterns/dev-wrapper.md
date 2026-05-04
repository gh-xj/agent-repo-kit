# Pattern: Dev Wrapper Repo

Sibling repo that holds maintainer-private process state for an OSS tool.
The OSS repo stays minimal (code, install, release). The dev wrapper holds
the `.work/` tracker, design docs, planning artifacts, and verification glue.

Codified from the work-cli-dev bootstrap (2026-05-03). Canonical instance:
`gh-xj/work-cli-dev` for `gh-xj/work-cli`.

## When To Use

Apply this pattern when:

- You maintain a public OSS tool whose repo should stay clean for adopters.
- You want a place for **maintainer-only** intake, design notes, and triage
  that doesn't pollute the OSS repo's history or surface area.
- You already have an opinionated work tracker (`.work/`) and want to dogfood
  it for evolving the tool itself.

Skip when:

- The OSS repo is small enough that a `docs/` folder + GitHub Issues is
  sufficient.
- The tool has a public roadmap; that wants to stay in the OSS repo.
- You'd be the only maintainer and you don't already use `.work/` daily.

## Three Load-Bearing Decisions

These are the decisions that actually matter. Don't deviate without a
reason recorded in `CLAUDE.md`.

### 1. Sibling layout, not git submodule

The OSS source and the dev wrapper are sibling directories on disk:

```
~/github/gh-xj/
├── <tool>/             # the OSS source (e.g. work-cli)
└── <tool>-dev/         # the dev wrapper (this pattern)
```

The wrapper never imports, vendors, or submodules the source. Submodules
add real friction (clone steps, detached HEAD, push-from-inside) for ~zero
value when both checkouts are already on disk.

### 2. Operations: one (`work`)

`.work/` is the centerpiece — that's why this repo exists. No `wiki`. Add
others only if the wrapper grows beyond a single tool's process state.

### 3. Omit `min_work_version`

This repo IS the dev environment. The maintainer constantly runs
unreleased builds from `../<tool>/`. Pinning to a release tag would gate
legitimate dev work. The verify script will skip the version gate when the
key is absent.

## File Inventory

12 files, single initial commit. Substitute `<tool>` and `<Tool>` for the
project name (e.g. `work-cli` and `work-cli`; or `foo` and `Foo`).

```
<tool>-dev/
├── .conventions.yaml      # declared opt-ins
├── .gitignore             # excludes .work/, .DS_Store
├── .work/                 # init via `work --store .work init`
├── CLAUDE.md              # agent contract (mirror to AGENTS.md if Codex drives this repo)
├── README.md              # one-page overview
├── Taskfile.yml           # verify, work, triage
├── docs/
│   ├── README.md
│   ├── requests/README.md
│   ├── planning/README.md
│   ├── plans/README.md
│   ├── implementation/README.md
│   └── taxonomy/README.md
└── scripts/verify.sh      # copy from agent-repo-kit; reads .conventions.yaml
```

## Templates

### `.conventions.yaml`

```yaml
# <tool>-dev: maintainer dev wrapper for gh-xj/<tool>.
# Read by the convention-engineering skill (in gh-xj/agent-repo-kit).
# Schema: ../agent-repo-kit/skills/convention-engineering/references/core/conventions-yaml.md.

agent_docs:
  - CLAUDE.md

docs_root: docs

taskfile: true

pre_commit: false

operations:
  - work

# min_work_version intentionally omitted: this repo IS the development
# environment for <tool>. The maintainer often runs an unreleased build
# from ../<tool>/, so a release-tag floor would gate legitimate dev work.

checks:
  - "CLAUDE.md exists; mirror to AGENTS.md by hand if Codex starts driving this repo."
  - "docs/{requests,planning,plans,implementation,taxonomy} directories exist."
  - "task verify exits 0 from a clean checkout (delegates to scripts/verify.sh)."
  - "scripts/verify.sh asserts declared opt-ins."
  - ".work/ exists with config.yaml; task work -- view ready --json succeeds."
  - "Inbox is triaged regularly; entries older than 30 days without an action plan get closed."
  - "../<tool>/ exists as a sibling checkout (sibling layout, not submodule)."
```

### `.gitignore`

```
.work/
.DS_Store
```

### `CLAUDE.md`

```markdown
# CLAUDE.md — <tool>-dev

Maintainer-private development wrapper for
[`gh-xj/<tool>`](https://github.com/gh-xj/<tool>). Holds intake (`.work/`),
design and planning docs, and ops glue. Does **not** contain `<tool>` source —
that lives as a sibling checkout at `../<tool>/`.

This repo is governed by the convention-engineering skill that ships in
[`gh-xj/agent-repo-kit`](https://github.com/gh-xj/agent-repo-kit). See
`.conventions.yaml` for declared opt-ins.

## Architecture

- `.work/` — inbox + triaged work items for evolving `<tool>` (gitignored).
- `docs/` — design (`planning/`), executable plans (`plans/`), implementation
  logs (`implementation/`), structured requests (`requests/`).
- `Taskfile.yml` + `scripts/verify.sh` — verification gate.

The actual `<tool>` source is at `../<tool>/` on disk. This repo never
imports or vendors it.

## Commands

\`\`\`bash
task verify # asserts declared opt-ins via scripts/verify.sh
task work -- inbox # see captured intake
task work -- view ready
task triage # short helper for the weekly triage pass
\`\`\`

## Non-Negotiable Rules

- `.work/` is gitignored — local intake state, never committed.
- `.conventions.yaml` and this `CLAUDE.md` are the source of truth for repo
  conventions; keep them in sync by hand.
- This repo MUST NOT contain `<tool>` source. Edit `../<tool>/`, not here.
- Inbox captures must include a source URL and a one-sentence rationale.
- `min_work_version` is intentionally not declared.

## Verification Gate

\`\`\`bash
task verify
\`\`\`

## Pointers

- Convention skill source: `../agent-repo-kit/skills/convention-engineering/SKILL.md`
- `<tool>` source on disk: `../<tool>/`
- Public OSS repo: https://github.com/gh-xj/<tool>
```

### `Taskfile.yml`

```yaml
version: "3"

tasks:
  verify:
    desc: Run the repo's canonical verification gate
    cmds:
      - bash scripts/verify.sh

  work:
    desc: Run the work CLI against this repo's .work store
    preconditions:
      - sh: command -v work >/dev/null 2>&1
        msg: "`work` not on PATH; install from https://github.com/gh-xj/work-cli or build from ../work-cli/"
    cmds:
      - work --store .work {{.CLI_ARGS}}

  triage:
    desc: List the inbox to drive the weekly triage pass
    preconditions:
      - sh: command -v work >/dev/null 2>&1
        msg: "`work` not on PATH; install from https://github.com/gh-xj/work-cli"
    cmds:
      - work --store .work inbox
      - echo
      - echo "Inspect    -> task work -- show IN-NNNN"
      - echo "Accept     -> task work -- triage accept IN-NNNN --area <area> --priority P[0-3]"
      - echo "Direct     -> task work -- new \"title\" --area <area> --priority P[0-3]"
```

### `scripts/verify.sh`

Copy verbatim from `agent-repo-kit/scripts/verify.sh`. The script already
handles missing `min_work_version` (skips the gate when the key is absent).
No project-specific edits needed.

### `docs/` README stubs

Six small files. The root `docs/README.md` lists the four sub-folders;
each sub-README states its filename contract (see `references/core/docs-taxonomy.md`).

## Execution Steps

1. **Create the directory.** `mkdir -p ~/github/gh-xj/<tool>-dev && cd $_`
2. **Write the 12 files** from the templates above. Substitute `<tool>`.
3. **Initialize the work store.** `work --store .work init`. If a `.work/`
   already exists from prior local state (e.g. you had it inside the OSS
   repo before this pattern), `mv` it here instead of running `init`.
4. **Smoke the gate.** `git init -b main && bash scripts/verify.sh`. Must
   print `verify: opt-ins ok`. If `taxonomy/` is the failing line, the
   bootstrap-workflow doc skipped it but the verify script requires it —
   add `docs/taxonomy/README.md`.
5. **Initial commit.**
   ```bash
   git add -A
   git commit -m "chore: bootstrap convention-engineering wrapper"
   ```
6. **Optional: publish the remote.** Confirm with the user before
   `gh repo create` — publishing (even private) is a "visible to others"
   action.

## Gotchas Observed In Practice

- `scripts/verify.sh` requires a git repo (uses `git rev-parse
--show-toplevel`). Run `git init` before the first verify.
- The bootstrap-workflow lists four docs subdirs (`requests/`, `planning/`,
  `plans/`, `implementation/`) but the verify script also asserts
  `taxonomy/`. Create all five.
- `.conventions.yaml` requires mikefarah/yq (Go), not Python yq, for
  consistent string output. The script doesn't enforce which yq is
  installed; flag this if a future `task verify` produces quoted strings.

## Anti-Patterns

- **Adding `min_work_version`**: defeats the purpose. The wrapper is for
  iterating on the binary, not gating on a release.
- **Adding the source as a submodule**: the source lives in its own repo
  with its own release cadence; the wrapper tracks process, not code.
- **Adding agent-side code (Go/Python/etc.)**: this repo is Markdown + YAML
  - a shell script. Resist scope creep. If you find yourself wanting code
    here, ask whether it belongs in the OSS source instead.

## Worked Example

`~/github/gh-xj/work-cli-dev/` (initial commit) is the canonical instance
of this pattern. Diff against this doc when refining.
