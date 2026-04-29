package arkcli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gh-xj/agent-repo-kit/cli/internal/adapters"
)

// AdaptersListLinksCmd prints <srcAbs>\t<dstAbs> per discovered skill
// for a harness.
type AdaptersListLinksCmd struct {
	Target   string `help:"harness target name (e.g. claude-code, codex)" required:""`
	Manifest string `help:"path to adapters/manifest.json (default: <repo-root>/adapters/manifest.json)"`
	RepoRoot string `name:"repo-root" help:"path to the agent-repo-kit checkout" default:"."`
}

func (c *AdaptersListLinksCmd) Run(globals *CLI) error {
	if strings.TrimSpace(c.Target) == "" {
		return fmt.Errorf("--target is required")
	}
	manifest := c.Manifest
	if strings.TrimSpace(manifest) == "" {
		manifest = filepath.Join(c.RepoRoot, "adapters", "manifest.json")
	}
	return adapters.ListLinks(globals.stdout(), c.RepoRoot, manifest, c.Target)
}
