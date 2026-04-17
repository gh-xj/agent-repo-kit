# Docs Taxonomy

Use intent-first folders so every document has one obvious home.

## 1) Pick One Root Per Repo

Choose exactly one active docs root:

- Tracked mode: `docs/`
- Local-overlay mode (git-exclude): `.docs/`

Do not split active docs across both roots in the same repo unless there is an explicit migration plan.

## 2) Canonical Tree

Apply this tree under the chosen root (`docs` or `.docs`):

```text
<root>/
├── requests/        # RFI / asks / scope + acceptance
├── planning/        # design decisions and architecture choices
├── plans/           # executable implementation plans (task-by-task)
├── implementation/  # execution notes, rollout logs, post-implementation reports
├── taxonomy/        # domain data taxonomies (schemas, mapping, query semantics)
└── reviews/         # optional review artifacts from humans/subagents
```

## 3) Filename Contracts

Use stable date-prefixed names:

- `requests/YYYYMMDD_rfi_<topic>.md`
- `planning/YYYY-MM-DD_<topic>_design.md`
- `plans/YYYY-MM-DD-<topic>.md`
- `implementation/YYYY-MM-DD_<topic>_impl_report.md`
- `taxonomy/<domain>/<subject>.md`
- `reviews/<artifact_basename>.<reviewer>.md`

Keep topic slugs short and lowercase with `_` or `-`.

## 4) Lifecycle Contract

For feature work:

1. Start with `requests/` (why, scope, acceptance).
2. Record approved architecture in `planning/`.
3. Generate execution steps in `plans/`.
4. Capture results in `implementation/`.

Recommended handoff bundle for implementation owner:

1. RFI (`requests/`)
2. Design (`planning/`)
3. Plan (`plans/`)

## 4b) Additional Artifact Categories

The lifecycle folders above cover feature work. A mature repo also surfaces
the following categories. Each is an **invariant** (present when the category
applies); shape and cadence are the author's call.

- **Active vs completed plans.** Split in-flight from shipped. Tracked repos
  may use `plans/active/` + `plans/completed/` or any equivalent split. The
  invariant is that stale plans do not shadow in-flight ones.

- **Tech-debt tracker (flat file).** A single checked-in file (e.g.
  `plans/tech-debt-tracker.md`) enumerates known debt items. Use this for
  many small related items instead of one ticket per line.

- **External references (`references/`).** LLM-readable external docs —
  vendor docs, API specs, fetched articles — named by topic. The llms.txt
  convention (`<topic>-llms.txt`) is the recommended format when available.
  These are distinct from `.wiki/raw/` which stores source material the repo
  summarises and cites.

- **Generated artifacts (`generated/`).** Auto-generated from code or
  tooling — schema dumps, API surface snapshots, dependency graphs. The
  invariant: marked clearly as non-editable-by-hand, regeneration command
  documented, reviewed as part of the generating PR.

- **Topical dashboards (root markdown files).** Cross-cutting syntheses
  discoverable by name rather than ID — e.g. `ARCHITECTURE.md`, `DESIGN.md`,
  `SECURITY.md`, `RELIABILITY.md`, `QUALITY_SCORE.md`. Each is a living
  overview of its domain. The invariant: exists iff the repo has enough
  cross-cutting work in that domain to justify a synthesis layer; owned by
  someone (human or recurring agent).

- **Per-subtree index.** When a subtree has more than a handful of files,
  an `index.md` at its root lists contents with one-line summaries. May be
  generated or hand-maintained; the invariant is that it stays in sync with
  the files it lists (mechanical check if possible).

## 5) Minimum Audit Checks

At minimum, verify:

- docs root exists (`docs/` or `.docs/`)
- `requests/`, `planning/`, `plans/` exist
- new feature work has all three artifacts before implementation starts

Example contract fragments:

Tracked mode:

```json
{
  "contract_version": 1,
  "mode": "tracked",
  "profiles": ["go"],
  "docs_root": "docs",
  "ownership_policy": {
    "portable_skill_authoring_owner": "skill-author",
    "domain_knowledge_owner": "domain-skills",
    "repo_local_skills": {
      "allowed": false,
      "placement_roots": [".claude/skills", ".agents/skills"],
      "authoring_owner": "skill-author",
      "requires_justification": true
    }
  },
  "required_files": ["docs/requests", "docs/planning", "docs/plans"]
}
```

Local-overlay mode:

```json
{
  "contract_version": 1,
  "mode": "overlay",
  "profiles": ["go"],
  "docs_root": ".docs",
  "ownership_policy": {
    "portable_skill_authoring_owner": "skill-author",
    "domain_knowledge_owner": "domain-skills",
    "repo_local_skills": {
      "allowed": false,
      "placement_roots": [".claude/skills", ".agents/skills"],
      "authoring_owner": "skill-author",
      "requires_justification": true
    }
  },
  "required_files": [".docs/requests", ".docs/planning", ".docs/plans"]
}
```
