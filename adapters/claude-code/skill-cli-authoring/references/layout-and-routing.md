# Layout And Routing

Choose the smallest owning surface that gives the agent a reliable contract.

## Decision Table

| Surface           | Use When                                                        |
| ----------------- | --------------------------------------------------------------- |
| Skill prose       | Judgment-heavy workflow; no deterministic repeated operation    |
| `scripts/`        | One narrow operation, few flags, no stable public contract       |
| `cli/`            | Multiple commands, structured IO, tests, or repeated operations |
| Repo `tools/<x>/` | Shared outside one skill, release-worthy, or cross-repo policy  |
| External CLI      | Domain already has a maintained tool                            |

## Skill-Local CLI Layout

Use this shape for a CLI owned by one skill:

```text
skill-name/
  SKILL.md
  cli/
    README.md
    go.mod
    cmd/<tool>/main.go
    internal/...
    testdata/...
  references/
    command-contract.md
    verification.md
```

For a very small Go tool, a flatter shape is acceptable:

```text
skill-name/
  SKILL.md
  cli/
    main.go
    go.mod
    testdata/...
```

Promote to a repo-owned CLI when another skill starts importing its concepts or
when users need to call it without loading the owning skill.

## Parent Skill Routing

`SKILL.md` should include:

- when to run the CLI
- the command examples agents should prefer
- which commands are read-only
- which commands require dry-run first
- what output proves success
- where the CLI source lives

Do not paste every flag into `SKILL.md`. Keep detailed command docs in
`cli/README.md` or a reference file.

## Naming

Prefer domain-specific names over generic names:

- Good: `paper-vet`, `work-audit`, `skill-sync`
- Weak: `helper`, `runner`, `agent-tool`

For skill-local commands that are not installed globally, document invocation
relative to the skill root:

```bash
go run ./cli --json ...
```

If installed on PATH, document the binary name and a setup check:

```bash
skill-tool version --json
```

## Ownership Smells

Move out of a skill-local CLI when:

- multiple skills need it
- it reads or mutates repo-wide policy
- it needs releases, packaging, or user-facing docs
- its behavior is no longer understandable from the owning skill

Move back to prose or scripts when:

- every command is a wrapper around one shell command
- the CLI has no tests and no stable output contract
- the operation still requires human judgment at every branch
