# Migration: 1.x → 2.0

The 2.0 release reorganizes `references/` into four buckets: `concepts/`, `archetypes/`, `recipes/`, `meta/`. The old `core/`, `operations/`, `patterns/`, `templates/` directories are removed.

If you have docs (`CLAUDE.md`, `AGENTS.md`, comments, READMEs) that reference the old paths, update them per the table below.

## Path mapping

| Old path                                                           | New path                                                            |
| ------------------------------------------------------------------ | ------------------------------------------------------------------- |
| `references/core/agent-first-principles.md`                        | `references/concepts/agent-first-principles.md`                     |
| `references/core/conventions-yaml.md`                              | `references/concepts/conventions-yaml.md`                           |
| `references/core/verification-gates.md`                            | `references/concepts/verification.md` *(merged + trimmed)*          |
| `references/core/verify-script-pattern.md`                         | `references/recipes/verify-script/pattern.md`                       |
| `references/core/agent-knowledge.md`                               | `references/recipes/agent-knowledge/pattern.md`                     |
| `references/core/docs-taxonomy.md`                                 | `references/recipes/docs-taxonomy/pattern.md`                       |
| `references/core/project-skill-placement.md`                       | `references/recipes/project-skill-placement/pattern.md`             |
| `references/core/adapter-contract.md`                              | `references/meta/adapter-contract.md`                               |
| `references/operations/bootstrap-workflow.md`                      | `references/meta/bootstrap-workflow.md`                             |
| `references/operations/audit-workflow.md`                          | `references/meta/audit-workflow.md`                                 |
| `references/operations/work.md`                                    | `references/recipes/work/pattern.md`                                |
| `references/operations/wiki.md`                                    | `references/recipes/wiki/pattern.md`                                |
| `references/operations/bootstrap-core-beliefs.md`                  | `references/recipes/core-beliefs/bootstrap.md`                      |
| `references/patterns/epic-wrapper.md`                              | `references/archetypes/epic-wrapper/pattern.md`                     |
| `references/patterns/brain.md`                                     | `references/archetypes/brain/pattern.md`                            |
| `references/patterns/core-beliefs.md`                              | `references/recipes/core-beliefs/pattern.md`                        |
| `references/templates/wiki/`                                       | `references/recipes/wiki/templates/`                                |
| `references/templates/core-beliefs/`                               | `references/recipes/core-beliefs/templates/`                        |

## Concept renames

The mental model is now: **archetype** (whole-repo shape, pick one) + **recipes** (stackable conventions, layer many) + **concepts** (always-on principles) + **meta** (about the skill itself).

- Old `core/` was a junk drawer of foundational + opinion + meta. Split into:
  - Foundational principles → `concepts/`
  - Opinionated reusable conventions → `recipes/`
  - About the skill itself → `meta/`
- Old `operations/` mixed generic workflows with operation-domain docs:
  - Generic workflows (`bootstrap-workflow`, `audit-workflow`) → `meta/`
  - Operation-domain docs (`work`, `wiki`) were really recipes → `recipes/`
- Old `patterns/` mixed whole-repo shapes with stackable conventions:
  - Whole-repo shapes (`epic-wrapper`, `brain`) → `archetypes/`
  - Stackable conventions (`core-beliefs`) → `recipes/`

## Updating an adopter repo

1. **`.conventions.yaml`** — the schema location is unchanged (`skills/convention-engineering/schemas/conventions.schema.json`). No edit needed if you reference the schema.
2. **`CLAUDE.md` / `AGENTS.md`** — search for `references/(core|operations|patterns|templates)/` and replace per the table above.
3. **In-repo docs** — same search/replace.

A one-shot perl sweep against your repo:

```bash
find . -type f \( -name "*.md" -o -name "*.yaml" \) -exec perl -i -pe '
  s|references/core/agent-first-principles\.md|references/concepts/agent-first-principles.md|g;
  s|references/core/conventions-yaml\.md|references/concepts/conventions-yaml.md|g;
  s|references/core/verification-gates\.md|references/concepts/verification.md|g;
  s|references/core/verify-script-pattern\.md|references/recipes/verify-script/pattern.md|g;
  s|references/core/agent-knowledge\.md|references/recipes/agent-knowledge/pattern.md|g;
  s|references/core/docs-taxonomy\.md|references/recipes/docs-taxonomy/pattern.md|g;
  s|references/core/project-skill-placement\.md|references/recipes/project-skill-placement/pattern.md|g;
  s|references/core/adapter-contract\.md|references/meta/adapter-contract.md|g;
  s|references/operations/bootstrap-workflow\.md|references/meta/bootstrap-workflow.md|g;
  s|references/operations/audit-workflow\.md|references/meta/audit-workflow.md|g;
  s|references/operations/work\.md|references/recipes/work/pattern.md|g;
  s|references/operations/wiki\.md|references/recipes/wiki/pattern.md|g;
  s|references/operations/bootstrap-core-beliefs\.md|references/recipes/core-beliefs/bootstrap.md|g;
  s|references/patterns/epic-wrapper\.md|references/archetypes/epic-wrapper/pattern.md|g;
  s|references/patterns/brain\.md|references/archetypes/brain/pattern.md|g;
  s|references/patterns/core-beliefs\.md|references/recipes/core-beliefs/pattern.md|g;
  s|references/templates/core-beliefs/|references/recipes/core-beliefs/templates/|g;
  s|references/templates/wiki/|references/recipes/wiki/templates/|g;
' {} +
```

## Adapter mirrors

Adapter copies under `adapters/<harness>/` will drift after the canonical restructure. Run `scripts/sync-adapters.sh` to bring mirrors back in sync. CI's `task verify` includes the drift check.

## Schema changes

- `operations:` enum is now open (was `["work", "wiki"]`). Recommended values are `work` and `wiki`; project-local recipes are permitted.
- `min_work_version` is marked deprecated in the description. It still works for backward compatibility. Tool-specific version constraints should now live in `checks:`.
- New typed key: `core_beliefs:` — see `references/recipes/core-beliefs/pattern.md`.
