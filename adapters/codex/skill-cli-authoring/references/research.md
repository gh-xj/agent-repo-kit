# Research

These references inform the skill's defaults. Use them as lenses, not rules to
copy blindly.

## CLI Design

- Command Line Interface Guidelines: human-first CLIs can still be composable;
  stdout/stderr, exit codes, help, dry-runs, and JSON matter.
  <https://clig.dev/>
- Heroku CLI Style Guide: offer machine-readable formats when valuable and
  avoid breaking stdout contracts after general availability.
  <https://devcenter.heroku.com/articles/cli-style-guide>
- GitHub CLI formatting: `--json`, `--jq`, and templates make command output
  reusable in scripts.
  <https://cli.github.com/manual/gh_help_formatting>

## Agent Tooling

- OpenAI Codex use case, "Create a CLI Codex can use": emphasizes predictable
  JSON, exact reads by ID, paged search, downloaded files, local indexes, and
  draft-before-write commands.
  <https://developers.openai.com/codex/use-cases/agent-friendly-clis>
- OpenAI function calling: tool inputs are defined by JSON schema and tool
  outputs are fed back into the model.
  <https://developers.openai.com/api/docs/guides/function-calling>
- OpenAI structured outputs: JSON Schema-backed outputs can be parsed and
  validated as typed data.
  <https://developers.openai.com/api/docs/guides/structured-outputs>
- Model Context Protocol tool spec: tools can expose input schemas, output
  schemas, structured content, and annotations for read-only, destructive, and
  idempotent behavior.
  <https://modelcontextprotocol.io/specification/2025-06-18/server/tools>

## Agent Instructions

- AGENTS.md: repository-local instructions give agents a predictable place to
  learn setup, testing, and command usage.
  <https://github.com/agentsmd/agents.md>

## Practical Interpretation

For skill-local CLIs, the strongest common pattern is:

1. keep the human command surface small
2. make every machine contract explicit
3. expose safe read commands before write commands
4. make writes plan-first when possible
5. validate output in smoke tests
6. teach agents the canonical commands in the parent skill
