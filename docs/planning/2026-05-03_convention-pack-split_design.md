# Convention Pack vs Adoption — Split Decision RFC

Status: discussion only. Cross-checked against the kit's actual layout
and recent commit history. Does not commit to a split.

## The Problem

`agent-repo-kit` does two jobs:

1. **Publishes the convention pattern.** Canonical sources under
   `skills/convention-engineering/`, `skills/convention-evaluator/`,
   the `.conventions.yaml` schema doc, and the harness adapters under
   `adapters/<harness>/`.
2. **Adopts that pattern on itself.** Its own `.conventions.yaml`,
   `AGENTS.md`/`CLAUDE.md`, root `Taskfile.yml`, `scripts/verify.sh`,
   `.github/workflows/ci.yml`, `docs/` taxonomy.

A change can be either "the pattern got better" or "this kit's adoption
got better" — and the same PR routinely contains both. During the
W-0010 refactor (the recent five-commit epic) every commit touched the
pattern surface; several also moved adoption files in the same diff.

That is a conflation cost. It hurts:

- **External adopters.** A new repo wanting "the convention" reads the
  whole kit and has to filter out the kit's own dogfood material.
- **Pinning.** An adopter cannot pin to "convention pattern v1.3" without
  also pinning to a specific kit-adoption snapshot. Versioning the kit
  conflates two release cadences.
- **Reviewability.** Reviewers cannot tell from the diff whether a change
  is a pattern release (high-blast-radius) or a kit-local fix (low).
- **Multi-agent reach.** The pattern surface is supposed to be
  harness-agnostic. Mixing it with the kit's own adoption (which uses
  Claude/Codex specifically) makes that less obvious.

## Cost Today: Commit Classification

Last 30 days, classified by surface touched (excluding 22 docs/intent-only
commits):

| Class    | Count | Share |
| -------- | ----- | ----- |
| PATTERN  | 15    | 41%   |
| BOTH     | 16    | 44%   |
| ADOPTION | 5     | 14%   |

44% of working commits touched both surfaces. Heuristic from the original
suggestion: `<20%` → split is overkill; `>50%` → split is overdue. We are
in the middle. The conflation is real, the cost is not yet daily-painful.

## Options

### (a) Split into two repos

`agent-conventions/` (pattern):

- `skills/convention-engineering/`, `skills/convention-evaluator/`
- `skills/skill-builder/`, `skills/skill-evolution/`,
  `skills/taskfile-authoring/`, `skills/harness-router/`,
  `skills/work-cli/`, `skills/attack-architecture/`,
  `skills/paper-vetting/`, `skills/go-scripting/`
- `adapters/<harness>/` mirror surface
- `schemas/conventions.schema.json` (when it exists)
- A pattern-level `Taskfile` for releasing skill bundles
- No `.conventions.yaml` of its own (it _is_ the pattern)

`agent-repo-kit/` (the kit's own adoption + dogfood):

- `.conventions.yaml`
- `AGENTS.md`, `CLAUDE.md`, `Taskfile.yml`, `scripts/verify.sh`
- `docs/` taxonomy + reviews
- A pinned `agent-conventions` version (manifest, git submodule, or
  npx-skills tag)
- Self-evaluation reports under `docs/reviews/`

Pros: each repo has one job; pattern can release on its own cadence;
external adopters consume `agent-conventions` cleanly; reviewers see one
surface per PR.

Cons: every cross-cutting change becomes two PRs; the kit becomes thin
and may feel pointless if nothing else dogfoods it; submodule hygiene
costs.

### (b) Single repo, clearer subdir contract

Keep one repo but make the dual purpose mechanical:

- `pattern/` (or keep current `skills/` + `adapters/` + future
  `schemas/`) is the published surface. Treated as if vendored.
- Root-level files (`.conventions.yaml`, `AGENTS.md`, `CLAUDE.md`,
  `Taskfile.yml`, `scripts/verify.sh`, `docs/`) are this kit's adoption.
- Add a CI check that flags PRs touching both, prompting "split this if
  unrelated."
- Tag releases as `pattern-v1.3` vs `kit-v1.3` to give them independent
  cadence even inside one repo.

Pros: zero migration cost; preserves the natural co-evolution when both
surfaces _do_ need to move together (the 44% case); release tags give
adopters something to pin.

Cons: discipline depends on convention rather than mechanism; new readers
still parse "what is the pattern?" out of a mixed tree.

### (c) Status quo

Document the dual purpose in `AGENTS.md` and live with it.

Pros: zero work.

Cons: the conflation cost compounds as more repos adopt and as the
pattern stabilises into something with external adopters.

## Recommendation (no commitment)

**Lean (b) for now**, with a tripwire to revisit (a):

- Take (b)'s subdir discipline + dual tags. ~1 day of work to set up
  release-tag aliases and CI guard.
- Set the tripwire: split to (a) when **either** is true:
  - First external adopter shows up (someone else's repo consumes
    `agent-conventions`).
  - The BOTH-class share crosses 50% over a 60-day window. Currently
    44%; one more cycle of feature work like W-0010 could push it over.

(c) is rejected. Even if no work is needed today, the absence of a clear
boundary will keep producing mixed PRs.

## Multi-Agent Compatibility (cross-cutting concern)

The pattern surface is supposed to be harness-agnostic. Today only
`adapters/claude-code/` and `adapters/codex/` ship installers; `cursor/`
is placeholder docs. The convention-engineering description even says
"stack-agnostic" but the adapter coverage is two-of-many.

Other coding agents and their discovery models:

| Harness     | Skill discovery surface                                        |
| ----------- | -------------------------------------------------------------- |
| Claude Code | `~/.claude/skills/<name>/SKILL.md` (frontmatter)               |
| Codex       | `~/.codex/skills/<name>/SKILL.md` (frontmatter; same shape)    |
| Cursor      | `.cursor/rules/*.md` MDC files at repo root                    |
| Copilot     | `.github/copilot-instructions.md` (single repo file)           |
| Aider       | conventions live in `CONVENTIONS.md` referenced via `--read`   |
| Continue    | `.continue/config.yaml` + custom slash commands                |
| Gemini Code | `.idx/airules.md` or codespace-style repo files                |
| Generic     | `AGENTS.md` (the proposal floor — already part of the pattern) |

Implications for the split decision:

- **The pattern repo should explicitly support new harnesses** by
  documenting the adapter contract — what an adapter must produce given
  a canonical `skills/<name>/SKILL.md`. This contract today is implicit
  ("hand-mirror SKILL.md into the right path"); making it explicit
  enables third parties to add Cursor/Copilot/Aider adapters without
  forking.
- **Adapter shape varies wildly.** Claude/Codex consume per-skill
  directories. Copilot consumes one big file. Aider consumes a
  CONVENTIONS.md and runtime flags. The adapter contract has to allow
  for "fan out one skill per file" and "fold all skills into one file."
- **AGENTS.md is the universal floor.** Almost every harness, including
  the ones without a skill discovery model, can read `AGENTS.md` at the
  repo root. The pattern should treat AGENTS.md as the lowest-common-
  denominator surface and any harness-specific adapter as an
  optimisation on top.
- **Whether (a) or (b)**, the pattern's adapter contract is what makes
  multi-agent compatibility real. Without an explicit contract, every
  new harness requires upstream changes; with one, harness support
  becomes a satellite concern.

This argues mildly **for (a) eventually**. A pure pattern repo could
publish a small adapter-author CLI / skill, and harness adapters could
live in their own packages (`agent-conventions-cursor`,
`agent-conventions-copilot`). It does not change the recommendation for
**now**, but it should shape (b)'s subdir contract: keep
`adapters/<harness>/` symmetric so it can later become its own repo per
harness.

## What To Do Next (no implementation)

1. Pick (b) or status quo. If (b), draft the subdir contract: what files
   live where, what release tags exist, what CI guard fires.
2. Write the **adapter contract** as a separate doc under
   `skills/convention-engineering/references/core/adapter-contract.md`.
   This is independent of the split — it improves multi-agent reach
   either way.
3. Set a 60-day calendar reminder to re-classify commits and check the
   tripwire conditions for (a).

## Out Of Scope

- Any actual repo split.
- Any change to skill content.
- Concrete implementation of new harness adapters (covered by the
  separate adapter-contract work).
