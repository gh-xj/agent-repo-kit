---
name: go-scripting
description: "Use when writing a new Go CLI tool or script, bootstrapping a cobra subcommand tree, choosing a logger for a Go project, or deciding between stdlib loops and samber/lo collection helpers. Trigger on 'write a Go CLI', 'build a Go script', 'cobra subcommand', 'which Go logger', 'slog vs zerolog', 'samber lo', 'Go lodash', 'Go stdlib vs lo', 'scaffold Go tool'."
---

<!-- agent-repo-kit:skill-sync — do not edit; regenerate with `ark skill sync` -->

# Go Scripting

Opinionated guide for small Go CLIs and scripts. Prescribes a narrow, modern stack so you don't re-litigate the same choices on every new tool.

## When to use

- Writing a new Go CLI tool, even a one-file script.
- Adding a subcommand to an existing cobra app.
- Deciding which logger to use in a Go project.
- Deciding whether samber/lo earns its keep in your code.

## When NOT to use

- Go library code — this guide is for CLIs and scripts.
- Go web services — framework-specific guides fit better.
- Shell territory (one grep, one curl): write a shell script.

## Stack (at a glance)

| Layer              | Pick                                                           | Why                                               |
| ------------------ | -------------------------------------------------------------- | ------------------------------------------------- |
| CLI router         | spf13/cobra                                                    | De-facto standard; subcommand tree + flag parsing |
| Logger             | stdlib `slog` + lmittmann/tint                                 | Stdlib + tiny color handler; modern, swappable    |
| Collection helpers | stdlib loops by default; samber/lo only when it earns its keep | Avoid lodash-style sprawl                         |

See references for full rationale.

## Decision: Go vs shell

Go when you need **any** of:

- Structured config, JSON parsing, or typed values.
- Multiple subcommands.
- Cross-platform (no bash assumptions).
- Real error handling.
- Tests.

Shell when the whole thing is a sequence of ~5 commands a POSIX `sh` can run in 30 lines.

## Scale tiers

| Tier       | Shape                                          | Stack                                          |
| ---------- | ---------------------------------------------- | ---------------------------------------------- |
| **Script** | One `main.go`, ≤200 LOC, 0–1 flag              | stdlib `flag`; stdlib `slog` if you log at all |
| **Tool**   | Multiple subcommands, `cmd/` layout, ≤1000 LOC | cobra + slog+tint                              |
| **CLI**    | Production-grade, config, releases, tests      | cobra + slog+tint, optionally samber/lo        |

Skip cobra below Tier 2. Skip samber/lo below Tier 3 unless you've already written the same helper twice.

## Non-negotiables

- **Use slog.** Do not introduce a third-party logger in new tools. If you think you need zerolog, read `references/logging.md` first.
- **Ship `--verbose`, `--no-color`, and respect `NO_COLOR`.** Users expect these flags.
- **Exit codes:** `0` success, `1` runtime error, `2` usage error. Don't get cute.
- **Never `log.Fatal` or `os.Exit` from library code.** Exit from `main` only.
- **samber/lo is not a default.** Write the loop first. See `references/utilities.md`.

## Common mistakes

- Reaching for cobra on a 50-line script. Use stdlib `flag`.
- Using `fmt.Println` as your logger. Use slog.
- `panic` as error handling. Return errors; exit from `main`.
- Importing samber/lo for a single `lo.Filter`. Write the loop.
- Keeping timestamps in CLI output. Strip them — the shell prefix is enough.

## References

| File                           | Use For                                                      |
| ------------------------------ | ------------------------------------------------------------ |
| `references/logging.md`        | zerolog vs slog+tint pros/cons; Setup boilerplate; idioms    |
| `references/cobra-patterns.md` | Subcommand registry, global flags, exit codes, TTY detection |
| `references/utilities.md`      | When samber/lo earns its keep; stdlib alternatives           |
