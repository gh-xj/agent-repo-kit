package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/prompt"
	"github.com/gh-xj/agent-repo-kit/cli/internal/scaffold"
)

func init() {
	registerCommand("init", InitCommand())
}

// Wizard option lists (locked by install-v2.md §6c).
var (
	wizardProfileOptions = []string{"web-service", "library", "monorepo", "data-pipeline", "tool"}
	wizardOpsOptions     = []string{"tickets", "wiki", "taskfile"}
	wizardRiskOptions    = []string{"standard", "elevated", "critical"}
)

// InitCommand wires the scaffold.RunInit entry point as the top-level
// `ark init` command.
//
// It has two modes:
//   - Non-interactive (--non-interactive, non-TTY stdin, or any of the
//     value flags set): the historical flag-driven path runs unchanged.
//   - Interactive: a tiny stdlib-prompt wizard collects profiles,
//     operations, and repo-risk, then calls scaffold.RunInit with the
//     same Options struct.
func InitCommand() command {
	return command{
		Description: "scaffold tracked convention contract into a repo",
		Configure: func(command *cobra.Command) {
			command.Flags().String("profiles", "", "comma-separated repo profiles (auto-detect if empty)")
			command.Flags().String("ops", "tickets,wiki", "comma-separated operations to scaffold")
			command.Flags().String("repo-risk", "standard", "repository risk classification")
			command.Flags().String("repo-root", ".", "path to the repository root")
			command.Flags().Bool("non-interactive", false, "skip wizard even when stdin is a TTY")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			nonInteractive, _ := command.Flags().GetBool("non-interactive")
			repoRoot, _ := command.Flags().GetString("repo-root")

			if nonInteractive || !prompt.IsInteractive(os.Stdin) {
				return runNonInteractiveInit(command, repoRoot, os.Stdout, os.Stderr)
			}
			return runInteractiveInit(command, repoRoot, os.Stdin, os.Stdout, os.Stderr)
		},
	}
}

func runNonInteractiveInit(command *cobra.Command, repoRoot string, stdout, stderr io.Writer) error {
	profiles, _ := command.Flags().GetString("profiles")
	ops, _ := command.Flags().GetString("ops")
	repoRisk, _ := command.Flags().GetString("repo-risk")

	opts := scaffold.Options{
		Profiles:   []string{profiles},
		Operations: []string{ops},
		RepoRisk:   repoRisk,
	}
	return invokeScaffold(repoRoot, opts, stdout, stderr)
}

func runInteractiveInit(command *cobra.Command, repoRoot string, stdin io.Reader, stdout, stderr io.Writer) error {
	absRoot, err := filepath.Abs(repoRoot)
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

	// 2. Profiles.
	profilesIdx, err := prompt.MultiSelect(stdin, stdout,
		"Which profiles describe this repo?",
		wizardProfileOptions, nil)
	if err != nil {
		return err
	}
	profiles := pickByIndex(wizardProfileOptions, profilesIdx)

	// 3. Operations (default: tickets, wiki).
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
	if err := invokeScaffold(repoRoot, opts, stdout, stderr); err != nil {
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
