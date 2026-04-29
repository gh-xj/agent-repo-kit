package arkcli

import (
	"fmt"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/skillsync"
)

// SkillCheckCmd re-renders every adapter target in memory and reports
// drift without writing to disk. Skills and targets are auto-discovered
// from <repo-root>/skills/.
type SkillCheckCmd struct {
	RepoRoot string `name:"repo-root" help:"path to the repository root" default:"."`
}

func (c *SkillCheckCmd) Run(globals *CLI) error {
	plan, err := loadSkillPlan(c.RepoRoot)
	if err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	drifts, err := skillsync.Check(c.RepoRoot, plan)
	if err != nil {
		fmt.Fprintln(globals.stderr(), err.Error())
		return appctx.NewExitError(appctx.ExitUsage, "")
	}

	out := globals.stdout()
	if len(drifts) == 0 {
		fmt.Fprintln(out, "skill-sync: in sync")
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
		fmt.Fprintln(out, line)
	}
	return appctx.NewExitError(appctx.ExitError, "")
}
