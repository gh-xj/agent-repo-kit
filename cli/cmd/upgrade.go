package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/upgrade"
)

func init() {
	registerCommand("upgrade", UpgradeCommand())
}

// UpgradeCommand wires `ark upgrade`. Flavor detection (clone vs prebuilt)
// is done by upgrade.DetectFlavor against os.Executable().
func UpgradeCommand() command {
	return command{
		Description: "upgrade the running ark binary in place",
		Configure: func(command *cobra.Command) {
			command.Flags().String("target", "", "harness target to re-link after upgrade (empty to skip)")
			command.Flags().Bool("dry-run", false, "preview actions without touching the filesystem or network")
			command.Flags().String("prefix", "", "binary install directory (default: directory of os.Executable())")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}
			target, _ := command.Flags().GetString("target")
			dryRun, _ := command.Flags().GetBool("dry-run")
			prefix, _ := command.Flags().GetString("prefix")
			if prefix == "" {
				if self, err := os.Executable(); err == nil {
					prefix = filepath.Dir(self)
				}
			}
			return upgrade.RunUpgrade(target, prefix, dryRun)
		},
	}
}
