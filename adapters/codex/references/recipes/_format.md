# Recipe Format

A recipe is a stackable convention you can apply on top of any archetype. Recipes are smaller and more focused than archetypes.

A new recipe lives at `references/recipes/<name>/` and contains:

- `pattern.md` — the description (required, ≤ 150 lines)
- `bootstrap.md` — only when adoption requires more than reading `pattern.md` (optional)
- `templates/` — concrete scaffold files (optional)

Many recipes are a single `pattern.md`.

## Required sections in `pattern.md`

In order:

1. **Title + Status header**

   ```markdown
   # Recipe: <Name>

   Status: stable | experimental | deprecated
   Codified: YYYY-MM-DD
   ```

2. **Brief intro** (1–3 lines).
3. **When To Use** — when to adopt; when to skip.
4. **The Convention** — the actual rules / structure / artifacts the recipe defines.
5. **Adopt** — short adoption steps; for complex recipes, point at `bootstrap.md` instead.

## Optional sections

- **Templates** — pointer to `templates/` (when the recipe ships scaffold files).
- **`.conventions.yaml` Extensions** — only when the recipe adds a typed key.
- **Anti-Patterns** — short list, only if the recipe has a recurring misadoption shape.
- **Worked Example** — pointer to a real instance.

## Length target

`pattern.md` ≤ 150 lines. The smallest recipes are ≤ 80 lines.

## Versioning

Same `Status` + `Codified` header convention as archetypes. Breaking changes warrant a CHANGELOG entry.
