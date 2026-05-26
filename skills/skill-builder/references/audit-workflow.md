# Audit Workflow — three-phase recipe

For auditing an existing skill collection (one repo's `.claude/skills/` or `~/.claude/skills/`), this 3-phase order proved reliable. Each phase ends in 1-2 commits and is independently shippable.

Workflow added 2026-05-25 after a 9-skill audit of `xj-private-finance/.claude/skills/` (1464 lines total). Total session time: ~90 min for 9 skills + 5 commits across 2 repos.

## When to use this recipe

- Catalog has grown past ~5 skills.
- Multiple status mismatches between docs and actual files.
- One or more skills are over the 200-line "consider refactor" threshold.
- A skill's scope feels mismatched to its container repo.

If only one of the above, do the targeted fix directly — don't run the full recipe.

## Phase 1 — doc hygiene

**Goal**: get the agent-facing skill table to match `ls .claude/skills/` reality.

Findings to look for:

- Skills listed in CLAUDE.md but missing from disk → stale row, delete.
- Skills on disk but not in CLAUDE.md → add row.
- Skills marked "not yet" / "planned" that have actually shipped.
- Status info that's older than the latest behavior change.

Output: 1-2 small commits (one per affected repo's CLAUDE.md + AGENTS.md mirror). Lowest risk. Always start here — the cleaned table informs the next two phases.

## Phase 2 — per-skill size refactor

**Goal**: bring any over-200-line `SKILL.md` back under the router-rule threshold by extracting content to `references/`.

Triggers:

- `wc -l SKILL.md` returns ≥ 200 ("consider refactor") or ≥ 400 ("refactor now").
- Multiple inline sub-capabilities each with their own queries, examples, schemas.
- Recurring sections (e.g., "capability 1", "capability 2", ...) that could share a directory.

Typical extraction:

```
.claude/skills/<name>/
├── SKILL.md                              # router: triggers + when-to-invoke + index
└── references/
    └── capabilities/
        ├── <capability-1>.md            # query + answer-shape + caveats
        ├── <capability-2>.md
        └── ...
```

Keep the slim `SKILL.md` with:

- frontmatter (unchanged)
- when-to-invoke
- preconditions
- data sources / inputs router
- **capability INDEX** (one-line per ref + link)
- workflow / boundaries / verification

**No behavioral change** in this phase. Same triggers, same workflow, same boundaries — just relocated. One commit per skill refactored.

## Phase 3 — scope-mismatch lift

**Goal**: move skills whose target data is broader than the hosting repo to the runtime's global home (`~/.claude/skills/`).

Lift criteria (any one is sufficient):

1. The skill reads/writes data outside the hosting leaf.
2. The skill is stack-generic (e.g., a `gws-gmail` style wrapper bound to a local account, not a project).
3. Multiple leaves want to invoke it.

Mechanics:

- `cp -r` from project to global location (or `git mv` if both are in the same repo's working tree, as `~/.claude/skills/` often is when the global config is a tracked dotfiles repo).
- `git rm -r` in the source repo. (Do **not** use `-rf` — most pre-commit hooks block that pattern.)
- Update CLAUDE.md skill tables: add a "cross-repo utility skills" note in the leaf, add row(s) in the brain's "Related skills" section.
- Two commits (one per repo).

Worked example from W-0064 audit: `rga` (cross-format full-text search over finance + xj-private-info + dropbox-inbox + private-config) and `singlefile-clipper` (reads from `~/Dropbox/dropbox-inbox/evidence/`) both lifted out of `xj-private-finance/.claude/skills/` into `~/.claude/skills/` because neither was finance-bound.

## Phase 4 (rare) — create planned skills

Only run this if Phases 1-3 surfaced legitimate gaps (planned-but-not-built rows in the CLAUDE.md table). Don't speculate. See `workflows.md § Create` for the per-skill workflow.

## Order matters — don't interleave

- Phase 1 cleans the source of truth (the CLAUDE.md table). Phase 2 and 3 will reference this updated table for their boundary decisions.
- Phase 2 refactors but doesn't move. Phase 3 moves but doesn't refactor. Keeping them separate makes each commit reviewable in isolation.

## Verification per phase

| Phase | Verify |
| --- | --- |
| 1 | `diff <(ls .claude/skills/) <(grep -oE '\`[a-z-]+\`' CLAUDE.md skill-table)` — sets match |
| 2 | `wc -l .claude/skills/<name>/SKILL.md` under threshold; same trigger list, same boundaries |
| 3 | global skill appears in the runtime's skill list; CLAUDE.md tables updated in both repos |

## Common pitfalls

- **Rewriting prose during a "refactor"**: Phase 2 is a relocation, not a rewrite. If you find yourself improving the prose mid-refactor, stop, commit the relocation as-is, then open a separate edit commit.
- **Deciding lift criteria mid-flight**: choose the lift criteria BEFORE Phase 3 starts, not while staring at a tempting candidate. Otherwise every borderline skill ends up lifted.
- **Skipping the doc commit**: Phase 1 looks trivial and gets skipped. Don't. The committed CLAUDE.md table is the artifact that informs the next two phases.
