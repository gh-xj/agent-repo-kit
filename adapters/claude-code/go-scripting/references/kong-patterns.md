# Kong Patterns

## Why kong

[`alecthomas/kong`](https://github.com/alecthomas/kong) parses argv into a Go
struct via struct tags — the same mental model Go developers already use for
JSON, YAML, env-var binding, and SQL scanning.

- **Declarative.** The whole CLI shape is visible in one type.
- **No registry plumbing.** Subcommands are struct fields, not `init()`
  side effects.
- **Typed.** Flags and args become typed struct fields, not `string` blobs
  parsed per-subcommand.
- **Extensible.** Custom decoders, validators, completions, and a plugin
  hook for dynamic subcommands.

## When kong is worth it

- 2+ subcommands, or
- ≥5 flags, or
- You want auto-generated `--help`, shell completion, and typed flag parsing
  without ceremony.

For a one-file script with 1–2 flags, stdlib `flag` is still enough.

## The declarative shape

```go
package main

import (
	"github.com/alecthomas/kong"
)

// CLI is the root shape. Global flags live at the top; subcommands are
// fields tagged with `cmd:""`.
type CLI struct {
	Verbose bool   `short:"v" help:"enable debug logs"`
	NoColor bool   `name:"no-color" help:"disable colored output"`
	JSON    bool   `help:"emit machine-readable JSON output"`
	Config  string `help:"config file path"`

	Build  BuildCmd  `cmd:"" help:"build a target"`
	Deploy DeployCmd `cmd:"" help:"deploy to an environment"`
}

type BuildCmd struct {
	Target string `arg:"" help:"target to build"`
	Clean  bool   `help:"clean before building"`
}

func (c *BuildCmd) Run(globals *CLI) error {
	// ...
	return nil
}

type DeployCmd struct {
	Env string `arg:"" help:"environment name" enum:"dev,staging,prod"`
}

func (c *DeployCmd) Run(globals *CLI) error {
	// ...
	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("mytool"),
		kong.Description("opinionated example"),
	)
	if err := ctx.Run(&cli); err != nil {
		ctx.FatalIfErrorf(err)
	}
}
```

Run `mytool build myservice --clean` and kong populates `CLI.Build.Target = "myservice"`, `CLI.Build.Clean = true`, then calls `(&cli.Build).Run(&cli)`.

## Nested subcommand groups

Some CLIs have groups like `git remote add` or `kubectl config set`. Model
a group as a struct whose fields are leaf-level subcommands:

```go
type CLI struct {
	Skill SkillCmd `cmd:"" help:"manage agent skills"`
}

type SkillCmd struct {
	Init  SkillInitCmd  `cmd:"" help:"scaffold a new skill"`
	Audit SkillAuditCmd `cmd:"" help:"audit a skill"`
	Sync  SkillSyncCmd  `cmd:"" help:"render adapter copies"`
}

type SkillInitCmd struct {
	Skill string `help:"skill id" required:""`
	Dir   string `help:"output dir" default:"."`
}

func (c *SkillInitCmd) Run(globals *CLI) error { /* ... */ }
```

`mytool skill init --skill foo` routes correctly. Each leaf is a file.

## Global flags every CLI ships

| Flag         | Short | Tag                                           |
| ------------ | ----- | --------------------------------------------- |
| `--verbose`  | `-v`  | `Verbose bool \`short:"v" help:"..."\``       |
| `--no-color` |       | `NoColor bool \`name:"no-color" help:"..."\`` |
| `--json`     |       | `JSON bool \`help:"..."\``                    |
| `--config`   |       | `Config string \`help:"..."\``                |
| `--dry-run`  |       | `DryRun bool \`name:"dry-run" help:"..."\``   |

Declare them once on `CLI`. Every `Run(globals *CLI)` can read them.

## Flag tags cheat sheet

| Tag                                            | Meaning                                                                |
| ---------------------------------------------- | ---------------------------------------------------------------------- |
| `arg:""`                                       | Positional argument (required by default; add `optional:""` to loosen) |
| `cmd:""`                                       | Subcommand                                                             |
| `default:"x"`                                  | Default value                                                          |
| `enum:"a,b,c"`                                 | Restrict to a small set                                                |
| `env:"MY_VAR"`                                 | Fall back to env var                                                   |
| `required:""`                                  | Fail parse if not provided                                             |
| `short:"v"`                                    | Short form                                                             |
| `name:"kebab-name"`                            | Override the kebab-case flag name                                      |
| `help:"..."`                                   | Help text (keep it short)                                              |
| `hidden:""`                                    | Omit from help                                                         |
| `type:"path"` / `existingfile` / `existingdir` | Built-in decoders                                                      |

## Exit code policy

- `0` — success
- `1` — runtime error
- `2` — usage error

kong emits exit `1` via `FatalIfErrorf` by default. For a strict usage/
runtime split, wrap `Parse`:

```go
var cli CLI
parser := kong.Must(&cli, kong.Name("mytool"))
ctx, err := parser.Parse(os.Args[1:])
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	return 2  // usage error from kong's Parse
}
if err := ctx.Run(&cli); err != nil {
	fmt.Fprintln(os.Stderr, err)
	return 1  // runtime error from a Run method
}
return 0
```

This gives you a deterministic usage-vs-runtime distinction, matching the
cobra-style resolver used in older tools.

## Versioning and testable exits

`kong.VersionFlag` (and `kong.HelpFlag`) print their output and call
`os.Exit` directly inside `Parse`, bypassing any `Execute(args) int`
return-code contract. In tests this looks like the test process abruptly
exits mid-run with kong frames in the stack trace.

Fix: thread `kong.Exit(...)` so kong-managed exits flow back through your
runner's normal int-return path:

```go
exitRequested := false
exitCode := 0
parser, err := kong.New(&cli,
    kong.Name("mytool"),
    kong.Vars(map[string]string{"version": appVersion}),
    kong.Exit(func(code int) {
        exitRequested = true
        exitCode = code
    }),
)
// ...
ctx, err := parser.Parse(args)
if exitRequested {
    return exitCode  // check BEFORE the err branch
}
if err != nil {
    return usageExitCode
}
```

Order matters. The version hook short-circuits Parse, but Parse still
returns an `expected one of <subcommands>` error because no subcommand
was given. Honor `exitRequested` first; treat the parse error as fatal
only when no exit was requested.

A root `--version` flag and a `version` subcommand can coexist as long as
the Go field names differ:

```go
type CLI struct {
    VersionFlag kong.VersionFlag `name:"version" help:"print version and exit"`
    Version     VersionCmd       `cmd:"" help:"print build metadata"`
}
```

The flag handles `mytool --version` (single line, hits kong's exit hook);
the subcommand handles `mytool version --json` and any richer output your
tooling needs.

## TTY detection

Same as with any CLI:

```go
import "golang.org/x/term"

if term.IsTerminal(int(os.Stdout.Fd())) {
	// interactive: colors, prompts, progress
}
```

Use this when deciding color, interactive wizards, progress bars.

## Suggested layout

```
myproject/
├── go.mod
├── main.go                 # one-liner: os.Exit(cmd.Execute(os.Args[1:]))
└── cmd/
    ├── root.go             # type CLI struct { ... }; Execute(args []string) int
    ├── build.go            # type BuildCmd struct + Run
    ├── deploy.go           # type DeployCmd struct + Run
    └── internal/
        ├── log/log.go      # slog+tint Setup
        └── <features>
```

Each subcommand is a single file: the struct, its flags, and its `Run`.
`root.go` holds the global flags, the top-level `CLI`, and `Execute`.

## Anti-patterns

- **`os.Exit` inside a `Run` method.** Return an error; let `Execute`
  resolve the code.
- **Reading `os.Args` inside a subcommand.** The Run method already has
  everything parsed on its receiver.
- **Side-effectful `init()` trying to register subcommands.** In kong the
  CLI shape is explicit; don't reintroduce cobra-style distributed
  registration.
- **A flag on `CLI` that's only used by one subcommand.** Put it on the
  subcommand struct instead.
- **Over-nesting.** `mytool group subgroup subsub command` rarely beats
  `mytool command --group X --subgroup Y`.

## When cobra still makes sense

- Codebases already deeply committed to cobra.
- You rely on cobra-specific tooling: `cobra-cli gen`, cobra docs
  generators, hooks that expect `*cobra.Command`.
- Readers on the team are cobra-fluent and kong is a step they don't want.

For new Go CLIs written without that inheritance: **pick kong.**
