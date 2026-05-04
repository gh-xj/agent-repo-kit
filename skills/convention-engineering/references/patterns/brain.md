# Pattern: Personal/Knowledge-Base Brain Repo

A git-versioned, mixed-authorship knowledge store: an owner writes high-signal
content, agents append captures from external sources, optional external
content sits in a library, and optional regenerable summaries live alongside.
Operated entirely by AI agents on the owner's behalf.

Codified from the `xj-private-brain` bootstrap (2026-05-03). Canonical
instance: `gh-xj/xj-private-brain`.

## When To Use

Apply this pattern when the repo will hold:

- Owner-written notes, journals, principles, or distilled learnings, AND
- Agent-captured raw material (chat exports, transcripts, ingested feeds), AND
- (Often) external authored content the owner has collected (books, articles,
  papers), AND
- (Often) machine-generated derivations over the above (indexes, summaries,
  graphs).

Skip when:

- The repo is just notes — use Obsidian, a wiki, or
  `references/operations/wiki.md` (lighter weight).
- The repo has source code as its primary artifact — use the core
  bootstrap workflow without this pattern.
- The owner won't actually use it. A brain repo is a forcing function for
  ingestion + review habits; it decays without those.

## Five Load-Bearing Decisions

These are non-negotiable at the _shape_ level. Names and exact gates are
substitutable; the structural choices are not.

### 1. Four realms, even if some start empty

The pattern mandates four realms. Each maps to a distinct write-permission
contract and gate set:

| Realm             | Owner writes? | Agent writes?                | Mutability                       | Typical name                        |
| ----------------- | ------------- | ---------------------------- | -------------------------------- | ----------------------------------- |
| Owner-authored    | yes           | no (without explicit ask)    | mutable                          | `human/`, `notes/`, `writing/`      |
| Agent-captured    | no            | append-only                  | **immutable after first commit** | `raw/`, `inbox/`, `captures/`       |
| External authored | yes           | yes (downloads, extractions) | mutable                          | `library/`, `sources/`, `external/` |
| Derived           | no            | yes                          | regenerable, safe to delete      | `derived/`, `built/`, `indexes/`    |

`library/` and `derived/` may start as a single README declaring the
contract — _the namespace reservation is the point_. Adding content later is
zero-cost; renaming a realm after content lands is a migration.

The owner-authored ↔ agent-captured split is the load-bearing choice. It is
what makes a brain repo not a code repo and not an Obsidian vault: agents
_append_ but never _overwrite_, and the owner _writes_ but never _manages
ingest plumbing by hand_.

### 2. Ingest sources are a registry, not enumerated

Inside the agent-captured realm, every external data source is a subdir with
its own README declaring:

- **Producer:** which CLI/tool/script writes here.
- **Filename pattern:** e.g. `YYYY-MM-DD_<id>.json`.
- **Schema:** frontmatter shape and/or JSON shape.
- **Cadence:** manual / cron / on-event.
- **Retention:** default `forever` (append-only).

The pattern does not enumerate sources. Each brain owner picks their own
(Gmail, calendar, transcripts, scrapers, dumps, etc.). The contract is the
registry shape; the membership is local.

### 3. Privacy is a posture, not a feature

A brain repo is private by default and private remote at most. The pattern
requires:

- **Secret scan** in the pre-commit hook (`gitleaks` or equivalent).
- **`.gitignore`** excludes operational state and any common credential
  paths the owner uses.
- **Optional but recommended:** banned-string scan (a small regex list of
  identifiers — names, employer slugs — that must never reach the remote)
  and a PII scrubber utility for any capture step that may catch terminal
  output containing credentials.

If the brain will not have a remote at all, the secret scan still applies —
clipboard leaks and shoulder-surfing are real.

**`.gitleaks.toml` allowlist for synthetic test values.** A PII scrubber
script will contain example bearer tokens, JWTs, fake credit cards, etc.
as test fixtures — and gitleaks will flag them on entropy alone (the
`generic-api-key` rule in particular). Ship a repo-root `.gitleaks.toml`
with a scoped allowlist for that script's path, not a wildcard:

```toml
[extend]
useDefault = true

[allowlist]
description = "Repo-scoped false-positive allowlist"
paths = [
  '''scripts/scrub-pii\.sh''',
]
```

This is the only sane way — globally relaxing the rule defeats the gate
everywhere; inline `# gitleaks:allow` comments inside the script add line
noise on every test case. The path-scoped allowlist is the smallest
hammer.

### 4. Temporal pointer is optional, but if present, canonical

If the brain has a daily/temporal shape (daily logs, journal entries):

- A `today.md` symlink at repo root points at the current day's file.
- A `templates/` directory holds the per-day (and per-week / per-month)
  templates.
- A verify gate asserts the symlink resolves to a real file under the
  owner-authored realm.

If the brain is issue-driven, project-driven, or topic-driven instead, this
section is skipped entirely. Don't fake a temporal shape to fit the pattern.

### 5. One canonical verify entry, gates composed per realm

`task verify` is the single entry point. The gate set is **composed from
declared realms and operations**, not copy-pasted:

| Realm/op declared    | Gate added                                                               |
| -------------------- | ------------------------------------------------------------------------ |
| Always               | Secret scan (gitleaks)                                                   |
| Always               | `CLAUDE.md` ↔ `AGENTS.md` mirror (if both declared)                      |
| Agent-captured realm | Immutability check (no file modified after first commit)                 |
| Agent-captured realm | Source-README presence (every subdir has a README)                       |
| Owner-authored realm | Schema check on owner-authored files (e.g. daily-log template structure) |
| Temporal pointer     | `today.md` symlink integrity                                             |
| `operations: [work]` | `work view ready --json` succeeds                                        |

Soft-pass during bootstrap: gates that have nothing to check yet exit 0
with an explanatory message ("no per-day logs yet — skipping schema check").
Don't make the bootstrap fail its own gates.

## `.conventions.yaml` Extensions

The brain pattern canonizes two keys that are otherwise "unknown but
allowed":

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

Realm _names_ are owner-chosen; realm _roles_ are fixed at four. An audit
agent checks the four roles are present, not that they're called `human/`
and `raw/`.

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
├── <captured-realm>/      # raw/
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

## Templates

### `.conventions.yaml` skeleton

```yaml
agent_docs:
  - CLAUDE.md
  - AGENTS.md
docs_root: docs
taskfile: true
pre_commit: true
skill_roots:
  - .claude/skills
operations:
  - work # or omit if not adopting
realms:
  owner:
    {
      path: human/,
      write: "owner only",
      rule: "agents may READ, may not write",
    }
  captured:
    {
      path: raw/,
      write: "agents append-only",
      rule: "immutable after first commit",
    }
  external:
    {
      path: library/,
      write: "owner + agents",
      rule: "prefer extracted md + SOURCES.md over binaries",
    }
  derived:
    {
      path: derived/,
      write: "agents only",
      rule: "regenerable; safe to delete",
    }
ingest_sources:
  registry: raw/README.md
  contract: |
    Each source under raw/<source>/ has its own README declaring
    producer, filename pattern, schema, cadence, retention.
checks:
  - "AGENTS.md mirrors CLAUDE.md (identical content)."
  - "task verify exits 0 from a clean checkout."
  - "No file under <captured-realm>/ is modified after its first commit."
  - "Every subdirectory under <captured-realm>/ has a README declaring its source contract."
  - "gitleaks protect --staged passes on every commit (enforced by pre-commit)."
```

### `CLAUDE.md` realms block

```markdown
## Realms — read carefully before writing

| Realm        | Path              | Who writes           | Rule                                                           |
| ------------ | ----------------- | -------------------- | -------------------------------------------------------------- |
| **owner**    | <owner-realm>/    | owner only           | Agents may **read**, must not write.                           |
| **captured** | <captured-realm>/ | agents (append-only) | Files are **immutable** after first commit.                    |
| **external** | <external-realm>/ | owner + agents       | Prefer extracted markdown + SOURCES.md pointers over binaries. |
| **derived**  | <derived-realm>/  | agents               | Regenerable. Safe to delete and rebuild.                       |
```

### `scripts/verify-raw-immutability.sh` (the load-bearing gate)

For each tracked file under the captured realm, compare HEAD content to the
blob from the first commit that introduced the file. README.md files are
exempt (schema documentation evolves).

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
git rev-parse HEAD >/dev/null 2>&1 || { echo "no commits — skip"; exit 0; }
mapfile -t files < <(git ls-files <captured-realm>/ 2>/dev/null || true)
[ "${#files[@]}" -eq 0 ] && { echo "<captured-realm>/ empty — skip"; exit 0; }
failed=0
for f in "${files[@]}"; do
  [[ "$f" == */README.md ]] && continue
  first=$(git log --diff-filter=A --follow --format=%H -- "$f" | tail -1)
  [ -z "$first" ] && continue
  first_blob=$(git rev-parse "${first}:${f}" 2>/dev/null || echo "")
  head_blob=$(git rev-parse "HEAD:${f}" 2>/dev/null || echo "")
  [ -z "$first_blob" ] || [ -z "$head_blob" ] && continue
  [ "$first_blob" != "$head_blob" ] && { echo "FAIL: $f modified" >&2; failed=1; }
done
[ "$failed" -ne 0 ] && exit 1
echo "OK: ${#files[@]} captured file(s) unchanged"
```

The other verify scripts (`verify-today-symlink.sh`,
`verify-source-readmes.sh`, `verify-schema.sh`) follow the same shape:
soft-pass when the realm is empty, hard-fail on violations otherwise.

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

## Execution Steps

1. **Pre-flight.** Confirm `git`, `task`, `gitleaks` are installed. Inspect
   any directories the user wants to absorb — _especially_ check for
   nested `.git` directories (other people's repos that should NOT be
   absorbed; see Anti-Patterns).
2. **Write `.conventions.yaml`** at the repo root using the skeleton.
   Realm names are the owner's choice; commit them deliberately.
3. **Scaffold the four realm directories** with a README each. `library/`
   and `derived/` get one-line READMEs declaring the contract. Owner and
   captured realms get fuller READMEs.
4. **Write `CLAUDE.md`** with the realms block, hard rules, layout overview,
   and pointers. Mirror to `AGENTS.md` if both are declared in
   `agent_docs:`.
5. **Write `Taskfile.yml`** with `task verify` as the canonical entry,
   `verify:secrets`, `verify:agent-docs-mirror`, and one subtask per
   declared realm/operation gate.
6. **Write `scripts/verify-*.sh`** — at minimum
   `verify-raw-immutability.sh`, `verify-source-readmes.sh`, and (if
   temporal) `verify-today-symlink.sh`. Make them soft-pass during
   bootstrap.
7. **Wire `.githooks/pre-commit`** with the cheap subset of verify gates
   (always: secrets + mirror + immutability + source-READMEs). Enable via
   `git config core.hooksPath .githooks`.
8. **Adopt operations as needed** — `.work/` for ingest/triage tracking,
   `.wiki/` for cross-cutting reference. Each adoption follows its own
   reference doc and adds its own gate.
9. **Migrate or seed initial content.** For _legacy_ personal-content
   directories (old journals, accumulated notes), prefer **preserve as-is
   in a `legacy-*/` subdirectory** over lossy schema conversion. The
   pattern's gates exempt `legacy-*/` if declared.
10. **Smoke-test.** `task verify` must exit 0. Run `git status` and check
    for unintended additions; commit the scaffold; only then push to a
    private remote (with explicit user confirmation — pushing is a
    "visible to others" action).

## Migration Safety (Live-Data Migrations)

When migrating live personal data into the brain (an old-format archive,
a sibling repo's content, scattered notes from another tool), the
migration script MUST:

1. **Default to `--dry-run`.** Never mutate the destination on bare
   invocation. Real migrations require an explicit `--apply` flag.
2. **Write dry-run output to a STAGING directory** under
   `.work/spaces/W-NNNN/migration-output/` (gitignored). Never to the real
   `human/` destination in dry-run mode. The owner spot-checks staging
   before committing.
3. **Treat the source directory as READ-ONLY.** No writes ever, not even
   to the source's own `.git/`.
4. **Log every per-file decision** (kept, dropped, anomaly) to a
   `MIGRATION_LOG.md` in the staging dir.
5. **Be re-runnable.** `--apply` should be idempotent — safe to re-run if
   the first attempt was interrupted mid-batch.

Pair this with the **two-phase shipping pattern**:

- **Phase 1:** commit the script + template + dry-run sample as safe
  artifacts. Owner reviews the staging output.
- **Phase 2:** owner says "apply"; you run `--apply` and commit the data
  as a separate commit (typically much larger than Phase 1).

Two phases keep the high-risk operation reversible until the last
moment, and produce a clean git history with the artifacts and the
data-mass separated. Proven on the canonical instance's xj_core daily-
log migration (69 files, zero owner regret).

## Gotchas Observed In Practice

- **Formatter hooks touching writes.** Some user setups have a
  PostToolUse hook that reformats markdown after every Write. This makes
  `CLAUDE.md` ↔ `AGENTS.md` mirroring drift silently between writes.
  Workaround: re-mirror immediately before commit; long-term, wire the
  formatter explicitly into `task fmt` so it's intentional.
- **External directories with nested `.git`.** "Absorb my downloads
  folder" almost always means absorbing other people's repos. Audit
  `find <target> -name .git -type d` first. The right move is usually a
  `library/<name>/SOURCES.md` pointer, not absorption.
- **Large binaries.** Brains accumulate PDFs, EPUBs, audio. Defer Git LFS
  until size pressure is real; document the LFS extension list in
  `.gitattributes` so flipping the switch later is a single command.
- **Daily-log format mismatch on legacy migration.** Old journals are
  often monthly files with day headers; the new format is per-day files.
  Preserve old-format under `<owner-realm>/<daily>/legacy-monthly/` (or
  similar); start per-day going forward. The schema gate exempts
  `legacy-*/`.
- **Soft-pass gates that never become hard-pass.** A gate that always
  prints "skipping (bootstrap state)" is dead. Re-audit after the first
  real content lands and tighten.

## Anti-Patterns

- **Absorbing other people's git repos** to grow the library quickly.
  Either pollutes the brain's history (if you copy contents) or creates
  submodule chaos (if you add as submodules). Use `SOURCES.md` pointers.
- **Realm policy fork.** Don't run mutating ops (formatters, linters)
  across realms uniformly. Owner-realm wants prose-formatting rules;
  captured-realm wants to be touched as little as possible; derived/ may
  not need formatting at all.
- **Schedulers in the descriptor.** Cron/heartbeat doctrine belongs in a
  per-archetype prose doc (e.g. a `HEARTBEAT.md` per the brain pattern's
  conventions), not as a `.conventions.yaml` schema field. Schedules
  evolve faster than the descriptor should.
- **Conflating capture with curation.** Captured-realm READMEs declaring
  "I'll review this someday" without a triage gate become a junk drawer.
  Pair `raw/` with `.work/` for triage discipline.
- **Promoting `derived/` content to source-of-truth.** If you find
  yourself hand-editing under `derived/`, that file belongs in
  owner-realm. Keep the regenerable-vs-authored boundary clean.

## Worked Example

`gh-xj/xj-private-brain` (initial commits 2026-05-03) is the canonical
instance. Its realm naming choices: `human/`, `raw/`, `library/`,
`derived/`. Its operations: `[work]`. Its verify gates: secrets, daily
schema, today-symlink, raw-immutability, agent-docs-mirror,
raw-source-readmes, work-check. Diff against this doc when the pattern
needs refinement.
