# Skill Evals

Use lightweight eval cases when a skill's trigger or output needs evidence.
Start in markdown. Add code only after the cases are stable and repeated.

## When Evals Are Required

Add eval cases when the skill:

- changes files or runs commands
- handles untrusted input or public artifacts
- affects CI, releases, permissions, generated docs, adapters, or conventions
- exists because of a prior failure or user correction
- produces structured output consumed by another tool or skill
- has a broad or ambiguous trigger

For low-risk skills, a few trigger cases and a concrete output description are
enough.

## Trigger Cases

Use trigger cases to test the `description`.

```markdown
### Trigger Case - <name>

Prompt: <realistic user request>

Expected: trigger | no_trigger

Reason: <why the skill should or should not load>
```

Include both positive and negative cases. Negative cases prevent broad
descriptions from becoming attractive nuisance.

## Output Cases

Use output cases when the skill must produce a concrete artifact.

```markdown
### Output Case - <name>

Task: <realistic task>

Setup: <files, commands, repo state, or assumptions>

Expected Artifact: <file, proposal, command output, report, or patch>

Success Criteria:
- <observable requirement>
- <observable requirement>

Evidence To Inspect:
- <command, file, diff, log, or path>
```

## Regression Cases

Use regression cases when a skill change fixes a known failure.

```markdown
### Regression Case - <name>

Prior Failure: <what went wrong>

Expected Corrected Behavior: <what should happen now>

Verification: <how to check it>
```

## Review Rules

- Always include at least one `no_trigger` case for broad descriptions.
- High-risk skills need explicit approval points and evidence inspection.
- Do not optimize only for the visible cases; add held-out or slightly varied
  cases when the skill is important.
- If a script or CLI can fake success without preserving evidence, the eval is
  not strong enough.

