# AGENTS.md — agent-repo-kit

You are inside **agent-repo-kit**. This repo publishes a convention for
_other_ repos to adopt, and it also adopts that same convention on itself
(see the `## Conventions` block below) — so `ark check --repo-root .` and
`task verify` both run here.

## Entry points

- `skills/` — canonical, harness-free skill sources. One directory per
  skill:
  - `skills/convention-engineering/` — content describing repo conventions
    (tickets, wiki, agent docs, verification gates, etc.). Canonical source.
  - `skills/convention-evaluator/` — scoring rubric used to grade a repo's
    adoption against the contract.
  - `skills/skill-builder/` — skill for authoring and auditing agent skills
    (trigger wording, portable structure, reference extraction, runtime
    placement).
  - `skills/taskfile-authoring/` — skill for writing canonical Taskfiles
    (structure, composition, anti-patterns, lint rules). Referenced by
    `ark taskfile lint`.
  - `skills/attack-architecture/` — adversarial architecture-review skill.
    Runs parallel lens attacks, ToT expansion, and attacker/defender debate
    against an existing codebase and writes a report under
    `.docs/arch-attacks/`.
- `examples/demo-repo/` — a working repo that shows the conventions applied
  end to end; the CI exercises it.
- `adapters/<harness>/` — thin shims that expose every skill under
  `skills/` to a specific harness. `claude-code/` and `codex/` are shipped
  install targets; `cursor/` is placeholder docs.
- `adapters/manifest.json` — machine-readable source of truth for which
  skill directories get symlinked into which harness. Consumed by
  `ark adapters link` and `ark adapters list-links`.
- `install.sh` — POSIX installer. Default path: download the prebuilt
  `ark` binary for the current OS/arch from the latest GitHub Release,
  drop it in `--prefix` (default `~/.local/bin`), then call
  `ark adapters link --target <harness>` to wire the symlinks. Pass
  `--from-source` to build `cli/` locally instead.
- `ark upgrade` — in-place upgrade. Detects whether `ark` lives inside a
  git clone (runs `git pull` + rebuild) or was installed from a release
  archive (downloads + atomically replaces the binary), then re-runs
  `adapters link`.
- `.goreleaser.yml` + `.github/workflows/release.yml` — release pipeline
  that publishes `ark-{version}-{os}-{arch}.tar.gz` + `checksums.txt` on
  each `v*` tag.

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
3. **Dual-write pointer blocks** — when adding a new convention, update
   both `examples/demo-repo/AGENTS.md` and `examples/demo-repo/CLAUDE.md`
   identically.
4. **Adapters re-export, they don't own.** An adapter file should be a
   short wrapper pointing at a skill under `skills/`.

## Testing

```bash
task -d examples/demo-repo/.tickets test   # expect 10/10
task -d examples/demo-repo/.wiki lint      # expect OK
```

CI runs both on every push and PR (see `.github/workflows/ci.yml`).

<!-- agent-repo-kit:init:start -->

## Conventions

- **Docs** — tracked repo docs live under `docs/` using the `requests/`,
  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.
- **Tickets** — flat-file work tracker at `.tickets/`. Read `.tickets/README.md`
  for the verb surface and `.tickets/harness/schema.yaml` for the state
  machine. Daily commands:
  `task -d .tickets {new|list|transition|close|test}`.
- **Wiki** — LLM-maintained knowledge base at `.wiki/`. Read `.wiki/RULES.md`
  for page types, frontmatter, and citation rules. Validate with
  `task wiki:lint` (or `task -d .wiki lint`).
- **Verification** — run `task verify` from the repo root to execute the
  convention verification gate.
- **Tracked contract** — `.convention-engineering.json` is the
  machine-readable convention contract for this repo.

Conventions are scaffolded by `agent-repo-kit` under `.convention-engineering/`.

<!-- agent-repo-kit:init:end -->
