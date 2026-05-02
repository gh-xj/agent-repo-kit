# Agent Contract Files

How `CLAUDE.md` and `AGENTS.md` should be structured so an AI agent can work
in any repo.

## Scope

This skill cares about two files:

- `AGENTS.md` — runtime-agnostic contract.
- `CLAUDE.md` — Claude-specific contract.

`ARCHITECTURE.md`, `DESIGN.md`, and similar synthesis docs are repo content,
not skill-owned. The skill does not prescribe their shape.

## Required Sections (minimum)

Each agent contract file should contain these sections. Order is suggested,
not enforced.

### 1. Architecture Pointer

A short layer/dependency sketch and the rule(s) the repo cares about. Keep it
to 5-10 lines. If the architecture warrants more, point at a separate doc.

### 2. Commands Cheat Sheet

Exact commands, not descriptions. Build, test, lint, format, verify, dev
server, codegen.

### 3. Non-Negotiable Rules

Repo-specific rules an agent must respect. Type-safety mode, forbidden
patterns, naming, file size limits, security hygiene.

### 4. Verification Gate

The single command that runs all checks (`task verify` or equivalent), plus
the pre-commit hook behavior.

### 5. Pointers

Links to deeper sources of truth: docs taxonomy root, skill directories,
operational ops (`.work/`, `.wiki/`) when adopted.

## Mirroring (Optional)

When the repo serves more than one AI runtime, `AGENTS.md` and `CLAUDE.md`
hold identical content. Pointer text:

> "This document is intentionally mirrored in `AGENTS.md` and `CLAUDE.md`."

If the repo only uses one runtime, a single file is sufficient. Do not add
the second file just for symmetry.

## Knowledge Freshness

Each contract file should explicitly say:

- What to verify against live code before trusting any section.
- How to update the file when drift is found during a session.

This makes every session a self-healing opportunity. The contract improves
with use.
