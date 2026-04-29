package arkcli

import (
	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

// CheckCmd validates repo conventions against the tracked contract.
type CheckCmd struct {
	RepoRoot string `name:"repo-root" help:"path to the repository root" default:"."`
}

func (c *CheckCmd) Run(globals *CLI) error {
	// An empty globals.Config means "use the tracked contract at
	// contract.DefaultConfigFile"; contract.Run handles the resolution.
	exitCode := contract.Run(c.RepoRoot, globals.Config, globals.JSON, globals.stdout(), globals.stderr())
	if exitCode == 0 {
		return nil
	}
	// contract.Run has already written output to stdout/stderr; surface
	// the exact exit code via appctx.NewExitError with an empty message
	// to avoid a double-printed summary line.
	return appctx.NewExitError(exitCode, "")
}
