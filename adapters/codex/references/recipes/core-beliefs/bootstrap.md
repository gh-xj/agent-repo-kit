# Bootstrap: Core Beliefs

Walk a project through producing the [`core-belief.md` / `invariants.md` / `anti-patterns.md` trio](./pattern.md). Three passes — audit, elicit, write. Skip a pass when the answer is already obvious; depth varies by project.

## Pass 1 — Audit invariants from code

Read the public surface (CLI flags, library API, on-disk schema, configuration). Write down the facts that are TRUE *now* and verifiable *without running the binary*. Each invariant gets a one-line claim and a concrete verify command (typically `grep`, occasionally a test reference).

If you cannot write a verify command for a candidate, it is an aspiration, not an invariant. Push it to the manifesto instead.

## Pass 2 — Elicit the manifesto and the refusals

The manifesto names what the project IS; the refusals name what it WILL NOT BECOME. Both flow from the audit and from the project's own design history (existing READMEs, design docs, "rejected" or "out of scope" sections).

Form is the project's choice. Unix-philosophy imperatives, numbered propositions, or a flowing essay all work — pick what fits the project's voice.

For each invariant, ask: *what feature shape would invert this?* The answer is often an anti-pattern. Anti-patterns name categories, not individual features.

## Pass 3 — Write, wire, and cross-link

Use the templates in [`./templates/`](./templates/). Each template is a skeleton with an inline comment block explaining usage; replace placeholders, strip the comment block, commit.

Wire the verify gate by adding `core_beliefs: true` (or `{ enabled: true, path: <docs_root> }`) to `.conventions.yaml`, then having `scripts/verify.sh` assert the three files exist when `core_beliefs.enabled` is true. Existence-only — do not attempt structural validation.

Cross-link from `README.md`, the agent contract files, and `docs/README.md` so the trio is discoverable.

## Smoke test

`task verify` (or equivalent) exits 0. A fresh reader can navigate from `README` to the trio. Pick a recent contentious PR and ask: would the trio have settled it? If not, tighten the entries that don't bite.
