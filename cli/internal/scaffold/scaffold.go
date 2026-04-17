// Package scaffold ports `convention-engineering/scripts/init.go` into the
// `ark` CLI. The public entry point is RunInit, which scaffolds a tracked
// convention contract into a repo and then delegates to contract.Run for
// post-scaffold validation.
package scaffold

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

const (
	// ManagedMarker identifies files/blocks scaffolded by `ark init`.
	ManagedMarker = "agent-repo-kit --init"
	// managedBlockStart / managedBlockEnd bracket the upsert region in
	// tracked agent docs (AGENTS.md, CLAUDE.md).
	managedBlockStart = "<!-- agent-repo-kit:init:start -->"
	managedBlockEnd   = "<!-- agent-repo-kit:init:end -->"
	// ConventionsTaskfile is the repo-relative path to the tracked
	// convention Taskfile written by init.
	ConventionsTaskfile = ".convention-engineering/Taskfile.yml"
	// ConventionsCheckPath is the repo-relative path to the tracked
	// convention check shim.
	ConventionsCheckPath = ".convention-engineering/check.sh"
	// ConfigPath is the repo-relative path to the tracked convention
	// contract JSON file.
	ConfigPath = contract.DefaultConfigFile
)

// Options mirrors the CLI flag surface for `ark init`.
type Options struct {
	Profiles   []string
	Operations []string
	RepoRisk   string
}

// RunInit scaffolds the tracked convention contract into repoRoot and then
// runs the contract checker. It returns the integer exit code to propagate
// to the OS.
func RunInit(repoRoot string, opts Options, stdout, stderr io.Writer) int {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve repo root: %v\n", err)
		return 2
	}

	normalized, err := normalizeOptions(root, opts)
	if err != nil {
		fmt.Fprintf(stderr, "failed to prepare init options: %v\n", err)
		return 2
	}

	repoName := filepath.Base(root)
	if err := scaffoldTrackedRepo(root, repoName, normalized); err != nil {
		fmt.Fprintf(stderr, "failed to scaffold repo conventions: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "initialized tracked conventions for %s\n", repoName)
	return contract.Run(root, ConfigPath, false, stdout, stderr)
}

func normalizeOptions(root string, opts Options) (Options, error) {
	normalized := Options{
		Profiles:   normalizeCSVItems(opts.Profiles),
		Operations: normalizeCSVItems(opts.Operations),
		RepoRisk:   strings.TrimSpace(opts.RepoRisk),
	}
	if normalized.RepoRisk == "" {
		normalized.RepoRisk = "standard"
	}
	if len(normalized.Operations) == 0 {
		normalized.Operations = []string{"tickets", "wiki"}
	}
	if len(normalized.Profiles) == 0 {
		detected, err := detectProfiles(root)
		if err != nil {
			return Options{}, err
		}
		normalized.Profiles = detected
	}
	if len(normalized.Profiles) == 0 {
		return Options{}, fmt.Errorf("could not detect repo profiles; pass --profiles")
	}

	allowedOps := map[string]bool{"tickets": true, "wiki": true}
	for _, op := range normalized.Operations {
		if !allowedOps[op] {
			return Options{}, fmt.Errorf("unsupported operation %q (allowed: tickets,wiki)", op)
		}
	}

	seenProfile := map[string]bool{}
	profiles := make([]string, 0, len(normalized.Profiles))
	for _, profile := range normalized.Profiles {
		if profile == "" || seenProfile[profile] {
			continue
		}
		seenProfile[profile] = true
		profiles = append(profiles, profile)
	}
	normalized.Profiles = profiles

	seenOp := map[string]bool{}
	ops := make([]string, 0, len(normalized.Operations))
	for _, op := range normalized.Operations {
		if op == "" || seenOp[op] {
			continue
		}
		seenOp[op] = true
		ops = append(ops, op)
	}
	sort.Strings(ops)
	normalized.Operations = ops
	return normalized, nil
}

func normalizeCSVItems(items []string) []string {
	normalized := make([]string, 0)
	for _, item := range items {
		for _, part := range strings.Split(item, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" || strings.EqualFold(trimmed, "none") {
				continue
			}
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func scaffoldTrackedRepo(root, repoName string, opts Options) error {
	sourceRoot, err := resolveConventionEngineeringRoot()
	if err != nil {
		return err
	}

	configPath := filepath.Join(root, ConfigPath)
	configBytes, err := buildInitConfig(repoName, opts)
	if err != nil {
		return err
	}
	if err := writeManagedJSON(configPath, configBytes); err != nil {
		return err
	}

	if err := ensureDocsReadmes(root, repoName, opts); err != nil {
		return err
	}
	if err := ensureTickets(root, repoName, hasOperation(opts.Operations, "tickets")); err != nil {
		return err
	}
	if err := ensureWiki(root, hasOperation(opts.Operations, "wiki")); err != nil {
		return err
	}
	if err := ensureConventionSupportFiles(root, sourceRoot, opts); err != nil {
		return err
	}
	if err := ensureRootTaskfile(root); err != nil {
		return err
	}
	if err := ensureAgentContractFiles(root, opts); err != nil {
		return err
	}
	return nil
}

func hasOperation(operations []string, target string) bool {
	for _, op := range operations {
		if op == target {
			return true
		}
	}
	return false
}
