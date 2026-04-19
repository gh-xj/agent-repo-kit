# Research Corpus Convention Profile

Conventions for structured research repositories that use a lifecycle-layer architecture (raw → reviews → gaps → sources → synthesis).

Composes with other profiles: a research repo with Go CLI tooling uses both `profiles/research-corpus.md` and `profiles/go.md`.

## Table of Contents

1. [Domain Structure](#1-domain-structure) — placement modes, lifecycle dirs, naming, README
2. [Raw Capture Contract](#2-raw-capture-contract) — 8-line header, capture methods, size limits
3. [Frontmatter Contract](#3-frontmatter-contract) — required fields, status taxonomy
4. [Semantic Gate Patterns](#4-semantic-gate-patterns) — gap schema, review hygiene, citations
5. [Run Metadata](#5-run-metadata) — location, required fields, activity logging
6. [INDEX Freshness](#6-index-freshness) — staleness detection, integration
7. [Verification Gate Template](#7-verification-gate-template) — Taskfile tasks, convention config

## 1. Domain Structure

### Placement Modes

Research can live at two locations depending on the repo's primary purpose:

| Mode          | Root Path        | When to Use                                                    |
| ------------- | ---------------- | -------------------------------------------------------------- |
| **Repo-root** | `.`              | Dedicated research repos where research IS the product         |
| **Nested**    | `docs/research/` | Code-first repos where research supports engineering decisions |

All paths below are relative to the chosen root. In nested mode, `NN-domain-name/` becomes `docs/research/NN-domain-name/`, shared dirs become `docs/research/_synthesis-cross/`, etc.

The convention config should declare the root:

```json
{
  "research_root": "."
}
```

or:

```json
{
  "research_root": "docs/research"
}
```

### Product-Group Tier (optional)

When a research repo covers multiple products or concern areas, domains can be nested under **product group directories** for a two-tier hierarchy:

```
{research_root}/
  group-a/           # Product group (e.g., a fintech repo with `payments`, `risk`, `ledger`)
    NN-domain/       # Shortened domain name
  group-b/
    NN-domain/
  _synthesis-cross/  # Stays at research root (cross-group)
  _meta/             # Stays at research root
```

| Decision              | Flat                           | Product-group tier                         |
| --------------------- | ------------------------------ | ------------------------------------------ |
| Domain path           | `NN-domain-name/`              | `group/NN-short-name/`                     |
| Glob pattern          | `[0-9][0-9]-*/`                | `*/[0-9][0-9]-*/` (or explicit group list) |
| Frontmatter `domain:` | `NN-domain-name`               | `group/NN-short-name`                      |
| Cross-references      | `NN-domain-name/synthesis/...` | `group/NN-short-name/synthesis/...`        |

When using product-group tier, declare the groups in the convention config:

```json
{
  "research_root": ".",
  "product_groups": ["payments", "risk", "ledger"]
}
```

Scripts and tooling must iterate groups explicitly:

```bash
for group in payments risk ledger; do
  for domain in "$group"/[0-9][0-9]-*/; do
    [ -d "$domain" ] || continue
    # ... per-domain checks
  done
done
```

### Required Directories per Domain

Each domain is a numbered directory (either at research root or under a product group):

```
{research_root}/
  [group/]NN-domain-name/
    README.md
    raw/
      materials/     # Downloaded source content (papers, specs, docs, repos, blogs)
      manifests/     # URL-to-local-path mapping indexes
      notes/         # Agent analysis notes (observations, comparison notes)
    reviews/         # Source quality critiques
    gaps/            # Actionable research tasks
    sources/         # Verified primary evidence packs
    synthesis/       # Domain-level synthesis documents
```

### Naming Convention

- Domains: `NN-kebab-name/` where NN is zero-padded number (01, 02, ...)
- Product groups (if used): lowercase kebab-case (e.g., `payments/`, `risk/`)
- Shared directories under research root: `_synthesis-cross/`, `_meta/`, `_archive/`
- Files: `kebab-case-vN.md` with version suffix

### Domain README

Each domain README must include:

| Section              | Content                                    |
| -------------------- | ------------------------------------------ |
| Scope                | What the domain covers                     |
| Maturity             | Active / Stale / Archived                  |
| Layer Status         | Table with file counts per lifecycle layer |
| Open Gaps            | Reference to gaps/ files                   |
| Next-Wave Priorities | Numbered priority list                     |

## 2. Raw Capture Contract

### 8-Line Header (required for raw/materials/ and raw/notes/)

```
Source URL: [actual URL or "n/a" for local analysis]
Captured at: YYYY-MM-DD
Domain: NN-domain-name
Material type: paper|spec|doc|repo-analysis|blog|talk|issue|notes
Tier: A|B|C|n/a
Publisher / project: [name]
Capture method: web-fetch-full|web-fetch-failed|n/a
Why relevant: [one sentence]
```

After header: blank line, `---`, blank line, then content.

### Capture Method Values

| Value               | Meaning                                             | Rule                        |
| ------------------- | --------------------------------------------------- | --------------------------- |
| `web-fetch-full`    | Actual source content downloaded                    | Required for materials      |
| `web-fetch-failed`  | Fetch failed; summary as fallback with `Note:` line | Acceptable with explanation |
| `web-fetch-summary` | **INVALID**                                         | Must be re-fetched          |

### Material Subdirectories

`raw/materials/{type}/YYYY-MM-DD-{origin-slug}-vN.md`

Types: `papers/`, `specs/`, `docs/`, `repos/`, `blogs/`, `talks/`, `issues/`

### Size Limits

- Text content > 50KB: truncate with `[Content truncated at 50KB — full source at URL above]`
- No files > 1MB (split if needed)

### Append-Only Rule

Never edit files in `raw/`. Create new versions (`-v2`, `-v3`) instead.

### Source Quality Tiers

| Tier | Sources                                                  | Usage                               |
| ---- | -------------------------------------------------------- | ----------------------------------- |
| A    | Peer-reviewed paper, official spec, source code analysis | Design decisions                    |
| B    | Official blog, documentation, conference talk            | Design decisions with caveats       |
| C    | Third-party blog, tutorial, summary article              | Discovery only; must upgrade to A/B |

## 3. Frontmatter Contract

### Which Layers Require Frontmatter

| Layer                   | Metadata Format                        |
| ----------------------- | -------------------------------------- |
| `raw/materials/`        | 8-line header (no frontmatter)         |
| `raw/notes/`            | 8-line header (no frontmatter)         |
| `raw/manifests/`        | Frontmatter (it's a tracking document) |
| `reviews/`              | Frontmatter required                   |
| `gaps/`                 | Frontmatter required                   |
| `sources/`              | Frontmatter required                   |
| `synthesis/`            | Frontmatter required                   |
| `_synthesis-cross/`     | Frontmatter required                   |
| `_meta/decision-memos/` | Frontmatter required                   |

### Required Fields

```yaml
---
id: {kind-prefix}-{domain-NN}-{slug}-vN
kind: source|review|gap|synthesis|cross-synthesis
domain: NN-domain-name           # flat layout
domain: group/NN-short-name      # product-group tier layout
topics: [topic1, topic2]
status: active|working|validated|evergreen|stale|archived
confidence: high|medium|low
---
```

### Optional Fields

- `refs`: list of IDs this document references
- `last_verified_at`: YYYY-MM-DD

### Status Taxonomy

Maintain valid statuses in `_meta/taxonomies/status-taxonomy.yaml`. All frontmatter `status` values must match an entry. Common values:

| Status      | Meaning                         |
| ----------- | ------------------------------- |
| `active`    | Currently in progress           |
| `working`   | Reusable but still evolving     |
| `validated` | Checked against source material |
| `evergreen` | Curated durable knowledge       |
| `stale`     | Needs review before reuse       |
| `archived`  | Retained but out of active use  |

## 4. Semantic Gate Patterns

### Gap Schema

Every task in `gaps/` files must follow:

```markdown
## Task N: [title]

[description paragraph]

- Search queries: ["query1", "query2"]
- Expected source type: A/B/C
- Why this matters: [one sentence]
```

Gate check: count `## Task` headings, verify each has all three fields.

### Review Hygiene

Review files must:

- Contain `Tier A`, `Tier B`, or `Tier C` markers (quality assessment)
- Reference `raw/materials/` files by name (traceability)

### Citation Traceability

- `synthesis/` files must cite local domain paths (e.g., `raw/materials/docs/...`)
- `_synthesis-cross/` files must cite files from **>= 2** different domain directories
- No unsourced claims in synthesis layers

### Verification Script Pattern

For flat layout (`NN-domain/`):

```bash
# Gap schema check
for f in */gaps/*.md; do
  tasks=$(grep -c "^## Task" "$f")
  queries=$(grep -c "Search queries:" "$f")
  [ "$tasks" -eq "$queries" ] || echo "FAIL: $f"
done

# Review tier markers
for f in */reviews/*.md; do
  grep -qE "Tier [ABC]" "$f" || echo "FAIL: $f missing tier markers"
done

# Synthesis citation traceability
for f in */synthesis/*.md; do
  grep -q "raw/\|reviews/\|sources/\|gaps/" "$f" || echo "FAIL: $f missing citations"
done

# Cross-synthesis multi-domain check
for f in _synthesis-cross/*.md; do
  domains=$(grep -oE '0[1-9]-[a-z-]+' "$f" | sort -u | wc -l)
  [ "$domains" -ge 2 ] || echo "FAIL: $f cites <2 domains"
done
```

For product-group tier (`group/NN-domain/`):

```bash
# Helper: iterate all domains across product groups
all_domains() {
  for group in payments risk ledger; do
    for d in "$group"/[0-9][0-9]-*/; do
      [ -d "$d" ] && echo "$d"
    done
  done
}

# Gap schema check — note */[0-9][0-9]-*/ glob for nested layout
for f in */[0-9][0-9]-*/gaps/research-tasks-v*.md; do
  [ -f "$f" ] || continue
  tasks=$(grep -c '^## Task [0-9]' "$f" || true)
  queries=$(grep -c '^- Search queries:' "$f" || true)
  types=$(grep -c '^- Expected source type:' "$f" || true)
  whys=$(grep -c '^- Why this matters:' "$f" || true)
  [ "$tasks" -gt 0 ] && [ "$tasks" = "$queries" ] && [ "$tasks" = "$types" ] && [ "$tasks" = "$whys" ] \
    || echo "FAIL: $f"
done

# Review tier markers
for f in */[0-9][0-9]-*/reviews/critique-v*.md; do
  [ -f "$f" ] || continue
  grep -qE 'Tier [ABC]' "$f" || echo "FAIL: $f missing tier markers"
done

# Synthesis citation traceability (must cite own domain path)
for f in */[0-9][0-9]-*/synthesis/*.md; do
  [ -f "$f" ] || continue
  domain=$(echo "$f" | sed 's|/synthesis/.*||')
  grep -q "$domain/" "$f" || echo "FAIL: $f missing local domain citations"
done

# Cross-synthesis must cite >=2 domain directories
for f in _synthesis-cross/*.md; do
  [ -f "$f" ] || continue
  cited=0
  while IFS= read -r domain; do
    dname=$(echo "$domain" | sed 's|/$||')
    grep -q "$dname" "$f" && cited=$((cited + 1))
  done < <(all_domains)
  [ "$cited" -ge 2 ] || echo "FAIL: $f cites only $cited domain(s)"
done
```

## 5. Run Metadata

### Location

`_meta/runs/run-YYYYMMDD-{slug}/meta.yaml`

### Required Fields

```yaml
id: run-YYYYMMDD-{slug}
goal: [one sentence describing what the run produces]
status: active|validated # must be valid taxonomy status
systems:
  - system-name
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
produced_objects: [] # list of artifact IDs created by this run
```

### Optional Fields

```yaml
open_questions:
  - [unresolved question from the run]
```

### Status Lifecycle

`active` (run in progress) → `validated` (run complete and verified)

### Activity Logging (optional)

For long-running multi-agent runs, append to `activity.jsonl`:

```jsonl
{
  "ts": "2026-03-11T10:00:00Z",
  "agent": "W1-A",
  "action": "capture",
  "file": "..."
}
```

## 6. INDEX Freshness

### Pattern

Maintain a root `INDEX.md` with per-domain file counts and corpus-level statistics. Auto-generate from filesystem scan.

### Staleness Detection

```bash
# Compare INDEX.md declared count vs actual
declared=$(grep "Total" INDEX.md | grep -oE '[0-9]+')
actual=$(find . -name "*.md" -not -path "./.git/*" | wc -l)
[ "$declared" -eq "$actual" ] || echo "INDEX.md stale: declared=$declared actual=$actual"
```

### Integration

Wire index rebuild into the verification pipeline:

```yaml
# In Taskfile.yml
verify:
  cmds:
    - task: rebuild-indexes # auto-refresh INDEX.md
    - task: check:structure
    - task: check:semantic-gates
```

Or run after any operation that adds/removes files:

```bash
cd tools/researchctl && go run . rebuild-indexes
```

## 7. Verification Gate Template

### Standard Taskfile Tasks

| Task                   | Purpose                                         | Gate Type        |
| ---------------------- | ----------------------------------------------- | ---------------- |
| `check:structure`      | Every domain has all lifecycle dirs + README.md | Structure        |
| `check:completeness`   | Every domain has >= 1 file in raw/              | Structure        |
| `check:orphans`        | No .md files outside accepted paths             | Boundary         |
| `check:sizes`          | No file over 1MB                                | Quality          |
| `check:index`          | INDEX.md freshness                              | Drift            |
| `check:conventions`    | Required root files and section headings        | Agent legibility |
| `check:semantic-gates` | Gap schema, tier markers, citation traceability | Semantic         |
| `stats`                | Per-domain corpus statistics                    | Informational    |

### Composite Verify Command

```yaml
verify:
  desc: Run all verification gates
  cmds:
    - task: check:structure
    - task: check:completeness
    - task: check:orphans
    - task: check:sizes
    - task: check:index
    - task: check:conventions
    - task: check:semantic-gates
    - task: stats
```

### Convention Contract Template

For research repos, the `.convention-engineering.json` should include:

```json
{
  "contract_version": 1,
  "mode": "tracked",
  "profiles": ["research-corpus"],
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
  "mirror_policy": {
    "mode": "mirrored",
    "files": ["CLAUDE.md", "AGENTS.md"]
  },
  "evaluation_inputs": {
    "repo_risk": "standard"
  },
  "chunk_plan": {
    "enabled": false,
    "chunks": []
  },
  "research_root": ".",
  "product_groups": ["payments", "risk", "ledger"],
  "required_files": [
    "CLAUDE.md",
    "AGENTS.md",
    "ARCHITECTURE.md",
    "INDEX.md",
    "README.md",
    "Taskfile.yml"
  ],
  "taskfile_checks": {
    "Taskfile.yml": ["verify", "check:structure", "check:semantic-gates"]
  },
  "content_checks": [
    {
      "name": "claude-md-sections",
      "file": "CLAUDE.md",
      "required_markers": [
        "## Architecture Contract",
        "## Commands Cheat Sheet",
        "## Non-Negotiable Rules",
        "## Verification Gates"
      ]
    },
    {
      "name": "agents-md-sections",
      "file": "AGENTS.md",
      "required_markers": [
        "## Architecture Contract",
        "## Commands Cheat Sheet",
        "## Non-Negotiable Rules",
        "## Verification Gates"
      ]
    },
    {
      "name": "architecture-sections",
      "file": "ARCHITECTURE.md",
      "required_markers": ["## Topology", "## Layer Contract", "## Governance"]
    }
  ]
}
```

## Boundaries

This profile covers:

- Lifecycle-layer research repositories
- Frontmatter and metadata conventions
- Semantic verification patterns

This profile does NOT cover:

- Domain-specific content rules (those belong in per-repo CLAUDE.md)
- CLI tooling implementation (use Go profile for researchctl-style tools)
- CI/CD pipeline configuration (use core verification-gates.md)
