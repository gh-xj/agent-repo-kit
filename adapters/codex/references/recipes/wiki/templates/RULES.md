# Wiki Rules

LLM-maintained knowledge base. Human drops raw sources; LLM builds pages.

## When to file here vs `docs/`

**File in `.wiki/` ONLY if the content cites an external source that lives in
`.wiki/raw/`.** Everything else goes in `docs/`.

Decision rule:

- **`.wiki/raw/<name>.ext`** — verbatim external material you didn't author
  (papers, vendor docs, transcripts, screenshots). Immutable.
- **`.wiki/pages/S-NNN-*.md`** — your summary of one file in `raw/`, pinned via
  `raw_path`. Every factual claim in later notes cites `[[S-NNN]]`.
- **`.wiki/pages/N-NNN-*.md`** — synthesis across sources, with `[[S-NNN]]`
  citations in the body.
- **`docs/*.md`** — everything else: setup guides, incident cards, decisions,
  human-requests, freeform notes. Hand-named, no ID ceremony, no lint.

If a note has zero `[[S-NNN]]` citations and no raw counterpart, it belongs in
`docs/`, not `pages/`. The wiki's value is enforced provenance; notes without
provenance are just docs with extra ceremony.

## Two layers

- `raw/` — immutable source documents (LLM reads only, never writes)
- `pages/` — LLM-written markdown pages with frontmatter

## Two page types

| Type   | Prefix  | Purpose                                        |
| ------ | ------- | ---------------------------------------------- |
| Source | `S-NNN` | Summary + key claims from one raw doc          |
| Note   | `N-NNN` | Concept, synthesis, decision, or open question |

Combine previously-distinct E/T/X/Q types into notes — distinguish by tags and
body content, not by schema.

## Frontmatter

```yaml
---
id: S-001 # or N-001
type: source # or note
title: Attention Is All You Need
created: 2026-04-13
updated: 2026-04-13
tags: [transformer, attention]
---
```

Source pages **must** record `raw_path`, which **must** start with `raw/` and
point to an existing file. They may add `author`, `year`, `tier`.
Note pages may add `status`, `confidence`, `sources: [S-001, S-003]`.

## Rules

1. **Never edit `raw/`.** Source documents are immutable.
2. **Cite every factual claim.** Use `[[S-NNN]]` wikilinks in the body.
3. **IDs are monotonic and never reused.** Per-type sequence (S-001, S-002; N-001, N-002).
4. **Filename matches ID:** `{id}-{slug}.md`, slug is kebab-case from title.
5. **Wikilinks use paths, not IDs alone:** `[[S-001-attention-is-all-you-need]]` or
   equivalently the lint resolves `[[S-001]]` by prefix.
6. **No hand-maintained index.** List with `task wiki:lint` or `ls pages/`.

## Operations

### Ingest

Human puts a source at `raw/<name>.ext`. Ask LLM to ingest:

1. Read the raw file.
2. Create `pages/S-NNN-slug.md` with summary and key claims.
3. If notable concepts deserve standalone notes, create `pages/N-NNN-slug.md`
   and cite the source via `[[S-NNN]]`.

### Query

Ask the LLM a question. It reads `pages/` (and follows wikilinks) to answer
with `[[S-NNN]]` citations. Reusable answers MAY be filed as a new `N-NNN` note.

### Lint

```bash
task -d .wiki lint
```

Checks: unique IDs, filename matches `id`, every `[[S-NNN]]` resolves to an
existing source page, every source page has a `raw_path` starting with `raw/`
that points to an existing file.
