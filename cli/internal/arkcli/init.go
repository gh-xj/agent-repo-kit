package arkcli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/prompt"
	"github.com/gh-xj/agent-repo-kit/cli/internal/scaffold"
)

// Wizard option lists match scaffold.normalizeOptions' allowlists
// (cli/internal/scaffold/scaffold.go) and scaffold.detectProfiles
// (cli/internal/scaffold/detect.go). Leaving profiles empty triggers
// auto-detect.
var (
	wizardProfileOptions = []string{"go", "typescript-react", "python"}
	wizardOpsOptions     = []string{"work", "wiki"}
	wizardRiskOptions    = []string{"standard", "elevated", "critical"}
)

// InitCmd scaffolds the tracked convention contract into a repo.
//
// Two modes:
//   - Non-interactive (--batch, non-TTY stdin, or any value flag set):
//     the flag-driven path runs unchanged.
//   - Interactive: a stdlib-prompt wizard collects profiles, operations,
//     and repo-risk, then calls scaffold.RunInit with the same Options.
type InitCmd struct {
	Profiles string `help:"comma-separated repo profiles (auto-detect if empty)"`
	Ops      string `help:"comma-separated operations to scaffold" default:"work,wiki"`
	RepoRisk string `name:"repo-risk" help:"repository risk classification" default:"standard"`
	RepoRoot string `name:"repo-root" help:"path to the repository root" default:"."`
	Batch    bool   `help:"skip wizard even when stdin is a TTY"`
}

func (c *InitCmd) Run(globals *CLI) error {
	if c.Batch || !prompt.IsInteractive(os.Stdin) {
		return c.runNonInteractive(globals.stdout(), globals.stderr())
	}
	return c.runInteractive(os.Stdin, globals.stdout(), globals.stderr())
}

func (c *InitCmd) runNonInteractive(stdout, stderr io.Writer) error {
	opts := scaffold.Options{
		Profiles:   []string{c.Profiles},
		Operations: []string{c.Ops},
		RepoRisk:   c.RepoRisk,
	}
	return invokeScaffold(c.RepoRoot, opts, stdout, stderr)
}

func (c *InitCmd) runInteractive(stdin io.Reader, stdout, stderr io.Writer) error {
	absRoot, err := filepath.Abs(c.RepoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	// 1. Overwrite guard.
	contractPath := filepath.Join(absRoot, ".convention-engineering.json")
	if _, err := os.Stat(contractPath); err == nil {
		ok, err := prompt.Confirm(stdin, stdout,
			fmt.Sprintf("Existing %s found. Overwrite?", contractPath), false)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(stdout, "[ark init] aborted")
			return nil
		}
	}

	// 2. Profiles. Empty selection → auto-detect via scaffold.detectProfiles.
	profilesIdx, err := prompt.MultiSelect(stdin, stdout,
		"Which profiles describe this repo? (empty = auto-detect)",
		wizardProfileOptions, nil)
	if err != nil {
		return err
	}
	profiles := pickByIndex(wizardProfileOptions, profilesIdx)

	// 3. Operations (default: work, wiki).
	defaultOps := []int{0, 1}
	opsIdx, err := prompt.MultiSelect(stdin, stdout,
		"Which operational tracks to scaffold?",
		wizardOpsOptions, defaultOps)
	if err != nil {
		return err
	}
	ops := pickByIndex(wizardOpsOptions, opsIdx)

	// 4. Repo risk.
	riskIdx, err := prompt.Select(stdin, stdout,
		"Repo risk classification?",
		wizardRiskOptions, 0)
	if err != nil {
		return err
	}
	risk := wizardRiskOptions[riskIdx]

	// 5. Scaffold confirmation.
	fmt.Fprintln(stdout, "")
	fmt.Fprintf(stdout, "  profiles    : %s\n", strings.Join(profiles, ", "))
	fmt.Fprintf(stdout, "  operations  : %s\n", strings.Join(ops, ", "))
	fmt.Fprintf(stdout, "  repo-risk   : %s\n", risk)
	fmt.Fprintln(stdout, "")
	ok, err := prompt.Confirm(stdin, stdout, "Scaffold with the above settings?", true)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(stdout, "[ark init] aborted")
		return nil
	}

	opts := scaffold.Options{
		Profiles:   profiles,
		Operations: ops,
		RepoRisk:   risk,
	}
	if err := invokeScaffold(c.RepoRoot, opts, stdout, stderr); err != nil {
		return err
	}

	// Non-fatal post-scaffold verification.
	if _, lookErr := exec.LookPath("task"); lookErr == nil {
		fmt.Fprintln(stdout, "[ark init] running `task verify` as smoke test")
		cmd := exec.Command("task", "verify")
		cmd.Dir = absRoot
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(stderr, "[ark init] WARN: task verify failed (non-fatal): %v\n", err)
		}
	} else {
		fmt.Fprintln(stdout, "[ark init] `task` not on PATH; skipping verify smoke test")
	}
	return nil
}

func invokeScaffold(repoRoot string, opts scaffold.Options, stdout, stderr io.Writer) error {
	exitCode := scaffold.RunInit(repoRoot, opts, stdout, stderr)
	if exitCode == 0 {
		return nil
	}
	// RunInit has already written detailed stderr output; surface the
	// exact exit code via appctx.ExitError with an empty message to avoid
	// double-printing a redundant summary line.
	return appctx.NewExitError(exitCode, "")
}

func pickByIndex(options []string, idx []int) []string {
	picked := make([]string, 0, len(idx))
	for _, i := range idx {
		if i >= 0 && i < len(options) {
			picked = append(picked, options[i])
		}
	}
	return picked
}
