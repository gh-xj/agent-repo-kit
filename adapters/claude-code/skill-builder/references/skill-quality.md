# Skill Quality

Use this reference when creating, updating, or auditing a skill. A skill should
exist only when it externalizes procedural expertise that agents would
otherwise rediscover, forget, or execute inconsistently.

## Evidence Basis

Name the evidence behind the skill or change:

- `authored`: designed from known expert workflow.
- `distilled`: extracted from repeated successful traces.
- `corrected`: based on repeated agent failure or user correction.
- `discovered`: based on research, external standard, or upstream guidance.
- `composed`: assembled from existing skills or shared references.

Weak evidence does not block a skill, but it should lower confidence and keep
the `## Gaps` section honest.

## Risk Tier

Name the review tier before packaging the skill:

- `low`: read-only, personal, formatting, no external actions.
- `standard`: repo-local work, bounded edits, normal verification.
- `high`: shell commands, CI, releases, generated docs, adapters, public input,
  permissions, or broad repo mutation.
- `dangerous`: secrets, credentials, destructive actions, external publishing,
  or untrusted input plus privileged tools.

Tier effects:

- `low`: trigger check and clear output are enough.
- `standard`: trigger check plus output success criteria.
- `high`: explicit approval points, eval cases, verification command, and risk
  notes.
- `dangerous`: do not package as a skill without dedicated policy/checks and
  human approval.

## Expertise Check

Audit the skill for three parts:

- `procedure`: the agent can follow the workflow end to end.
- `heuristics`: the agent knows branch choices, defaults, tradeoffs, and when
  to escalate.
- `constraints`: the skill names boundaries, approval points, evidence
  standards, and unsafe actions.

If the skill has only declarations, turn them into a procedure or move them to
docs. If the task is deterministic and repeated, move the fragile part into a
script or CLI.

## Trigger Check

The `description` carries the trigger burden. Test it with realistic prompts:

- 3-5 prompts that should trigger the skill.
- 3-5 prompts that should not trigger it.

Watch for two failures:

- too narrow: the agent misses relevant tasks
- too broad: the agent loads the skill for ordinary work it can already handle

Use `skill-evals.md` when cases should be preserved.

## Output Check

Use output evals when the skill affects important behavior, writes files, runs
commands, or has caused regressions.

Start small:

- one realistic prompt
- expected output or success criteria
- any required input files
- one comparison: without the skill, previous skill version, or current skill

Do not require eval files for every small skill. Require them when the behavior
is important enough that "looks right" is not a good standard.

## Eval-Required Triggers

Preserve eval cases when the skill:

- changes files or runs commands
- handles untrusted input or public artifacts
- affects CI, releases, permissions, generated docs, adapters, or conventions
- exists because of a prior failure or user correction
- produces structured output consumed by another tool or skill
- has a broad or ambiguous trigger

## Composition Check

Small skills compose only when ownership is clear:

- one skill owns each durable rule or concept
- dependencies are one-way
- shared vocabulary has a canonical owner
- two skills should not both mutate the same target surface
- a skill should say when another skill owns the task

## Retirement Signals

Refactor, merge, split, or delete a skill when:

- its trigger has become too broad to test
- it only stores context that belongs in docs, work, or memory
- deterministic behavior has moved into a CLI and the skill no longer adds
  judgment
- another skill owns the same workflow more clearly
- references are stale and no current workflow depends on them
