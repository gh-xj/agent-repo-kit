---
name: skill-cli-authoring
description: "Use when creating, refactoring, or promoting a CLI that lives inside an agent skill, especially a skill-local cli/ tool with agent-friendly JSON output, safe write commands, dry-run behavior, exit codes, Taskfile smoke tests, or a companion SKILL.md routing contract. Triggers on 'build a CLI inside a skill', 'skill-local CLI', 'agent-friendly CLI', 'add cli/ to a skill', 'promote skill script to CLI', 'JSON output contract for a skill tool', or 'CLI for agents to use from a skill'."
---

# Skill CLI Authoring

> **Canonical source**: `skill-cli-authoring/SKILL.md` in the
> `agent-repo-kit` repository. All `references/...` links below resolve from
> that path. If you are reading this in an adapter mirror
> (`adapters/claude-code/`, `adapters/codex/`, `adapters/cursor/`), open the
> canonical file to reach the references.

Design and build small CLIs that live inside skills. The goal is not a clever
terminal product; it is a deterministic tool surface an agent can discover,
run, parse, retry, and verify.

## When To Use

Use this skill when:

- A skill repeatedly asks agents to perform deterministic shell steps.
- A skill needs a local `cli/` tool instead of fragile prose or ad hoc scripts.
- A skill-local script has grown into multiple commands, flags, or data shapes.
- A command contract needs `--json`, `--ndjson`, exit codes, or output schemas.
- A mutating skill operation needs `--dry-run`, `--yes`, idempotency, or audit
  output.
- A skill's `SKILL.md` needs to route agents to a companion CLI.

Do not use for:

- General Go CLI implementation details — use `go-scripting`.
- Taskfile mechanics, build caching, or `task ci` design — use
  `taskfile-authoring`.
- Deciding whether knowledge belongs in a skill at all — use `skill-builder`
  or `harness-router`.
- One-off shell glue under ~30 lines with no stable contract — keep it as a
  script.

## Core Principles

1. **CLI only after repetition.** Build a CLI when the same deterministic
   operation is being re-derived, mis-run, or copied across sessions.
2. **Contracts beat prose.** Stable flags, JSON fields, exit codes, and smoke
   checks are easier for agents to honor than long instructions.
3. **Reads and writes stay distinct.** Read commands should be safe, parseable,
   and composable. Write commands should be explicit, auditable, and guarded.
4. **Human text is secondary.** Plain output can be friendly, but machine output
   is the durable contract.
5. **Skill prose routes, CLI executes.** `SKILL.md` should say when and why to
   use the CLI. The CLI should own deterministic behavior.

## Quality Notes

- **Evidence basis:** composed from `skill-builder`, `go-scripting`,
  `taskfile-authoring`, current agent-tooling docs, and repeated local need for
  skill-owned CLIs.
- **Risk tier:** high. This skill designs tools that may run shell commands or
  mutate files, so command contracts, safety classes, and verification are
  required before implementation.
- **Procedural burden:** medium. Keep judgment in references and deterministic
  behavior in the CLI once the pattern stabilizes.

## Workflow

1. **Classify the extraction.** Choose `script`, `skill-local CLI`, or
   `repo-owned CLI` using `references/layout-and-routing.md`.
2. **Define the command contract.** Write concrete examples first; then specify
   flags, JSON/NDJSON shape, exit codes, and no-op behavior. Use
   `references/command-contract.md`.
3. **Design safety for writes.** Mark each command read-only, additive,
   idempotent, destructive, or open-world. Add `--dry-run`, `--yes`, or
   explicit target flags where needed. Use `references/safety-and-mutations.md`.
4. **Choose implementation stack.** For Go CLIs, hand off implementation style
   to `go-scripting`. Keep this skill focused on the agent-facing contract.
5. **Wire verification.** Add smoke tests, fixtures, schema checks, and Taskfile
   wrappers using `references/verification.md` and `taskfile-authoring`.
6. **Update the parent skill.** Keep `SKILL.md` as the router: command examples,
   when to run the CLI, and what output proves success.

## Default Command Shape

Prefer subcommands with explicit nouns and verbs:

```bash
skill-tool search --query "..." --json
skill-tool show ITEM-ID --json
skill-tool plan --source input.json --json
skill-tool apply PLAN-ID --dry-run --json
skill-tool apply PLAN-ID --yes --json
```

For machine output:

```json
{
  "schema_version": "v1",
  "ok": true,
  "command": "apply",
  "changed": false,
  "result": {}
}
```

## References

| File                                      | Use For                                                |
| ----------------------------------------- | ------------------------------------------------------ |
| `references/core-principles.md`           | Product stance and AI-era CLI principles               |
| `references/layout-and-routing.md`        | Script vs skill-local CLI vs repo-owned CLI decisions  |
| `references/command-contract.md`          | Command grammar, JSON/NDJSON, errors, exit codes       |
| `references/safety-and-mutations.md`      | Dry-run, destructive gates, idempotency, audit output   |
| `references/verification.md`              | Smoke tests, schemas, fixtures, Taskfile wiring        |
| `references/research.md`                  | External references and comparable project patterns    |
| `references/evals.md`                     | Trigger and output eval cases for maintaining the skill |
| `references/templates/CLI_CONTRACT.md`    | Starter contract for a skill-local `cli/CONTRACT.md`   |
| `references/templates/README.skill-cli.md` | Starter `cli/README.md` for agent-facing commands      |
| `references/templates/Taskfile.skill-cli-smoke.yml` | Thin smoke-test Taskfile fragment              |

## Output Shape

When shaping a skill-local CLI, return:

- command examples
- contract decisions
- safety model
- layout and ownership
- verification plan
- parent `SKILL.md` routing text
- open gaps or promotion triggers

## Boundaries

- Do not hide complex workflow policy inside a CLI just because it is easier
  than writing a clear skill.
- Do not make the CLI depend on a hosted service unless that is the domain it
  wraps and offline failure is explicit.
- Do not duplicate implementation boilerplate from `go-scripting`; reference it.
- Do not duplicate Taskfile recipes from `taskfile-authoring`; reference it.
