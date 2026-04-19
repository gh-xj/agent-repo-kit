package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/skillsync"
)

// SkillSyncCmd renders every per-adapter target from the canonical SKILL.md
// source files discovered under <repo-root>/skills/.
type SkillSyncCmd struct {
	RepoRoot string `name:"repo-root" help:"path to the repository root" default:"."`
}

func (c *SkillSyncCmd) Run(globals *CLI) error {
	plan, err := loadSkillPlan(c.RepoRoot)
	if err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	if err := skillsync.Sync(c.RepoRoot, plan); err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	out := globals.stdout()
	for _, skill := range plan.Skills {
		for _, target := range skill.Targets {
			fmt.Fprintf(out, "synced %s -> %s\n", skill.ID, target.Adapter)
		}
	}
	return nil
}

// loadSkillPlan reads the repo-root config and auto-discovers the skill
// plan. Shared by `ark skill sync` and `ark skill check`.
func loadSkillPlan(repoRoot string) (skillsync.Plan, error) {
	cfg, err := skillsync.LoadConfig(filepath.Join(repoRoot, skillsync.ConfigDefaultPath))
	if err != nil {
		return skillsync.Plan{}, err
	}
	return skillsync.BuildPlan(repoRoot, cfg)
}
