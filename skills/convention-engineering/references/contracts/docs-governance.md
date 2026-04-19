# Docs Governance Contract

## Taxonomy Contract

- Select one canonical docs root per repo: `docs/` (tracked) or `.docs/` (local-overlay).
- Use intent-first subfolders from `references/core/docs-taxonomy.md`:
  - `requests/`, `planning/`, `plans/`, `implementation/`, `taxonomy/` (plus optional `reviews/`).
- For feature work, require the lifecycle chain:
  - request -> design -> plan (before implementation).
- Avoid mixed active roots (`docs/` + `.docs/`) unless migration is explicitly documented.

## Structural Contract

- Define required paths (files/dirs).
- Define forbidden legacy paths.
- Fail when required paths are missing or legacy paths reappear.

## Naming Contract

- Enforce date-prefixed file names for key docs:
  - `requests/YYYYMMDD_rfi_<topic>.md`
  - `planning/YYYY-MM-DD_<topic>_design.md`
  - `plans/YYYY-MM-DD-<topic>.md`
- Keep topic slugs lowercase and stable.

## Content Contract

- For key docs, require marker sections/phrases.
- Keep markers stable and concise.
- Use content markers for onboarding quality and convention discoverability checks.

## Pointer Contract

Use canonical pointer checks with mode:

- `all`: every pointer must match.
- `any`: at least one pointer must match (useful for mirrored-file variants).

## Output Contract

Each failed check must include:
- check name
- file/path scope
- missing marker/token/path detail
