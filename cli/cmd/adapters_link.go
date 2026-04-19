package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gh-xj/agent-repo-kit/cli/internal/adapters"
)

// AdaptersLinkCmd symlinks every skill under skills/ into a harness's
// skill root.
type AdaptersLinkCmd struct {
	Target   string `help:"harness target name (e.g. claude-code, codex)" required:""`
	Manifest string `help:"path to adapters/manifest.json (default: <repo-root>/adapters/manifest.json)"`
	RepoRoot string `name:"repo-root" help:"path to the agent-repo-kit checkout" default:"."`
	DryRun   bool   `name:"dry-run" help:"preview actions without touching the filesystem"`
}

func (c *AdaptersLinkCmd) Run(globals *CLI) error {
	if strings.TrimSpace(c.Target) == "" {
		return fmt.Errorf("--target is required")
	}
	manifest := c.Manifest
	if strings.TrimSpace(manifest) == "" {
		manifest = filepath.Join(c.RepoRoot, "adapters", "manifest.json")
	}
	return adapters.RunLink(c.RepoRoot, manifest, c.Target, c.DryRun)
}
