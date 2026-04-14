# Agent Knowledge Architecture

How to structure documentation so AI agents can navigate and work in any repo.

## Agent Contract Template

Every repo needs an agent contract file (commonly `AGENTS.md`, or a
runtime-specific name like `CLAUDE.md`) with these 5 sections (minimum):

### 1. Architecture Contract

- Layer dependency diagram (ASCII, 5-10 lines)
- Which packages/modules can import which
- Non-negotiable rules (e.g., "model has zero internal imports")

### 2. Commands Cheat Sheet

- Build, test, lint, format, verify commands
- Exact commands, not descriptions ("run `task verify`", not "run the verify command")
- Dev server, code generation, deployment commands if applicable

### 3. Code Generation Workflow (if applicable)

- Ordered steps: IDL change -> generate -> restart -> generate client -> type check
- Which files are generated (DO NOT EDIT markers)
- How to regenerate after schema changes

### 4. Non-Negotiable Rules

- Type safety requirements (strict mode, no `any`, forbidden patterns)
- Import restrictions beyond architecture contract
- Naming conventions, file size limits, complexity thresholds
- Security hygiene (no console.log, no credentials in code)

### 5. Verification Gates

- What to run before claiming done (exact commands)
- Pre-commit hook behavior
- Full verification command (`task verify` or equivalent)

## Agent Contract Mirroring

Pattern: when a repo serves more than one AI agent runtime, the runtime-specific
contract files (e.g. `AGENTS.md`, `CLAUDE.md`, or any other runtime convention)
contain identical content (mirrored).

- Both must be updated in the same change
- Machine-enforceable via the contract checker's docs-governance check
- Pointer text: "This document is intentionally mirrored in `<file-a>` and `<file-b>`."

When to use: repos where multiple AI agent runtimes may operate (resilience against tool-specific config discovery).

When to skip: repos that only use one AI agent runtime. A single contract file is sufficient.

## ARCHITECTURE.md Pattern

Reference-first, not narrative dump:

- Service map (ASCII diagram showing components and data flow)
- Data flow (entry points, storage layers, external integrations)
- Detection/business rules index (what rules exist, where they live)
- On-demand sections (agents read what they need, not the whole file)

Keep under 1000 lines. If it grows beyond that, split into reference docs and use a routing table.

## Cross-Repo Skill Creation Criteria

Create a dedicated cross-repo skill when:

- The service has 3+ consumers or dependents
- Cross-repo changes happen regularly
- Debugging requires knowledge spanning multiple repos
- A routing table would save agents significant exploration time

Skill structure: convention doc router + `references/` directory + changelog.

## Knowledge Freshness Contract

Every reference document should include:

1. **"Before trusting" section**: What to verify against live code before making decisions based on this doc.
2. **"After discovering" section**: How to update when drift is found during a session.
3. **Changelog table**: Date, change description, affected files.

This makes every session a self-healing opportunity. The skill improves with use.
