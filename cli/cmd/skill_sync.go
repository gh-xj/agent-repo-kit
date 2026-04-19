package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/skillsync"
)

// SkillSyncCmd renders every per-adapter target declared in the manifest
// from the canonical SKILL.md source files.
type SkillSyncCmd struct {
	Manifest string `help:"path to the skill-sync manifest" default:"${skillsync_manifest_default}"`
	RepoRoot string `name:"repo-root" help:"path to the repository root" default:"."`
}

func (c *SkillSyncCmd) Run(globals *CLI) error {
	manifestPath := c.Manifest
	if !filepath.IsAbs(manifestPath) {
		manifestPath = filepath.Join(c.RepoRoot, manifestPath)
	}

	manifest, err := skillsync.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	if err := skillsync.Sync(c.RepoRoot, manifest); err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	out := globals.stdout()
	for _, skill := range manifest.Skills {
		for _, target := range skill.Targets {
			fmt.Fprintf(out, "synced %s -> %s\n", skill.ID, target.Adapter)
		}
	}
	return nil
}
