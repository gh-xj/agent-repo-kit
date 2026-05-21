# Evals

Use these lightweight cases when changing this skill.

## Should Trigger

- "Build a CLI inside this skill so agents can run the repeated normalization
  workflow."
- "This skill has three scripts now; should we promote them into cli/?"
- "Design the JSON output contract for a skill-local command."
- "Add dry-run and yes flags to a destructive skill helper."
- "Create smoke tests for the CLI that a skill tells agents to run."

## Should Not Trigger

- "Write a Taskfile for this Go CLI." Use `taskfile-authoring`.
- "Create a new Go CLI with kong." Use `go-scripting`.
- "Should this learning become a skill or memory?" Use `harness-router` or
  `skill-builder`.
- "Run the existing work CLI to view ready tasks." Use `work-cli`.
- "Review this module architecture." Use `attack-architecture`.

## Output Success Criteria

For a realistic skill-local CLI design request, the answer should include:

- command examples before implementation details
- safety class for each command
- JSON or NDJSON output shape
- exit-code expectations
- verification commands
- parent `SKILL.md` routing text
- clear handoff to `go-scripting` or `taskfile-authoring` when appropriate

## First-Use Fixture Rule

Do not invent "real" eval fixtures before the skill has a real CLI trace. After
the first implementation that uses this skill, add one fixture capturing:

- the user prompt or task summary
- the command contract the agent proposed
- the safety classes selected
- the verification commands added
- one regression risk the eval should catch next time
