package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

func init() {
	registerCommand("check", CheckCommand())
}

// CheckCommand wires the contract.Run entry point as the top-level
// `ark check` command.
func CheckCommand() command {
	return command{
		Description: "validate repo conventions against the tracked contract",
		Configure: func(command *cobra.Command) {
			command.Flags().String("repo-root", ".", "path to the repository root")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			repoRoot, _ := command.Flags().GetString("repo-root")
			// The global --config persistent flag (default "") is plumbed
			// into app.Values["config"] by the root RunE wrapper. Empty
			// means "use the tracked contract at DefaultConfigFile".
			configPath, _ := app.Values["config"].(string)
			jsonMode, _ := app.Values["json"].(bool)

			exitCode := contract.Run(repoRoot, configPath, jsonMode, os.Stdout, os.Stderr)
			if exitCode == 0 {
				return nil
			}
			// contract.Run has already written output to stdout/stderr;
			// surface the exact exit code via appctx.NewExitError with an
			// empty message to avoid a double-printed summary line.
			return appctx.NewExitError(exitCode, "")
		},
	}
}
