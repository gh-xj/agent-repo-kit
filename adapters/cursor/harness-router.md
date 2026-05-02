
# Harness Router

Route useful session learnings into durable, reviewable agent knowledge
surfaces without silently rewriting them.

## Scope

Use this skill when the task is to:

- Decide where new agent/harness knowledge should live after a session.
- Turn user corrections or repeated agent mistakes into an enhancement proposal.
- Propose updates across agent instruction files, skills, docs, work records,
  memory, hooks, CI, evals, or structured local artifacts.
- Review whether a learning is durable enough to promote from chat/session
  context into a versioned artifact.

Do not use this skill for:

- Directly editing a requested target when the destination is already clear.
- Generic research summarization without a persistence decision.
- Creating or refactoring a skill's content; use `skill-builder`.
- Operating `.work/`; use `work-cli`.
- Changing repo convention contracts; use `convention-engineering`.

MCP is out of scope as a durable destination for this skill.

## First Actions

1. Identify the session learning, correction, or proposed enhancement.
2. Gather only the nearby evidence needed to classify it:
   - user correction or session summary
   - changed files and tests
   - active work item or research space
   - relevant agent instruction files
   - related or used skills
3. Read `references/externalization-model.md`, then
   `references/routing-taxonomy.md`.
4. Produce a proposal before making durable edits.

## Operating Loop

1. Extract candidate learnings in compact form.
2. Classify the externalized burden and artifact class using
   `references/externalization-model.md`.
3. Classify authority, scope, durability, load policy, consumer, enforcement
   need, sensitivity, provenance, and cost.
4. Pick candidate target surfaces using `references/target-surfaces.md`.
5. Format recommendations with `references/proposal-format.md`.
6. Ask for approval before changing instructions, skills, hooks, CI, evals, or
   memory.

## Routing Rule

- Route by externalized burden first, then by durable surface.
- Required always-on repo behavior goes to the repo agent instructions; use
  harness adapter files only when runtime-specific guidance is needed.
- Repeatable procedures go to skills or skill reference files.
- Interaction contracts go to protocol owners such as CLI contracts, schemas,
  adapters, Taskfiles, generated-surface rules, or convention docs.
- Bulky durable knowledge goes to docs or research pages.
- Active follow-up goes to `.work` records and spaces.
- Personal low-authority context goes to memory.
- Hard guarantees go to hooks, settings, CI, lint, or evals.
- Temporary state stays in session context, compaction, sandbox/workspace files,
  or run logs unless it proves durable value.

## Output

Default output is a reviewable proposal:

- short summary of the durable lesson
- recommended destination(s)
- externalized burden and artifact class
- compact proposed change
- compact evidence and reasoning tied to routing dimensions
- confidence and risks

For multi-item proposals, include a structured block using the schema in
`references/proposal-format.md` only when another tool or agent will parse the
result. Use readable markdown sections for human review.

## Gaps

This is a proposal skill. It does not yet own a script, CLI, load verifier, or
automatic mutation workflow. Add code only after the proposal pattern has worked
across multiple real sessions.
