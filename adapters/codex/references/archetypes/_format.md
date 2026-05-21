# Archetype Format

An archetype is a whole-repo shape. Adopting one means scaffolding the files, directory layout, and `.conventions.yaml` block that define the shape; layering recipes on top is then optional.

A new archetype lives at `references/archetypes/<name>/` and contains:

- `pattern.md` — the description (required, ≤ 250 lines)
- `bootstrap.md` — the procedure to apply this archetype (required)
- `migration.md` — when migrating from a predecessor or older variant (optional)
- `templates/` — concrete scaffold files referenced from `pattern.md` and `bootstrap.md`

## Required sections in `pattern.md`

In order:

1. **Title + Status header**

   ```markdown
   # Archetype: <Name>

   Status: stable | experimental | deprecated
   Codified: YYYY-MM-DD
   ```

2. **Brief intro** (2–4 lines).
3. **When To Use** — apply criteria, skip criteria.
4. **Load-Bearing Decisions** — numbered list of decisions whose change requires a new archetype version. Three to five is typical.
5. **File Inventory** — the directory tree.
6. **`.conventions.yaml` Extensions** — the YAML block(s) the archetype canonizes.
7. **Templates** — list of files in `templates/` with one-line descriptions.
8. **Worked Example** — pointer at a real instance.

## Optional sections

- **Per-Realm Gate Matrix** — only for archetypes that introduce a realm/permission split (e.g. `brain`).
- **Anti-Patterns** — categories of misadoption that recur.
- **Gotchas** — at most three; load-bearing only.

## What stays out of `pattern.md`

- Inline templates → `templates/`.
- Step-by-step bootstrap procedure → `bootstrap.md`.
- Migration procedure → `migration.md`.

## Length target

`pattern.md` ≤ 250 lines. Beyond that, extract more.

## Versioning

The `Status` header and `Codified` date are the lightweight version mechanism. A breaking change to load-bearing decisions warrants a status demotion (`stable` → `experimental`) until the new decision settles, or a new archetype with a different name. Document the change in the skill's top-level `CHANGELOG.md`.
