# Adapter Contract

What it takes for an agent harness to consume the canonical
`skills/<name>/SKILL.md` sources from this repo. A new harness is just a
new adapter that satisfies this contract; no upstream changes to canonical
content are required.

## Inputs

The adapter consumes:

- One or more **canonical skill directories** at `skills/<name>/`. Each
  contains:
  - `SKILL.md` with at minimum `name` + `description` frontmatter,
    optionally `version` (semver).
  - `references/` (markdown).
  - `schemas/`, `templates/`, `cli/`, etc. when the skill has them.
  - `CHANGELOG.md` when the skill is versioned.
  - `MAINTENANCE.md` when the skill is self-evolving.
- The adapter manifest at `adapters/manifest.json`. Today only declares
  `schema_version` and registered harness skill roots; future versions
  may add per-harness layout flags.

## Outputs

What the adapter produces under `adapters/<harness>/`. Every harness fits
one of three shapes:

### Shape A — per-skill directory mirror

Used by `claude-code` and `codex`. Each canonical skill becomes a
sibling directory:

```
adapters/<harness>/<skill>/SKILL.md
adapters/<harness>/<skill>/references/...      # if needed
```

Frontmatter is preserved byte-for-byte from canonical. The harness
discovers skills by walking the directory.

Special case: the **convention-engineering** skill — the harness's
"primary" skill in this kit — is mirrored at the adapter root rather
than in a subdirectory:

```
adapters/<harness>/SKILL.md      # convention-engineering, not in subdir
adapters/<harness>/<other>/SKILL.md
```

This asymmetry is historical (predates the multi-skill split) and may
be normalised in a future major release.

### Shape B — flat-file mirror

Used by `cursor`. Each canonical skill becomes one markdown file at the
adapter root:

```
adapters/<harness>/<skill>.md
```

Frontmatter may be stripped or rewritten depending on what the harness's
discovery model supports (Cursor reads `.mdc` files at the workspace
root and does not require Claude/Codex-style frontmatter).

### Shape C — single-file mirror

Used by harnesses with one global instruction file (Copilot, Aider,
generic AGENTS.md consumers). All canonical skills are folded into a
single document:

```
adapters/<harness>/<single-file>.md
```

Section headers identify each skill's region. Frontmatter is
discarded; the description text becomes a "When to use" prologue.

This shape is not yet implemented. The contract supports it.

## Sync Direction

Canonical → adapter. Adapter copies are read-only mirrors. Edits to an
adapter file are bugs unless the harness genuinely needs metadata not
expressible in canonical form (in which case the adapter file should
have a header pointer back to canonical and a clear marker for the
harness-specific delta).

The repo enforces this with `scripts/sync-adapters.sh`:

- No flag — overwrites adapter copies to match canonical.
- `--check` — exits non-zero on drift.

Wired into `task verify` (drift fails verification) and into
`task sync:adapters` / `task sync:adapters:check`.

## Versioning

When a canonical skill carries `version:` in its frontmatter:

- Adapter mirrors carry the same version (byte-identical frontmatter).
- The skill's `CHANGELOG.md` (canonical) is the upgrade reference;
  adapters do not maintain a separate changelog.
- Major version bumps may change the adapter contract itself (rename a
  canonical reference, restructure a directory). Adapter authors track
  the canonical changelog.

## Adding A New Harness

To add a new harness adapter:

1. Pick the output shape (A / B / C).
2. Add an entry to `adapters/manifest.json` declaring the harness name
   and its skill discovery root.
3. Create `adapters/<harness>/` and populate it by running
   `scripts/sync-adapters.sh`. Extend the script if the harness needs a
   shape it does not yet support.
4. Add a `README.md` under `adapters/<harness>/` describing how end
   users install it (e.g. `npx skills add ... -a <harness>`, or manual
   copy instructions for harnesses without a packaged installer).
5. If the new harness has a packaged installer in `npx skills`, ensure
   the manifest's skill_root matches what the installer expects.

## Anti-Patterns

- **Editing an adapter file directly.** The next sync wipes the change.
  Edit canonical, then sync.
- **Bidirectional sync.** Don't. Adapters never become a second source
  of truth.
- **Per-harness skill content.** If a skill needs different content for
  different harnesses, the divergence belongs in canonical (an `agents/`
  sidecar with runtime-specific metadata, or a separate skill), not in
  the adapter.
- **Stale auto-gen comments.** The skillsync runner is gone; do not
  reintroduce `<!-- regenerate with ark skill sync -->` markers. The
  sync script handles drift detection without inline markers.

## Out Of Scope

- Versioning of the adapter contract itself (this doc). Treated as
  evolving guidance, not as a stable API for third parties — yet.
  Promote to a stable contract version when an external party ships a
  third-party harness adapter.
- Per-harness frontmatter rewrite rules (e.g. Cursor `.mdc` flavour).
  Tracked here only as a shape; rewriters live in `scripts/sync-adapters.sh`.
