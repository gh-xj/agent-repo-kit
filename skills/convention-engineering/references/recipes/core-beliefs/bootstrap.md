# Bootstrap Core Beliefs

Walk a project through producing the `core-belief.md` / `invariants.md` /
`anti-patterns.md` trio. Reproducible procedure; an agent can follow this
cold against any project that fits the
[core-beliefs pattern](../patterns/core-beliefs.md).

## Pre-flight

Confirm the project fits the pattern's "When to use" criteria. Skip if the
project is a brain repo, an experiment, an internal service, a UI app, or
a one-off script — the trio adds maintenance cost there for little return.

If the project has a maintainer-private wrapper (epic-wrapper pattern),
the trio belongs in the PUBLIC repo, not the wrapper. The wrapper can
hold the design rationale that produced the trio.

## Phase 1 — Map the public surface

Goal: name what the project IS so the beliefs can be grounded.

1. Read the project's `README.md`, top-level `AGENTS.md` / `CLAUDE.md`,
   and any existing `docs/` content.
2. List the public contracts:
   - CLI surface (commands, flags, exit codes)
   - Library API (exported types and functions)
   - On-disk schema (file formats, directory layouts)
   - Configuration / DSL
3. Identify the project's *one-line stance* — the inversion that captures
   what makes it different. Examples:
   - work-cli: "A tracker records work; an orchestrator performs it."
   - Symphony (OpenAI): "manage work instead of supervising coding agents".
   - Linear: "lost art of building true quality software".

## Phase 2 — Audit invariants from code

Goal: 8–12 facts that are TRUE in the code today and verifiable.

For each candidate invariant:

1. State the claim in one sentence.
2. Identify the file(s) and line(s) it is grounded in.
3. Write the verify command — typically `grep`, occasionally a test
   reference. The verify must produce an unambiguous answer (specific
   match count or empty result).
4. Reject the candidate if no verify command can be written; move it to
   the belief layer instead.

Categories that yield strong invariants:

- Code organization (e.g. "no `os.Exit` outside `main.go`").
- Test seams ("the X interface is the only seam tests substitute").
- Atomicity / concurrency primitives ("mutations go through
  `withMutationLock`").
- Schema enums ("the set of `Status` values is exactly {a, b, c, d, e}").
- Schema versioning ("every durable record carries `schema_version`").
- Output contracts ("`--json` output is stamped with `schema_version`").
- Offline / network boundaries ("no subcommand requires network access
  for primary function").

## Phase 3 — Elicit the manifesto

Goal: ~10 short imperatives that capture the project's identity.

The recommended form is Unix philosophy: numbered list, each entry an
imperative + one supporting clause. Other forms (Tractatus, HashiCorp Tao
essay) are acceptable but Unix is most legible to outside readers.

Source material:

- The one-line stance from Phase 1.
- The invariants from Phase 2 (each invariant typically has a
  philosophical mate in the manifesto).
- The project's stated NON-goals (read existing design docs and look
  for "we will not", "out of scope", "rejected" sections).

Drafting heuristic: write 15 candidate beliefs, cut to 10 by merging
overlaps and discarding ones that don't recur in real decisions.

## Phase 4 — Identify anti-patterns

Goal: 6–10 categories of feature request the project explicitly refuses.

For each invariant from Phase 2, ask: *what feature shape would invert
this?* The answer is often an anti-pattern. Common categories:

- Daemons / long-running processes (inverts statelessness).
- External service dependencies (inverts locality).
- Persistent caches or sidecar databases (inverts statelessness).
- Coordination state in lifecycle field (inverts schema separation).
- Vendor-coupled identity (inverts agent-agnosticism).
- Configurable workflow engine (inverts simplicity).
- Lifecycle hooks that execute code (inverts tool-not-orchestrator).
- Friction at capture, automation at commitment (inverts intake design).

Each anti-pattern: looks-like / why-not / instead / *inverts INV-N,
belief #M*.

## Phase 5 — Write the three files

Use the templates from `references/recipes/core-beliefs/templates/`:

- `<docs_root>/core-belief.md` from `core-belief.md.tmpl`
- `<docs_root>/invariants.md` from `invariants.md.tmpl`
- `<docs_root>/anti-patterns.md` from `anti-patterns.md.tmpl`

Each template carries skeleton structure; replace placeholders with
project-specific content. Add cross-link footers to all three so a
reader can move between them.

## Phase 6 — Wire the verify gate

Add to `.conventions.yaml`:

```yaml
core_beliefs:
  enabled: true
  path: docs/
```

Update `scripts/verify.sh` (or the repo's verify equivalent) to assert
the three files exist when `core_beliefs.enabled` is true. Existence-only
check; do not attempt to validate INV/ANTI structure.

Reference snippet for `verify.sh`:

```bash
# core_beliefs trio
if [ "$(yq -r '.core_beliefs.enabled // .core_beliefs // false' .conventions.yaml)" = "true" ]; then
  beliefs_path=$(yq -r '.core_beliefs.path // .docs_root // "docs"' .conventions.yaml)
  for f in core-belief.md invariants.md anti-patterns.md; do
    [ -f "$beliefs_path/$f" ] || fail "core_beliefs: $beliefs_path/$f missing"
  done
fi
```

## Phase 7 — Cross-link

Add pointers in:

- `README.md` — one-line mention near the project status section.
- `AGENTS.md` / `CLAUDE.md` — pointer in the "Pointers" or "Architecture"
  section so agents can find the trio.
- `docs/README.md` — list the trio alongside the other doc folders.

## Smoke Test

1. `task verify` (or equivalent) exits 0 with the new files in place.
2. A fresh reader can find each file via README → AGENTS.md → docs trio.
3. Pick a recent contentious PR or feature debate and ask: "would the
   trio have settled this faster?" If not, the trio is too generic;
   tighten the entries that don't bite.

## Common Gotchas

- **Inventing invariants instead of auditing.** "No subcommand performs
  a network call" sounds good, but `grep -rn 'net/http'` may show
  exceptions. Audit first; phrase the invariant to match reality (name
  the exception explicitly when it exists).
- **Manifesto bloat.** Aiming for 15 beliefs instead of 10 produces
  filler. Cut hard. Each belief should map to a real recurring decision.
- **Cross-doc reference rot.** When you renumber an INV, update the
  matching ANTI's "inverts" trailer. Periodic audit pass.
- **Verbose templates that survive into the committed doc.** Template
  comment blocks must be stripped before commit. Skim the diff for
  `<!-- Template usage:` blocks.
