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
	registerSkillCommand("sync", SkillSyncCommand())
}

// SkillSyncCommand wires skillsync.Sync as the `ark skill sync` subcommand.
// It renders every per-adapter target declared in the manifest from the
// canonical SKILL.md source files.
func SkillSyncCommand() command {
	return command{
		Description: "render per-adapter SKILL files from canonical skill sources",
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

			if err := skillsync.Sync(repoRoot, manifest); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return appctx.NewExitError(appctx.ExitUsage, "")
			}

			for _, skill := range manifest.Skills {
				for _, target := range skill.Targets {
					fmt.Fprintf(os.Stdout, "synced %s -> %s\n", skill.ID, target.Adapter)
				}
			}
			return nil
		},
	}
}
