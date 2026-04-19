# Cobra Patterns

## When cobra is worth it

Reach for cobra when you have:

- 2+ subcommands, or
- ≥5 flags, or
- You want auto-generated `--help`, shell completion, or hierarchical command docs.

A one-file script with 1–2 flags: use stdlib `flag`. Cobra is overhead for that shape.

## Subcommand registry pattern

Instead of scattering `AddCommand` calls across files, register via `init()`:

```go
// cli/cmd/foo.go
func init() {
	registerCommand("foo", fooCommand())
}

func fooCommand() command {
	return command{
		Description: "do the foo thing",
		Configure: func(c *cobra.Command) {
			c.Flags().String("target", "", "target name")
		},
		Run: func(app *appctx.AppContext, c *cobra.Command, args []string) error {
			// ...
		},
	}
}
```

And a tiny `command` struct + registry in `root.go`:

```go
type command struct {
	Description string
	Configure   func(*cobra.Command)
	Run         func(*appctx.AppContext, *cobra.Command, []string) error
}

var commandRegistry = map[string]command{}

func registerCommand(name string, c command) { commandRegistry[name] = c }

func newRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "mytool", SilenceUsage: true}
	for _, name := range sortedKeys(commandRegistry) {
		root.AddCommand(newLeafCmd(name, commandRegistry[name]))
	}
	return root
}
```

This gives each subcommand a single-file home and keeps `root.go` stable as you add commands.

## Global flags every CLI ships

| Flag         | Short | Purpose                                      |
| ------------ | ----- | -------------------------------------------- |
| `--verbose`  | `-v`  | Lift log level to Debug                      |
| `--no-color` |       | Force plain output (overrides TTY detection) |
| `--json`     |       | Emit machine-readable output                 |
| `--config`   |       | Override default config file path            |
| `--dry-run`  |       | Preview without side effects                 |

Declare them as `root.PersistentFlags()` so every subcommand inherits.

## Exit code policy

- `0` — success
- `1` — runtime error (ordinary failure the user needs to fix)
- `2` — usage error (unknown flag, bad arg count, missing required, etc.)

Cobra emits `2` for usage errors by default — detect and preserve. A helper:

```go
func resolveCode(err error) int {
	if err == nil { return 0 }
	text := strings.ToLower(err.Error())
	for _, prefix := range []string{
		"unknown command",
		"unknown flag",
		"invalid argument",
		"flag needs an argument",
		"accepts ",
		"requires at ",
	} {
		if strings.HasPrefix(text, prefix) {
			return 2
		}
	}
	return 1
}
```

## TTY detection

Use `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))` to decide:

- Whether to show colors.
- Whether to run an interactive wizard or fall back to `--non-interactive`.
- Whether to emit progress bars vs plain logs.

## Suggested layout

```
myproject/
├── go.mod
├── main.go                      # one-liner: os.Exit(cmd.Execute(os.Args[1:]))
└── cmd/
    ├── root.go                  # root cobra.Command + global flags + command registry
    ├── foo.go                   # one subcommand per file
    ├── bar.go
    └── internal/
        ├── log/log.go           # slog+tint Setup (see references/logging.md)
        └── <feature-packages>
```

Subcommand files self-register via `init()`. `root.go` never changes when you add a subcommand.

## Anti-patterns

- **Command definitions in `main.go`.** `main.go` should be one line: call `cmd.Execute`.
- **Parsing flags in multiple places.** Flags belong in `Configure`; read them from `cmd.Flags()` inside `Run`.
- **Reading `os.Args` inside subcommands.** Take args from the `Run` callback's `args []string`.
- **`log.Fatal` inside a `Run` function.** Return an `error` so cobra surfaces the exit code consistently.
- **Over-nesting.** `mytool group subgroup command` is rarely worth the indirection over `mytool command --group X`.
