# Core Principles

Skill-local CLIs exist because agents are good at using stable tools and bad at
remembering fragile ritual. A skill should keep judgment in prose and move
deterministic, repeated behavior into code.

## The Product Stance

Build a boring CLI:

- deterministic input and output
- explicit command grammar
- stable JSON fields
- no hidden prompts in machine paths
- no ambient mutation from read commands
- enough output to verify what happened

Do not build a conversational assistant inside the CLI. The agent is already the
conversation layer. The CLI should be the reliable tool layer.

## What Makes A CLI Agent-Friendly

Agent-friendly means an agent can:

1. discover the command from the skill
2. run it from a clean shell
3. parse the output without terminal heuristics
4. recover from errors using documented exit codes
5. retry safe operations without damage
6. prove success with returned state or fixture output

Human-friendly and agent-friendly are not opposites. Human output can be compact
and readable. Machine output must be exact.

## Good Defaults

- Prefer `--json` for complete result objects.
- Prefer `--ndjson` for streams or long-running progress.
- Write primary output to stdout and diagnostics to stderr.
- Suppress color, tables, spinners, and progress bars in machine modes.
- Include `schema_version` in structured output.
- Include stable IDs and paths, not only prose summaries.
- Return the post-operation state for mutations.
- Make no-op success explicit: `changed: false`.
- Use absolute or repo-relative paths consistently.
- Keep timestamps in structured fields, not human prose.

## What Not To Automate

Do not encode judgment that belongs to the agent or user:

- whether a feature is worth building
- whether a destructive operation should proceed
- whether a policy exception is acceptable
- which external account, repo, or workspace should be trusted

The CLI can validate facts and prepare plans. The agent or user decides.

## Relationship To MCP

A skill-local CLI can become the implementation behind an MCP server later.
Designing commands with typed inputs, typed outputs, idempotency, and safety
classification maps cleanly to MCP tool schemas and annotations.

Do not start with MCP unless discovery by external agent hosts is already a
requirement. A local CLI is easier to test, inspect, and compose.
