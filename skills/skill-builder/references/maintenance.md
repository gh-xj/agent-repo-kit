# Maintenance

## Changelog

- `2026-03-18`: Refactored `SKILL.md` into a smaller router, moved runtime layout and backup-repo rules into references, and added repo-owned CLI guidance for `tools/<name>/`.
- `2026-03-14`: Added cross-platform (Claude Code + Codex) troubleshooting entries.
- `2026-03-08`: Moved health indicators and troubleshooting out of `SKILL.md` to keep the entry file under 500 lines.
- `2026-03-08`: Added skill metadata example alignment (`version`, `last_updated`) with current update cadence.

## Skill Health Indicators

Track informally over time:

| Metric              | Healthy                  | Unhealthy                 |
| ------------------- | ------------------------ | ------------------------- |
| Activation accuracy | Triggers when expected   | False positives/negatives |
| Workflow completion | Users finish tasks       | Abandoned mid-workflow    |
| Reference usage     | References actually read | Never accessed            |
| Update frequency    | Periodic refinement      | Stale >6 months           |

## Metadata

Portable custom skills should default to:

```yaml
---
name: skill-name
description: Use when ...
---
```

One optional, recognised extra field: `version:` (semver). Use it on
skills that have downstream adopters who need to pin or detect breaking
changes. Pair it with a per-skill `CHANGELOG.md` at the skill root. Skip
it on skills that are repo-private or pre-1.0 with no consumers.

If other extra metadata is useful, prefer:

- a changelog section in `SKILL.md` (small skills) or a `CHANGELOG.md`
  sibling (versioned skills)
- runtime-specific metadata such as `agents/openai.yaml`
- repo-local tracking docs

Do not treat other frontmatter fields as the shared default for portable
Claude/Codex skills.

## Versioning Discipline

When a skill carries a `version:`:

- **Major bump (`x.0.0`)** — breaking trigger phrasing, breaking schema,
  removed references that adopters might link to, renamed front-door
  files. Adopters need to migrate.
- **Minor bump (`0.x.0`)** — new capability, new optional opt-in,
  expanded trigger coverage that strictly adds. Adopters can ignore.
- **Patch bump (`0.0.x`)** — wording fixes, reference cleanup, clarity
  edits.

Land every breaking change with a `CHANGELOG.md` entry that names the
break explicitly. Adopters read the changelog, not the diff.

## Troubleshooting

| Problem                                  | Cause                               | Solution                                                                                                                                                                                                                                           |
| ---------------------------------------- | ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Skill not activating                     | Description too narrow              | Add trigger phrases to description                                                                                                                                                                                                                 |
| Skill activating too often               | Description too broad               | Be more specific about triggers                                                                                                                                                                                                                    |
| Changes not reflected                    | Session cache or stale managed copy | Restart Claude Code or Codex, then re-materialize managed copies if this runtime is repo-managed                                                                                                                                                   |
| File not read                            | Not referenced                      | Add explicit read instruction in SKILL.md                                                                                                                                                                                                          |
| SKILL.md too long                        | Too much detail                     | Move to `references/` files                                                                                                                                                                                                                        |
| Skill works in Claude Code but not Codex | Wrong path or weak trigger wording  | Verify the Codex copy exists under `.agents/skills/<name>/SKILL.md` or `~/.agents/skills/<name>/SKILL.md`; `npx skills ls` can confirm managed installs but may omit manual maintainer symlinks. Then make the `description` more trigger-specific |
| Codex ignores skill                      | Description not specific enough     | Codex relies solely on description for auto-activation — add explicit trigger phrases                                                                                                                                                              |
| Skill works in Codex but not Claude Code | Not in Claude's skill path          | Verify `~/.claude/skills/<name>/SKILL.md` exists                                                                                                                                                                                                   |
