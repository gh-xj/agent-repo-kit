# Verification

A skill-local CLI is not done until an agent can prove it works from a cold
shell.

## Minimum Checks

Every CLI should have:

- `--help` for root and subcommands
- at least one read command smoke test
- at least one machine-output parse test
- tests for usage errors
- fixtures for representative inputs
- documented command examples in the parent skill
- a `cli/CONTRACT.md` or equivalent command contract when the CLI has mutating
  commands or structured output consumed by agents

For JSON:

```bash
tool show fixture-id --json > test/smoke/show.json
jq -e . test/smoke/show.json >/dev/null
```

For NDJSON:

```bash
tool stream --ndjson | jq -c . >/dev/null
```

## Schema Checks

For contracts consumed by agents or scripts, commit a JSON Schema under the CLI
or skill:

```text
cli/
  testdata/
    schema/
      show.v1.schema.json
```

Validate smoke output against the schema when the repo already has a validator.
If no validator exists, `jq -e` is an acceptable v0 parseability gate.

## Taskfile Wiring

Use `taskfile-authoring` for the actual Taskfile mechanics. The common shape is:

```yaml
smoke:
  desc: Smoke test the skill-local CLI
  cmds:
    - go run ./cli --help >/dev/null
    - go run ./cli show fixture --json | jq -e . >/dev/null
```

Keep wrappers thin. The Taskfile should run checks; it should not contain CLI
business logic.

## Contract File

For skill-local CLIs, prefer a `cli/CONTRACT.md` using
`references/templates/CLI_CONTRACT.md` as the starter. This file should be the
first place an agent reads after the parent `SKILL.md`.

It should answer:

- what commands exist
- which commands are read-only or mutating
- what JSON shapes are stable
- which exit codes matter
- which command proves the CLI works
- when to run dry-run before apply

## Fixture Strategy

Good fixtures are:

- small
- checked into the repo
- stable across time zones and machines
- scrubbed of secrets
- representative of real edge cases

Avoid tests that depend on current date, network state, user home directory, or
installed global config unless the command explicitly owns those concerns.

## Output Evals

Use output evals when the CLI contract is broad or critical:

- prompt: build or modify a CLI for a specific skill
- expected command examples
- expected safety classification
- expected JSON contract
- expected verification commands

Do not overbuild eval infrastructure for a v0 CLI. One realistic case is enough
to catch drift.

## Release Criteria

A skill-local CLI is ready when:

- the parent `SKILL.md` routes to it clearly
- read and write commands have explicit safety classes
- JSON output is parseable and versioned
- failures have stable messages or codes
- smoke tests run from a fresh checkout
- promotion triggers are documented
