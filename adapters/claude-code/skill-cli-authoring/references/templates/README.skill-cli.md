# <tool-name>

Skill-local CLI for `<skill-name>`.

This CLI owns deterministic behavior used by the skill. The parent `SKILL.md`
owns when agents should run it and how to interpret the result.

## Common Commands

```bash
go run ./cli --help
go run ./cli show <id> --json
go run ./cli plan --input <path|-> --json
go run ./cli apply <plan-id> --dry-run --json
go run ./cli apply <plan-id> --yes --json
```

## Contract

Read `CONTRACT.md` before changing command names, JSON fields, exit codes, or
safety behavior.

## Verification

```bash
task -d cli verify
```

If this CLI is wired from the repo root, the root `Taskfile.yml` should delegate
to the CLI task instead of reimplementing command logic in shell.
