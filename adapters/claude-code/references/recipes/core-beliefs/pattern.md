# Recipe: Core Beliefs

Status: stable
Codified: 2026-05-06

A trio of public docs that gives a project legible identity and decision traction: a manifesto of what the project IS, a list of invariants that MUST hold, and a list of anti-patterns the project explicitly REFUSES.

The trio split exists because each doc does a different job and benefits from a different shape. A single combined doc tries to be both poetic and verifiable and ends up neither.

## When To Use

Apply when:

- The project has a stable PUBLIC SURFACE — a CLI, library API, on-disk schema, configuration DSL, or other contract that outside users depend on.
- Identity-defending decisions recur ("should we add X?", "is feature Y in scope?"). The trio gives PR review a written compass.
- The audience includes maintainers AND outside contributors / agents who need to understand what the project is and isn't.
- The maintainer cares enough to write down "what we will not become."

Skip when:

- The project is internal-only with no public contributors or agent consumers.
- The project is too young to have stable identity. Adopt the trio after the first contentious feature debate, not before it.
- The project's audience is a single private owner (the manifesto and refusals don't have a reader).
- The project is a single-file utility or one-off script.

## Three Load-Bearing Decisions

These decisions matter at the *shape* level. File names and exact content are project-specific; the structural choices are not.

### 1. Three documents, three jobs

The trio splits "what should this project become?" into three artifacts that don't try to be each other:

| Doc                | Job                       | Audience                            | Citable as       |
| ------------------ | ------------------------- | ----------------------------------- | ---------------- |
| `core-belief.md`   | Identity / spirit         | Anyone — newcomers, marketing       | "belief #N"      |
| `invariants.md`    | Hard constraints / facts  | Maintainers, PR reviewers, agents   | "violates INV-N" |
| `anti-patterns.md` | Explicit refusals         | Feature-request triagers, reviewers | "matches ANTI-N" |

Combining them produces a doc that is too long to read end-to-end (loses manifesto quality) and too vague to cite (loses engineering value). Splitting them lets each artifact be the right shape for its job.

### 2. Invariants must be grounded in code, with verify commands

Every entry in `invariants.md` is a fact that is true *now* and verifiable *without running the binary*. The verify line is the load-bearing payload:

```
INV-1   <load-bearing claim>.
        Verify: `<concrete grep / test command>` <expected result>.
```

An invariant you cannot grep is an aspiration, not an invariant. If the verify command would be impractical (slow, requires runtime), reframe the claim or move it to `core-belief.md`.

### 3. Anti-patterns describe categories, not individual features

Each anti-pattern names a *class* of feature request: daemons, external service dependencies, vendor-coupled identity. Per-anti-pattern shape: looks-like / why-not / instead / inverts.

The trailing "*Inverts INV-N, belief #M.*" cross-links the three docs. A reviewer hitting an anti-pattern can jump to the underlying invariant or belief and confirm the rejection has teeth.

## File Inventory

```
<project>/<docs_root>/
├── core-belief.md      # manifesto / identity layer
├── invariants.md       # RFC-2119 normative facts, grounded in code
└── anti-patterns.md    # explicit refusals, cross-linked
```

Required count: 3 files. Typically under `docs/` (or whatever `docs_root:` declares). The trio lives next to the existing docs taxonomy; it does not replace it. Most projects also cross-link from `AGENTS.md` / `CLAUDE.md` / `README.md`.

## `.conventions.yaml` Extensions

The recipe adds one well-known top-level key:

```yaml
core_beliefs:
  enabled: true
  path: docs/        # optional; defaults to <docs_root>
```

`scripts/verify.sh` asserts that `<path>/core-belief.md`, `<path>/invariants.md`, and `<path>/anti-patterns.md` all exist. The verify gate is *existence-only* by design — structurally validating INV/ANTI numbering is more cost than benefit at this scale.

The shorthand `core_beliefs: true` is also accepted (path defaults to `<docs_root>`).

## Templates

Reference templates live in [`./templates/`](./templates/):

- `core-belief.md.tmpl` — manifesto skeleton. Unix-philosophy form (numbered imperatives + refusals + coda) is recommended for legibility, but the form is the project's choice.
- `invariants.md.tmpl` — RFC-2119 normative; INV-N entries with claim + why + verify command.
- `anti-patterns.md.tmpl` — ANTI-N entries with looks-like / why-not / instead / inverts.

Each template carries a skeleton and an inline comment block explaining usage; replace placeholders, then strip the comment block before commit.

## Adopt

See [`bootstrap.md`](./bootstrap.md) for the procedure.

## Anti-Patterns

- **Stitching the three into one file feels tidy and is wrong.** The combined doc tries to be both poetic and verifiable; it ends up neither. Resist the merge even when the project is small.
- **Aspirational invariants poison the file.** A claim without a verify command is a wish, not an invariant. Move wishes to `core-belief.md`.
- **Anti-patterns that name a single feature.** Anti-patterns name categories that recur across feature requests, not individual PRs.
- **Generating the trio mechanically from one of the three.** Each doc benefits from independent judgment.
- **Adopting the trio without verify wiring.** Without `scripts/verify.sh` checking the files exist, the trio drifts.
- **Numbering churn from edits.** Append new invariants at the end. Renumber only on major review passes. Stable IDs make external citations (`#INV-4` from a PR comment) keep meaning.
