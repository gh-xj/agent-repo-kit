# Pattern-to-Script Extraction

When maintaining skills, proactively identify repeated patterns that should become scripts. Scripts make skills faster, more reliable, and save context window.

## Core Principle

**If Claude re-derives the same logic every session, it should be a script.** Any deterministic, repeated operation — regardless of domain — is a candidate.

## When to Propose

During any skill create/update, look for:

- **Deterministic logic** described in prose that Claude must interpret and execute step-by-step each time
- **Data lookups or mappings** (tables, dictionaries, categorization rules) embedded in SKILL.md
- **Multi-step transformations** where input -> processing -> formatted output follows the same pattern every time
- **External system interaction** (databases, APIs, file systems) with boilerplate connection/query logic
- **Calculations or arithmetic** that Claude might get wrong (offsets, conversions, aggregations)
- **Formatting rules** with specific conventions (separators, markers, grouping, ordering)

**Threshold:** If you spot 2+ of these in a skill, propose extraction.

## How to Propose

```
I notice this skill has [describe the repeated patterns]. These are candidates
for script extraction:

1. **[pattern]** -> `script-name command1` - [what it replaces in SKILL.md]
2. **[pattern]** -> `script-name command2` - [what it replaces]

This would: reduce SKILL.md by ~[N] lines, make output deterministic,
and eliminate [specific error risk].

Shall I build this?
```

## Why Extract to Scripts

**Determinism.** Claude re-deriving logic each session introduces variance — different formatting, missed edge cases, inconsistent output. A script runs the same way every time. The more deterministic the operation, the stronger the case for extraction.

## Prerequisite: Verified Patterns Only

**Never extract unverified patterns.** A pattern must be confirmed working across 2+ sessions before it becomes a script. Premature extraction bakes in bugs.

| Pattern State                            | Action                             |
| ---------------------------------------- | ---------------------------------- |
| Just emerged (1 session)                 | Keep as SKILL.md instructions      |
| Verified (2+ sessions, edge cases known) | Extract to script                  |
| Evolving (works but still changing)      | Wait — extract after it stabilizes |

## Language Choice

**Default to Go** for skill scripts — follow your repo's Go style conventions for project structure and CLI helpers.

| Signal                                             | Language         |
| -------------------------------------------------- | ---------------- |
| Compiled binary, fast startup, type safety         | **Go** (default) |
| Quick data munging, prototyping, heavy library use | **Python**       |
| Simple glue, <20 lines, no logic                   | **Bash**         |

Go is preferred because: compiled binaries have zero startup overhead, type safety catches errors at build time, and common Go tooling conventions (structured logging, a Taskfile for local/CI parity, cobra for command wiring) are already widely available.

## Script Design Guidelines

- **One tool, multiple commands** — group related operations under a single entry point
- **Structured output (JSON) for data** — Claude can parse and reason over it
- **Plain text output for paste-ready content** — when output goes directly into files
- **Consistent flags** — same parameter names across commands for the same concept
- **Context-aware behavior** — handle edge cases in code (e.g., past vs present, empty data, missing fields)
- **Safe defaults** — read-only access, no destructive operations without explicit flags

## Script Location Convention

```
skill-name/
├── SKILL.md
├── scripts/           # Source code
│   └── ...            # Language-appropriate structure
└── skill-binary       # Built artifact (if compiled)
```

Build automation: use `Taskfile.yml` with language-appropriate tooling (e.g. `go run` / `go test` for Go, `uv run` for Python).

## After Extraction: Update SKILL.md

Replace verbose instructions with command references:

**Before (N steps of prose):**

```markdown
## Do the Thing

1. Connect to [system]
2. Look up [data] using [these rules]
3. Transform using [this mapping table with 30 entries]
4. Format according to [these 5 conventions]
5. Handle [these edge cases]
```

**After (1 command + output description):**

```markdown
## Do the Thing

`skill-tool do-thing --param value`
Returns: `{ structured: "output" }` — or paste-ready text.
```

## What Makes a Good Extraction

| Good Candidate                 | Poor Candidate               |
| ------------------------------ | ---------------------------- |
| Same logic every invocation    | Different approach each time |
| Deterministic output           | Requires judgment/creativity |
| Involves data/computation      | Purely conversational        |
| Error-prone when done manually | Simple enough to never fail  |
| Used frequently                | Rare one-off operation       |
| Can be validated by running it | Needs human review to verify |
