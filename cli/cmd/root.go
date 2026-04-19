// Package cmd wires the `ark` command-line surface.
//
// The CLI is declarative: the whole shape lives on the `CLI` struct below.
// Subcommand groups (`skill`, `taskfile`, `adapters`) are nested structs;
// leaf subcommands are fields tagged with `cmd:""` and implement
// `Run(globals *CLI) error`. See skills/go-scripting/references/kong-patterns.md
// for the style guide this file follows.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/log"
	"github.com/gh-xj/agent-repo-kit/cli/internal/skillsync"
)

const binaryName = "ark"

// Build-time metadata, overridden via -ldflags by the build toolchain.
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

// CLI is the root command surface. Global flags are top-level fields;
// subcommands and subcommand groups are fields tagged with `cmd:""`.
//
// Run methods read global flags directly (e.g. `globals.JSON`) and write
// output to `globals.stdout()` / `globals.stderr()` so tests can capture
// output by calling Run directly on a CLI with out/err pre-populated.
type CLI struct {
	Verbose bool   `short:"v" help:"enable debug logs"`
	NoColor bool   `name:"no-color" help:"disable colorized output"`
	JSON    bool   `help:"emit machine-readable JSON output"`
	Config  string `help:"path to contract checker config JSON (default: tracked .convention-engineering.json)"`

	Version     VersionCmd     `cmd:"" help:"print build metadata"`
	Init        InitCmd        `cmd:"" help:"scaffold tracked convention contract into a repo"`
	Check       CheckCmd       `cmd:"" help:"validate repo conventions against the tracked contract"`
	Orchestrate OrchestrateCmd `cmd:"" help:"run the convention orchestrator and launch the evaluator"`
	Upgrade     UpgradeCmd     `cmd:"" help:"upgrade the running ark binary in place"`
	Skill       SkillCmd       `cmd:"" help:"manage agent skills (init, audit, sync, check)"`
	Taskfile    TaskfileCmd    `cmd:"" help:"author and lint Taskfile.yml files"`
	Adapters    AdaptersCmd    `cmd:"" help:"manage harness adapter symlinks (link, list-links)"`

	// out and err are populated by Execute; unexported so kong doesn't
	// treat them as flags. Run methods call stdout()/stderr().
	out io.Writer
	err io.Writer
}

func (c *CLI) stdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}

func (c *CLI) stderr() io.Writer {
	if c.err != nil {
		return c.err
	}
	return os.Stderr
}

// Execute parses argv and dispatches to the selected subcommand's Run
// method. Exit codes:
//
//	0 — success
//	1 — runtime error from a Run method
//	2 — usage error from kong's parser (unknown flag, bad args, etc.)
//
// Subcommands override the runtime code by returning an
// *appctx.ExitCodeError (for example, `taskfile lint` uses 2 for YAML
// parse errors even though those surface from Run).
func Execute(args []string) int {
	return execWriters(args, os.Stdout, os.Stderr)
}

// execWriters is the testable core: it lets the test suite inject
// buffer writers without wrestling with os.Stdout redirection or
// subprocess orchestration.
func execWriters(args []string, stdout, stderr io.Writer) int {
	cli := CLI{out: stdout, err: stderr}
	parser, err := kong.New(&cli,
		kong.Name(binaryName),
		kong.Description("agent-repo-kit unified CLI"),
		kong.Writers(stdout, stderr),
		kong.Vars{
			"version":                    appVersion,
			"skillsync_manifest_default": skillsync.ManifestDefaultPath,
		},
	)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return appctx.ExitError
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		fmt.Fprintf(stderr, "ark: %s\n", err)
		return appctx.ExitUsage
	}
	log.Setup(log.Options{Verbose: cli.Verbose, NoColor: cli.NoColor, Writer: stderr})
	if err := ctx.Run(&cli); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(stderr, msg)
		}
		return appctx.ResolveExitCode(err)
	}
	return appctx.ExitSuccess
}
