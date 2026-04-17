package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/skillsync"
)

func init() {
	registerSkillCommand("check", SkillCheckCommand())
}

// SkillCheckCommand wires skillsync.Check as the `ark skill check` subcommand.
// It re-renders every adapter target in memory and reports drift without
// writing to disk.
func SkillCheckCommand() command {
	return command{
		Description: "verify per-adapter SKILL files are in sync with canonical sources",
		Configure: func(command *cobra.Command) {
			command.Flags().String("manifest", skillsync.ManifestDefaultPath, "path to the skill-sync manifest")
			command.Flags().String("repo-root", ".", "path to the repository root")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			manifestFlag, _ := command.Flags().GetString("manifest")
			repoRoot, _ := command.Flags().GetString("repo-root")

			manifestPath := manifestFlag
			if !filepath.IsAbs(manifestPath) {
				manifestPath = filepath.Join(repoRoot, manifestFlag)
			}

			manifest, err := skillsync.LoadManifest(manifestPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return appctx.NewExitError(appctx.ExitUsage, "")
			}

			drifts, err := skillsync.Check(repoRoot, manifest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return appctx.NewExitError(appctx.ExitUsage, "")
			}

			if len(drifts) == 0 {
				fmt.Fprintln(os.Stdout, "skill-sync: in sync")
				return nil
			}

			for _, d := range drifts {
				adapter := d.Target.Adapter
				if adapter == "" {
					adapter = "source"
				}
				line := fmt.Sprintf("drift: %s [%s] %s", d.SkillID, adapter, d.Reason)
				if d.Diff != "" {
					line += " (" + d.Diff + ")"
				}
				fmt.Fprintln(os.Stdout, line)
			}
			return appctx.NewExitError(appctx.ExitError, "")
		},
	}
}
