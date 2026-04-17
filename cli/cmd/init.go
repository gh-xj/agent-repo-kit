package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/scaffold"
)

func init() {
	registerCommand("init", InitCommand())
}

// InitCommand wires the scaffold.RunInit entry point as the top-level
// `ark init` command.
func InitCommand() command {
	return command{
		Description: "scaffold tracked convention contract into a repo",
		Configure: func(command *cobra.Command) {
			command.Flags().String("profiles", "", "comma-separated repo profiles (auto-detect if empty)")
			command.Flags().String("ops", "tickets,wiki", "comma-separated operations to scaffold")
			command.Flags().String("repo-risk", "standard", "repository risk classification")
			command.Flags().String("repo-root", ".", "path to the repository root")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			profiles, _ := command.Flags().GetString("profiles")
			ops, _ := command.Flags().GetString("ops")
			repoRisk, _ := command.Flags().GetString("repo-risk")
			repoRoot, _ := command.Flags().GetString("repo-root")

			exitCode := scaffold.RunInit(repoRoot, scaffold.Options{
				Profiles:   []string{profiles},
				Operations: []string{ops},
				RepoRisk:   repoRisk,
			}, os.Stdout, os.Stderr)
			if exitCode == 0 {
				return nil
			}
			// RunInit has already written detailed stderr output; surface
			// the exact exit code via appctx.ExitError with an empty
			// message to avoid double-printing a redundant summary line.
			return appctx.NewExitError(exitCode, "")
		},
	}
}
