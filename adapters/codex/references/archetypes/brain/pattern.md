# Archetype: Personal/Knowledge-Base Brain Repo

Status: stable
Codified: 2026-05-03
Last revised: 2026-05-05 (4-realm → 3-realm refactor)

A git-versioned, mixed-authorship knowledge store: an owner writes high-signal content, agents append captures from external sources, optional external content sits in a library, and optional regenerable summaries live alongside. Operated entirely by AI agents on the owner's behalf.

When applying this archetype, the canonical instance is whatever brain repo you're authoring or maintaining — substitute your own repo path in the per-instance references below.

## When To Use

Apply when the repo will hold:

- Owner-written notes, journals, principles, or distilled learnings, AND
- Agent-captured raw material (chat exports, transcripts, ingested feeds), AND
- (Often) external authored content the owner has collected (books, articles, papers), AND
- (Often) machine-generated derivations over the above (indexes, summaries, graphs).

Skip when:

- The repo is just notes — use Obsidian, a wiki, or [`recipes/wiki/pattern.md`](../../recipes/wiki/pattern.md) (lighter weight).
- The repo has source code as its primary artifact — use [`meta/bootstrap-workflow.md`](../../meta/bootstrap-workflow.md) without this archetype.
- The owner won't actually use it. A brain repo is a forcing function for ingestion + review habits; it decays without those.

## Five Load-Bearing Decisions

These are non-negotiable at the *shape* level. Names and exact gates are substitutable; the structural choices are not.

### 1. Three core realms (and an optional fourth)

The archetype mandates three core realms. A fourth (agent-captured) is optional — adopt it only if your workflow actually accumulates raw exports that benefit from append-only preservation. Each realm maps to a distinct write-permission contract and gate set:

| Realm             | Required? | Owner writes? | Agent writes?                | Mutability                       | Typical name                        |
| ----------------- | --------- | ------------- | ---------------------------- | -------------------------------- | ----------------------------------- |
| Owner-authored    | yes       | yes           | no (without explicit ask)    | mutable                          | `human/`, `notes/`, `writing/`      |
| External authored | yes       | yes           | yes (downloads, extractions) | mutable                          | `library/`, `sources/`, `external/` |
| Derived           | yes       | no            | yes                          | regenerable, safe to delete      | `derived/`, `built/`, `indexes/`    |
| Agent-captured    | optional  | no            | append-only                  | **immutable after first commit** | `raw/`, `inbox/`, `captures/`       |

`library/` and `derived/` may start as a single README declaring the contract — *the namespace reservation is the point*. Adding content later is zero-cost; renaming a realm after content lands is a migration.

**When to adopt the captured realm vs skip it.** Adopt it when an automated pipeline genuinely lands raw exports faster than the owner can pre-process them, or when the owner wants raw transcripts preserved verbatim for later re-derivation. Skip it when the owner pre-processes every incoming artifact before it lands. The gates (`raw-immutability`, `raw-source-readmes`) only earn their cost when the realm is actually populated.

When skipped, the owner-authored ↔ derived/regenerable split carries the load: agents may write derivations but only the owner edits `human/`. "Agent-immutable raw" stops being a valid third option.

*Decided 2026-05-05 in the canonical instance:* captured realm dropped. The owner's "always process before load" workflow never populated the realm; the gates were policing an empty dir. This is the worked example for "skip the captured realm" — when nothing lands raw, the realm and its gates earn nothing.

### 2. Ingest sources are a registry, not enumerated *(applies only when the captured realm is adopted)*

Inside the agent-captured realm, every external data source is a subdir with its own README declaring:

- **Producer:** which CLI/tool/script writes here.
- **Filename pattern:** e.g. `YYYY-MM-DD_<id>.json`.
- **Schema:** frontmatter shape and/or JSON shape.
- **Cadence:** manual / cron / on-event.
- **Retention:** default `forever` (append-only).

The archetype does not enumerate sources. Each brain owner picks their own (Gmail, calendar, transcripts, scrapers, dumps, etc.). The contract is the registry shape; the membership is local.

If the captured realm is skipped (decision #1), this entire section is inapplicable — pre-processed content lands directly in `human/` (curated) or `derived/` (regenerable) instead of going through a per-source registry.

### 3. Privacy is a posture, not a feature

A brain repo is private by default and private remote at most. The archetype requires:

- **Secret scan** in the pre-commit hook (`gitleaks` or equivalent).
- **`.gitignore`** excludes operational state and any common credential paths the owner uses.
- **Optional but recommended:** banned-string scan (a small regex list of identifiers — names, employer slugs — that must never reach the remote) and a PII scrubber utility for any capture step that may catch terminal output containing credentials.

If the brain will not have a remote at all, the secret scan still applies — clipboard leaks and shoulder-surfing are real.

A PII scrubber script will contain example bearer tokens, JWTs, fake credit cards, etc. as test fixtures — and gitleaks will flag them on entropy alone. Ship a repo-root `.gitleaks.toml` with a path-scoped allowlist for that script, not a wildcard:

```toml
[extend]
useDefault = true

[allowlist]
description = "Repo-scoped false-positive allowlist"
paths = [
  '''scripts/scrub-pii\.sh''',
]
```

The path-scoped allowlist is the smallest hammer — globally relaxing the rule defeats the gate everywhere; inline `# gitleaks:allow` comments on every test case add line noise.

### 4. Temporal pointer is optional, but if present, canonical

If the brain has a daily/temporal shape (daily logs, journal entries):

- A `today.md` symlink at repo root points at the current day's file.
- A `templates/` directory holds the per-day (and per-week / per-month) templates.
- A verify gate asserts the symlink resolves to a real file under the owner-authored realm.

If the brain is issue-driven, project-driven, or topic-driven instead, this section is skipped entirely. Don't fake a temporal shape to fit the archetype.

### 5. One canonical verify entry, gates composed per realm

`task verify` is the single entry point. The gate set is **composed from declared realms and operations**, not copy-pasted:

| Realm/op declared    | Gate added                                                               |
| -------------------- | ------------------------------------------------------------------------ |
| Always               | Secret scan (gitleaks)                                                   |
| Always               | `CLAUDE.md` ↔ `AGENTS.md` mirror (if both declared)                      |
| Agent-captured realm | Immutability check (no file modified after first commit)                 |
| Agent-captured realm | Source-README presence (every subdir has a README)                       |
| Owner-authored realm | Schema check on owner-authored files (e.g. daily-log template structure) |
| Temporal pointer     | `today.md` symlink integrity                                             |
| `operations: [work]` | `work view ready --json` succeeds                                        |

Soft-pass during bootstrap: gates that have nothing to check yet exit 0 with an explanatory message ("no per-day logs yet — skipping schema check"). Don't make the bootstrap fail its own gates.

## File Inventory

Minimal viable bootstrap. Substitute realm names per owner choice.

```
<brain>/
├── .conventions.yaml      # declares realms, operations, checks
├── .gitignore             # excludes /.docs, /.work, secrets, OS noise
├── .gitattributes         # text/binary normalization; LFS prep
├── .githooks/pre-commit   # gitleaks + cheap subset of verify gates
├── CLAUDE.md              # agent contract — realms + hard rules
├── AGENTS.md              # mirrored if both declared
├── Taskfile.yml           # canonical task verify entry + per-gate subtasks
├── <owner-realm>/         # human/ (or chosen name)
│  └── README.md
├── <captured-realm>/      # raw/ (only if adopted)
│  └── README.md           # registry contract
├── <external-realm>/      # library/
│  └── README.md
├── <derived-realm>/       # derived/
│  └── README.md
├── templates/             # only if temporal pointer in use
├── today.md               # symlink, only if temporal pointer in use
├── docs/                  # convention docs (requests/planning/plans/implementation)
├── scripts/               # verify-*.sh per gate
├── .claude/               # repo-local settings + skills (if Claude Code drives this)
└── .work/                 # if operations: [work] adopted (gitignored)
```

## `.conventions.yaml` Extensions

The brain archetype canonizes two keys that are otherwise "unknown but allowed":

```yaml
realms:
  owner: # required
    path: human/ # owner's choice of name
    write: "owner only"
    rule: "agents may READ, may not write"
  captured: # required
    path: raw/
    write: "agents append-only"
    rule: "immutable after first commit"
  external: # required (may start as a single README)
    path: library/
    write: "owner + agents (downloads, extractions)"
    rule: "prefer extracted markdown + SOURCES.md pointers over large binaries"
  derived: # required (may start as a single README)
    path: derived/
    write: "agents only"
    rule: "regenerable; safe to delete and rebuild"

ingest_sources:
  registry: <captured-realm-path>/README.md
  contract: |
    Each source under <captured-realm>/<source>/ has its own README
    declaring producer, filename pattern, schema, cadence, retention.
```

Realm *names* are owner-chosen; realm *roles* are fixed at four. An audit agent checks the four roles are present, not that they're called `human/` and `raw/`.

## Templates

Concrete scaffold files in [`templates/`](./templates/):

- `conventions.yaml.tmpl` — descriptor with realm and ingest blocks
- `CLAUDE.md.tmpl` — agent contract with realms block
- `verify-raw-immutability.sh` — load-bearing gate for the captured realm

## Per-Realm Gate Matrix

| Gate                                | Owner         | Captured | External           | Derived        | Always |
| ----------------------------------- | ------------- | -------- | ------------------ | -------------- | ------ |
| Secret scan                         | —             | —        | —                  | —              | ✅     |
| Doc mirror (CLAUDE/AGENTS)          | —             | —        | —                  | —              | ✅     |
| Schema check (template conformance) | ✅            | —        | —                  | —              | —      |
| Immutability check                  | —             | ✅       | —                  | —              | —      |
| Source-README presence              | —             | ✅       | —                  | —              | —      |
| `today.md` symlink integrity        | (if temporal) | —        | —                  | —              | —      |
| `library/SOURCES.md` presence       | —             | —        | (if pointers used) | —              | —      |
| `derived/` regenerability smoke     | —             | —        | —                  | (if non-empty) | —      |

## Bootstrap

See [`bootstrap.md`](./bootstrap.md) for the step-by-step procedure.

## Migration

See [`migration.md`](./migration.md) for live-data migration safety (when absorbing legacy notes or other people's content into the brain).

## Worked Example

The canonical instance (initial commits 2026-05-03; raw-realm-dropped refactor 2026-05-05) ended up with:

- **Realm naming:** `human/`, `library/`, `derived/` — the 3-realm variant per decision #1's "skip the captured realm" branch.
- **Operations:** `[work]` only. No `[wiki]`.
- **Verify gates (8):** secrets, banned-strings, scrub-pii (self-test), daily-schema, today-symlink, agent-docs-mirror, library-sources, work-check.

Substitute your own instance's choices when adopting the archetype. Diff against this doc when the archetype itself needs refinement; the instance and the archetype can intentionally diverge as long as the divergence is named in the instance's contract files.

## Anti-Patterns

- **Absorbing other people's git repos** to grow the library quickly. Either pollutes the brain's history (if you copy contents) or creates submodule chaos (if you add as submodules). Use `SOURCES.md` pointers.
- **Realm policy fork.** Don't run mutating ops (formatters, linters) across realms uniformly. Owner-realm wants prose-formatting rules; captured-realm wants to be touched as little as possible; derived/ may not need formatting at all.
- **Schedulers in the descriptor.** Cron/heartbeat doctrine belongs in a per-archetype prose doc (e.g. a `HEARTBEAT.md`), not as a `.conventions.yaml` schema field. Schedules evolve faster than the descriptor should.
- **Conflating capture with curation.** Captured-realm READMEs declaring "I'll review this someday" without a triage gate become a junk drawer. Pair `raw/` with `.work/` for triage discipline.
- **Promoting `derived/` content to source-of-truth.** If you find yourself hand-editing under `derived/`, that file belongs in owner-realm. Keep the regenerable-vs-authored boundary clean.

## Gotchas

- **Formatter hooks touching writes.** Some user setups have a PostToolUse hook that reformats markdown after every Write. This makes `CLAUDE.md` ↔ `AGENTS.md` mirroring drift silently between writes. Workaround: re-mirror immediately before commit.
- **External directories with nested `.git`.** "Absorb my downloads folder" almost always means absorbing other people's repos. Audit `find <target> -name .git -type d` first.
- **Soft-pass gates that never become hard-pass.** A gate that always prints "skipping (bootstrap state)" is dead. Re-audit after the first real content lands and tighten.
