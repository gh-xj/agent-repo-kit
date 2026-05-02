# Repo-Local Skill Placement

Use this policy when deciding whether a repo should own local agent-skill
discovery (e.g. project-scoped reusable prompts, references, or scaffolds your
AI agent loads on demand).

## When To Use Repo-Local Roots

Use a repo-local skill root (for example `.<runtime>/skills/`) when the repo
needs project-local discovery for a specific AI agent runtime.

Use multiple roots when the repo needs more than one runtime to discover the
same local skill surface.

If the workflow is personal, environment-wide, or not meant to be versioned
with the repo, keep it in the user's global skill root instead of adding
repo-local placement.

## Dual-Runtime Placement

Keep the portable skill contract aligned across runtimes when more than one
is present:

- same skill intent
- same repo-local policy
- same reference model where practical

Only add runtime-specific metadata when the runtime actually needs it. Do not
split the core policy unless the runtime behavior genuinely differs.

## What This Convention May Create

This convention may create the runtime root directories themselves when a
repo needs local skill discovery (for example `.<runtime>/skills/`).

It may also create the immediate namespace directories under those roots if
the repo needs to reserve placement for future local skills.

## Handoff To A Skill-Authoring Surface

This convention owns placement policy only. It does not author skills.

Hand off to your skill-authoring surface (a dedicated skill-authoring tool,
template repo, or generator) for any of the following:

- skill router/manifest files (the convention doc itself)
- `references/`
- runtime metadata files
- skill scaffolds

Declare the repo's chosen `skill_roots:` in `.conventions.yaml` so the audit
workflow can verify them.
