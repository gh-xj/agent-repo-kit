# Workflows

Use this file for the concrete create, update, audit, and migrate workflows.

## Confidence Check

| Signal                                        | Action                     |
| --------------------------------------------- | -------------------------- |
| 3+ solid examples or one deep complex session | Build the full skill       |
| 1-2 examples with open edge cases             | Ship `v0.x` with `## Gaps` |
| Pattern just emerged                          | Capture notes only         |

If you cannot name the edge cases yet, do not fake confidence.

## Build Approach

| Approach         | Use When                                        |
| ---------------- | ----------------------------------------------- |
| Direct build     | The builder already understands the domain well |
| Q&A-driven build | The user holds important tacit knowledge        |

Q&A-driven build:

1. Draft the smallest honest skill.
2. List what you are not confident about.
3. Ask targeted questions.
4. Update the draft.
5. Repeat until the core path is solid.

## Create

1. Define the problem, boundary, and success signal.
2. Write the trigger-oriented `description`.
3. Decide the operating surface:
   - `SKILL.md` only
   - extracted references
   - script
   - skill-local CLI in `cli/`
   - repo-owned CLI
4. Draft `SKILL.md` as a router.
5. Add only the references needed for the request.
6. Validate activation with 3-5 should-trigger and 3-5 should-not-trigger prompts.
7. If the skill owns a stable command surface, bootstrap `cli/` with your repo's Go scaffolder and keep root Task wrappers thin.

## Update

1. Read the existing skill first.
2. Find duplication, contradictions, or oversized sections.
3. Consolidate into one source of truth.
4. Extract config, formats, orchestration, or domain knowledge into references.
5. Re-test activation and references.

## Audit

Focus on findings first:

- trigger too broad or too narrow
- portable core polluted by runtime- or repo-specific rules
- duplicated facts
- `SKILL.md` too long to function as a router
- deterministic logic left in prose after the pattern is already stable

Then propose the smallest full refactor.

## Migrate

Use this when capability is moving between README, docs, skills, scripts, Taskfiles, or repo CLIs.

Rules:

1. Move the owning logic to the new surface.
2. Remove stale entrypoints in the same change.
3. Sweep `README*`, `docs/**`, `.github/workflows/**`, and `Taskfile*` for old names.
4. Name the verification gate explicitly.

## Refactor Thresholds

| Signal                                              | Action                |
| --------------------------------------------------- | --------------------- |
| `SKILL.md` > 200 lines                              | consider extraction   |
| `SKILL.md` > 400 lines                              | extract now           |
| Config mixed with workflow                          | move to `references/` |
| Stable deterministic steps re-derived every session | move to code          |

## Validation Checklist

- Test trigger quality.
- Check that every referenced file exists.
- Verify the workflow can be followed end to end.
- Ensure the portable core still uses `name` + `description`.
- If both Claude and Codex copies exist, verify parity after the update.
