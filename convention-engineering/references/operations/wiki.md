# Operation: Wiki

Minimal LLM-maintained knowledge base. Two page types, one lint script,
no schema-surgery or repo-root patching.

## Adopt this convention in 3 steps

1. **Copy the template and initialize.**

   ```bash
   cp -R ~/.claude/skills/convention-engineering/references/templates/wiki \
         <repo>/.wiki
   cd <repo>
   task -d .wiki init
   ```

2. **Wire discoverability into `CLAUDE.md` and `AGENTS.md`.**
   Add (or merge into an existing `## Conventions` section):

   ```markdown
   - **Wiki** — LLM-maintained knowledge base at `.wiki/`. Read `.wiki/RULES.md`
     for page types, frontmatter, and citation rules. Validate with
     `task wiki:lint` (or `task -d .wiki lint`).
   ```

3. **Optionally wire up the root Taskfile.**
   To invoke `task wiki:lint` from repo root, add to your root `Taskfile.yml`:
   ```yaml
   includes:
     wiki:
       taskfile: .wiki/Taskfile.yml
       dir: .wiki
   ```
   Then commit `.wiki/`. Raw sources under `.wiki/raw/` are immutable once committed.

Manual copy-paste, by design — replaces the 1358-line bootstrap that used to
do `awk`-based text surgery on CLAUDE.md / AGENTS.md / Taskfile.yml.

## When to adopt

Adopt `.wiki/` ONLY if the repo actually ingests external sources you want to
cite. If you don't have files to drop in `raw/`, you don't need a wiki.

Adoption test — the wiki earns its keep when ALL three are true:

1. You have external material (papers, vendor docs, transcripts, PDFs) to track.
2. You want summaries pinned to those sources via `raw_path` + `[[S-NNN]]`.
3. You want broken-citation lint to catch drift.

If any of those is "no," use `docs/` instead. The wiki's surface area (IDs,
lint, raw/pages split) is a tax you should only pay when enforced provenance
is the point.

### Wiki vs `docs/` — decision rule

- **Cites `[[S-NNN]]` or lives in `raw/`** → `.wiki/`
- **Everything else** (setup guides, incident cards, decisions, plans,
  human-requests) → `docs/`

A "note" with zero `[[S-NNN]]` citations is a `docs/` file wearing a wiki
costume. Don't file it in `pages/`.

## Verb surface

| Verb   | Usage                |
| ------ | -------------------- |
| `init` | `task -d .wiki init` |
| `lint` | `task -d .wiki lint` |

No `ingest` or `query` verbs. Those are LLM operations, not shell commands.
You perform them by asking the LLM directly: "ingest `raw/<file>`" or
"what does the wiki say about X?"

## Two page types

| Type   | Prefix  | Purpose                                                 |
| ------ | ------- | ------------------------------------------------------- |
| Source | `S-NNN` | Summary + key claims from one immutable raw document    |
| Note   | `N-NNN` | Concept, synthesis, decision, or question — LLM-derived |

The old taxonomy (S/E/T/X/Q) is collapsed. Distinguish notes by body content
and tags, not by elaborate schema.

## Invariants enforced by lint

- Every page has a frontmatter `id:` that matches its filename prefix
- IDs are unique per type prefix
- Every `[[S-NNN]]` or `[[N-NNN]]` wikilink resolves to an existing page
- Every source page's `raw_path:` points to an existing file under `raw/`

**Not** enforced by lint (by design):

- Source-citation coverage ("every claim cites `[[S-NNN]]`") — too noisy, LLM reads pages anyway
- `links:` frontmatter / body parity — dropped the `links:` field entirely
- Log append-only discipline — dropped `log.md`; git log is the audit trail

## Directory layout

```
.wiki/
├── Taskfile.yml        # ~25 lines
├── RULES.md            # terse page-type + rule reference
├── raw/                # immutable sources, committed
│   └── <name>.ext
├── pages/              # LLM-owned markdown pages
│   ├── S-001-…md
│   └── N-001-…md
└── scripts/lint.sh     # <75 lines
```

Total template surface: ~170 lines across 3 files.

## What this replaces

The old wiki skill shipped ~2,500 lines across a 1,358-line bootstrap, five
JSON schemas, an `index.md` + `log.md` the LLM had to maintain by hand, a
302-line linter that did not enforce the two most important claimed invariants,
a project-local generated skill, and repo-root file patching. Codex review
verdict: "massively over-engineered — rebuild with ~10% of current code."

## Template location

`~/.claude/skills/convention-engineering/references/templates/wiki/`.
