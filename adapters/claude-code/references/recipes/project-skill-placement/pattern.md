# Recipe: Project Skill Placement

Status: stable
Codified: 2026-05-02

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

## Policy-Backed Skill Surface

When a repo has more than one project-local skill, add an explicit policy file
instead of relying on directory presence:

- `tools/skills/project-skill-policy.yml` — declared allowlist, canonical root,
  mirror root, placement reason, and external-skill metadata.
- `tools/skills/verify_project_skills.sh` — deterministic gate that checks the
  policy against the working tree.
- `skills-lock.json` — optional lock file for external skills copied from an
  upstream source.

Declare the same policy in `.conventions.yaml`:

```yaml
skill_roots:
  - .agents/skills
  - .runtime/skills

skill_policy:
  canonical_project_root: .agents/skills
  mirror_project_root: .runtime/skills
  mirror_rule: ".runtime/skills/<name> must be a symlink to ../../.agents/skills/<name>."
  policy_file: tools/skills/project-skill-policy.yml
  wrapper_local_rule: "Only add project-local skills for this repo's wrapper/router behavior."
  global_skill_rule: "Account, machine, or multi-repo skills belong in the global skill set by default unless listed under explicit_project_exceptions."
  explicit_project_exceptions:
    - name: example-skill
      decided: "YYYY-MM-DD"
      reason: "Why this account/tooling workflow belongs at project level."
      boundary: "What this project-level exception must not do."
```

The policy file is the executable allowlist. The `.conventions.yaml` block is
the human contract and should name every `project-exception` skill from the
policy file.

Use `references/recipes/project-skill-placement/templates/project-skill-policy.yml`
and `references/recipes/project-skill-placement/templates/verify_project_skills.sh`
as starting points.

## External And Official Skills

External skills are allowed when the repo needs a project-level copy of an
upstream skill, but they need stronger drift controls:

- keep only the files declared under `retainedPaths`;
- lock the copied skill file in `skills-lock.json` with source, source type,
  skill path, and SHA-256 hash;
- verify the paired CLI/package is installed locally when the skill depends on
  one;
- keep network freshness checks separate from the default local verification
  gate.

Recommended task split:

| Task | Network? | Purpose |
| --- | --- | --- |
| `task skills:verify` | No | Local skill shape, symlink mirrors, locks, hashes, and installed package presence. |
| `task skills:official:verify` | Yes | Latest upstream/package freshness for deliberate drift audits. |
| `task skills:official:update -- <skill>` | Yes | Update the package, refresh the copied skill, prune retained paths, and refresh lock hash. |

Document the cadence in the repo's agent contract. For local-only wrapper repos,
the default is manual: run the official check when updating external skills,
before publishing skill-policy changes, or during a deliberate drift audit.
Do not add network freshness checks to default `task verify` unless the repo has
explicitly chosen scheduled/network verification.

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
