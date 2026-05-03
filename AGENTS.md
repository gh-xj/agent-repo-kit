# AGENTS.md — agent-repo-kit

You are inside **agent-repo-kit**. This repo publishes a convention for
_other_ repos to adopt, and it also adopts that same convention on itself
(see the `## Conventions` block below) — so `task verify` runs here.

## Core belief

Before changing conventions, persistent state models, agent workflows, or
repo-wide architecture, read `docs/core-belief.md`. It is the philosophical
north star for this repo; this file is the operational map.

## Entry points

- `skills/` — canonical, harness-free skill sources. One directory per
  skill:
  - `skills/convention-engineering/` — content describing repo conventions
    (`.conventions.yaml`, agent docs, docs taxonomy, verification gates,
    work tracking, optional wiki). Canonical source.
  - `skills/convention-evaluator/` — scoring rubric used to grade a repo's
    adoption of its declared conventions.
  - `skills/skill-builder/` — skill for authoring and auditing agent skills
    (trigger wording, portable structure, reference extraction, runtime
    placement).
  - `skills/taskfile-authoring/` — skill for writing canonical Taskfiles
    (structure, composition, anti-patterns).
  - `skills/attack-architecture/` — adversarial architecture-review skill.
  - `skills/harness-router/` — proposal-only router for deciding where
    session learnings, user corrections, and harness improvements should
    persist across instructions, skills, docs, work records, memory, and
    verification surfaces.
  - `skills/work-cli/` — operating the `.work/` tracker.
  - `skills/paper-vetting/` — vet a research paper through three
    independent lenses (team / citation context / claim-level evidence)
    before reading it; outputs a calibrated trust band and falsifier.
- `cli/` — Go source for the `work` CLI.
- `adapters/<harness>/` — thin shims that expose every skill under
  `skills/` to a specific harness. `claude-code/` and `codex/` are shipped
  install targets; `cursor/` is placeholder docs.
- `adapters/manifest.json` — machine-readable source of truth for which
  skill directories belong to which harness.

## Rules for editing this repo

1. **Do not** add harness-specific frontmatter (e.g. Claude skill YAML) to
   files under `skills/convention-engineering/` or
   `skills/convention-evaluator/`. That belongs in
   `adapters/claude-code/SKILL.md` and equivalents.
   `skills/skill-builder/SKILL.md` is the exception: its portable
   frontmatter (`name` + `description` only) is the skill's interface.
2. **Do not** reference absolute user paths like `/Users/...` or
   `~/.claude/` inside any skill surface. Those are environment specifics.
   `skills/convention-engineering/` and `skills/convention-evaluator/`
   must also avoid the harness names "Claude", "Skill tool", and "Codex"
   — but `skills/skill-builder/` and `skills/attack-architecture/` may
   name them since the runtimes (and their agent/tool APIs) are the
   subject matter of those skills.
3. **Adapters re-export, they don't own.** An adapter file should be a
   short wrapper pointing at a skill under `skills/`.

## Testing

```bash
task verify   # asserts declared opt-ins via scripts/verify.sh
```

CI runs `task verify` on every push and PR (see `.github/workflows/ci.yml`).
The job installs the external `work` binary and `yq` before running.

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Work** — local-first work tracker at `.work/`. The CLI is the external
  [`work-cli`](https://github.com/gh-xj/work-cli) binary; ARK no longer
  ships a copy. Install it (`go install github.com/gh-xj/work-cli/cmd/work@latest`
  or via the release tarball), then drive it through `task work -- ...`.
  Local state lives in the ignored `.work/config.yaml` and `.work/items/*.yaml`.
  Daily commands: `task work -- inbox`, `task work -- inbox add "title"`,
  `task work -- triage accept IN-0001`, `task work -- view ready`,
  and `task work -- show W-0001`.
- **Conventions descriptor** — `.conventions.yaml` at the repo root declares
  which conventions this repo opts into, including `min_work_version` for
  the external `work` CLI. Read by the convention-engineering skill for
  bootstrap and audit; enforced by `scripts/verify.sh`.
