package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/adapters"
	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

var adaptersCommandRegistry = map[string]command{}

func init() {
	registerAdaptersCommand("link", adaptersLinkCommand())
	registerAdaptersCommand("list-links", adaptersListLinksCommand())
}

func registerAdaptersCommand(name string, cmd command) {
	adaptersCommandRegistry[name] = cmd
}

func newAdaptersCmd() *cobra.Command {
	adaptersCmd := &cobra.Command{
		Use:          "adapters",
		Short:        "manage harness adapter symlinks (link, list-links)",
		SilenceUsage: true,
	}
	names := make([]string, 0, len(adaptersCommandRegistry))
	for name := range adaptersCommandRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		adaptersCmd.AddCommand(newLeafCmd(name, adaptersCommandRegistry[name]))
	}
	return adaptersCmd
}

func adaptersLinkCommand() command {
	return command{
		Description: "symlink every skill under skills/ into a harness's skill root",
		Configure: func(command *cobra.Command) {
			command.Flags().String("target", "", "harness target name (e.g. claude-code, codex)")
			command.Flags().String("manifest", "", "path to adapters/manifest.json (default: <repo-root>/adapters/manifest.json)")
			command.Flags().String("repo-root", ".", "path to the agent-repo-kit checkout")
			command.Flags().Bool("dry-run", false, "preview actions without touching the filesystem")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}
			target, _ := command.Flags().GetString("target")
			manifest, _ := command.Flags().GetString("manifest")
			repoRoot, _ := command.Flags().GetString("repo-root")
			dryRun, _ := command.Flags().GetBool("dry-run")

			if strings.TrimSpace(target) == "" {
				return fmt.Errorf("--target is required")
			}
			if strings.TrimSpace(manifest) == "" {
				manifest = filepath.Join(repoRoot, "adapters", "manifest.json")
			}
			return adapters.RunLink(repoRoot, manifest, target, dryRun)
		},
	}
}

func adaptersListLinksCommand() command {
	return command{
		Description: "print <srcAbs>\\t<dstAbs> per discovered skill for a harness",
		Configure: func(command *cobra.Command) {
			command.Flags().String("target", "", "harness target name (e.g. claude-code, codex)")
			command.Flags().String("manifest", "", "path to adapters/manifest.json (default: <repo-root>/adapters/manifest.json)")
			command.Flags().String("repo-root", ".", "path to the agent-repo-kit checkout")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}
			target, _ := command.Flags().GetString("target")
			manifest, _ := command.Flags().GetString("manifest")
			repoRoot, _ := command.Flags().GetString("repo-root")

			if strings.TrimSpace(target) == "" {
				return fmt.Errorf("--target is required")
			}
			if strings.TrimSpace(manifest) == "" {
				manifest = filepath.Join(repoRoot, "adapters", "manifest.json")
			}
			return adapters.ListLinks(os.Stdout, repoRoot, manifest, target)
		},
	}
}
