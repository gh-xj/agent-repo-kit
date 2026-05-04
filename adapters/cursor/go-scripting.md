---
name: go-scripting
description: "Use when writing a new Go CLI tool or script, declaring a kong-based subcommand tree, choosing a logger (slog+tint), working with collections via stdlib iter or slices, or adding tables and spinners to CLI output. Trigger on 'write a Go CLI', 'build a Go script', 'kong subcommand', 'cobra vs kong', 'which Go logger', 'slog vs zerolog', 'Go iter', 'iter.Seq', 'samber lo', 'Go stdlib vs lo', 'scaffold Go tool', 'Go CLI table', 'Go spinner', 'go-pretty'."
---

# Go Scripting

Opinionated guide for small Go CLIs and scripts. Prescribes a narrow,
modern stack so you don't re-litigate the same choices on every new tool.

## When to use

- Writing a new Go CLI tool, even a one-file script.
- Adding a subcommand to an existing kong-based CLI.
- Deciding which logger, CLI framework, or collection helper to use.
- Adding tables, spinners, or progress output.

## When NOT to use

- Go library code — this guide is for CLIs and scripts.
- Go web services — framework-specific guides fit better.
- Shell territory (one grep, one curl): write a shell script.

## Stack (at a glance)

| Layer               | Pick                                                 | Why                                                                  |
| ------------------- | ---------------------------------------------------- | -------------------------------------------------------------------- |
| CLI router          | alecthomas/kong                                      | Declarative struct-tag parsing; the whole CLI is visible in one type |
| Logger              | stdlib `slog` + lmittmann/tint                       | Stdlib + tiny color handler; modern, swappable                       |
| Collections         | stdlib `slices` / `maps` / `cmp` + `iter` (Go 1.23+) | Stdlib + range-over-func covers most lodash territory                |
| Tables              | jedib0t/go-pretty                                    | Rich, actively maintained, Unicode-aware                             |
| Spinner             | charmbracelet/huh/spinner or briandowns/spinner      | Pick one per tool; don't mix                                         |
| Last-resort helpers | samber/lo                                            | Only for specific helpers stdlib+iter can't provide                  |

See references for the full rationale.

## Decision: Go vs shell

Go when you need **any** of:

- Structured config, JSON parsing, or typed values.
- Multiple subcommands.
- Cross-platform (no bash assumptions).
- Real error handling.
- Tests.

Shell when the whole thing is a sequence of ~5 commands a POSIX `sh` can
run in 30 lines.

## Scale tiers

| Tier       | Shape                                          | Stack                                            |
| ---------- | ---------------------------------------------- | ------------------------------------------------ |
| **Script** | One `main.go`, ≤200 LOC, 0–1 flag              | stdlib `flag`; stdlib `slog` if you log at all   |
| **Tool**   | Multiple subcommands, `cmd/` layout, ≤1000 LOC | kong + slog+tint                                 |
| **CLI**    | Production-grade, config, releases, tests      | kong + slog+tint, optionally go-pretty + spinner |

Skip kong below Tier 2 — stdlib `flag` is enough for a 2-flag script.
Skip go-pretty/spinner unless you have row-oriented data or long-running
operations.

## Non-negotiables

- **Use slog.** Don't introduce a third-party logger. If you think you
  need zerolog, read `references/logging.md`.
- **Use kong for new tools.** cobra is acceptable only when the codebase
  is already on it. Mixing routers across one project is worse than
  either.
- **Ship `--verbose`, `--no-color`, and respect `NO_COLOR`.** Users
  expect these flags.
- **Exit codes:** `0` success, `1` runtime error, `2` usage error.
- **Never `log.Fatal` or `os.Exit` from library code.** Exit from main
  only.
- **Stdlib collections first.** `slices.Contains` before `lo.Contains`.
  `iter.Seq` before `lo.Map` chains. samber/lo is a last resort.
- **Gate all visual output (colors, spinners, tables) on
  `term.IsTerminal && NO_COLOR == ""`.** Piped or CI output must be
  plain.

## Common mistakes

- Reaching for kong on a 50-line script. Use stdlib `flag`.
- Using `fmt.Println` as your logger. Use slog.
- `panic` as error handling. Return errors; exit from main.
- Importing samber/lo for a single Filter. Use stdlib slices or a loop.
- Materializing a slice when `iter.Seq` suffices. Return the iterator.
- Spinners in CI logs. Detect the terminal and suppress.
- Emitting a table when `--json` was passed. Respect the flag.
- Keeping timestamps in CLI output. Strip them — the shell prefix is
  enough.

## References

| File                            | Use For                                                                     |
| ------------------------------- | --------------------------------------------------------------------------- |
| `references/logging.md`         | zerolog vs slog+tint tradeoffs; Setup boilerplate; idioms                   |
| `references/kong-patterns.md`   | Declarative struct shape, nested subcommands, global flags, exit codes, TTY |
| `references/collections.md`     | stdlib slices/maps/cmp; Go 1.23 iter; when samber/lo earns its keep         |
| `references/progress-output.md` | Tables (go-pretty), spinners, progress bars, TTY gating                     |
