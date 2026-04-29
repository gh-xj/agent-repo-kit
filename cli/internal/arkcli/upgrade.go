package arkcli

import "github.com/gh-xj/agent-repo-kit/cli/internal/upgrade"

// UpgradeCmd upgrades the running ark/work binaries in place.
// Flavor detection (clone vs prebuilt) is done by upgrade.DetectFlavor
// against os.Executable().
type UpgradeCmd struct {
	Target string `help:"harness target to re-link after upgrade (empty to skip)"`
	DryRun bool   `name:"dry-run" help:"preview actions without touching the filesystem or network"`
	Prefix string `help:"binary install directory (default: directory of os.Executable())"`
}

func (c *UpgradeCmd) Run(globals *CLI) error {
	return upgrade.RunUpgrade(c.Target, c.Prefix, c.DryRun)
}
