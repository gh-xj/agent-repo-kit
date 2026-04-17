package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/tasklint"
)

func init() {
	registerTaskfileCommand("lint", TaskfileLintCommand())
}

// TaskfileLintCommand wires tasklint.Lint as `ark taskfile lint`.
// V1 runs the fixed structural ruleset — no --strict or --config flags.
func TaskfileLintCommand() command {
	return command{
		Description: "lint Taskfile.yml against structural rules",
		Configure: func(command *cobra.Command) {
			command.Flags().String("repo-root", ".", "path to the repository root (for .gitignore lookup)")
			command.Flags().String("path", "Taskfile.yml", "path to the Taskfile.yml to lint (relative to repo-root or absolute)")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			repoRoot, _ := command.Flags().GetString("repo-root")
			taskfilePath, _ := command.Flags().GetString("path")
			if len(args) == 1 {
				taskfilePath = args[0]
			}

			jsonMode, _ := app.Values["json"].(bool)

			absRoot, err := filepath.Abs(repoRoot)
			if err != nil {
				return fmt.Errorf("resolve repo-root: %w", err)
			}

			absPath := taskfilePath
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(absRoot, absPath)
			}

			findings, err := tasklint.Lint(absPath, tasklint.Options{RepoRoot: absRoot})
			if err != nil {
				return err
			}

			if jsonMode {
				return emitFindingsJSON(command, findings)
			}
			emitFindingsHuman(command, findings)

			if len(findings) == 0 {
				return nil
			}
			return appctx.NewExitError(1, "")
		},
	}
}

func emitFindingsHuman(command *cobra.Command, findings []tasklint.Finding) {
	out := command.OutOrStdout()
	if len(findings) == 0 {
		fmt.Fprintln(out, "taskfile lint: clean")
		return
	}
	for _, f := range findings {
		fmt.Fprintf(out, "%s:%d:%d: [%s] %s\n", f.Path, f.Line, f.Column, f.RuleID, f.Message)
		if f.Detail != "" {
			fmt.Fprintf(out, "  detail: %s\n", f.Detail)
		}
		if f.Fix != "" {
			fmt.Fprintf(out, "  fix:    %s\n", f.Fix)
		}
		if f.Docs != "" {
			fmt.Fprintf(out, "  docs:   %s\n", f.Docs)
		}
	}
	fmt.Fprintf(out, "\ntaskfile lint: %d finding(s)\n", len(findings))
}

type findingJSON struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Detail   string `json:"detail,omitempty"`
	Fix      string `json:"fix,omitempty"`
	Docs     string `json:"docs,omitempty"`
}

type reportJSON struct {
	Findings []findingJSON `json:"findings"`
	Count    int           `json:"count"`
}

func emitFindingsJSON(command *cobra.Command, findings []tasklint.Finding) error {
	report := reportJSON{
		Findings: make([]findingJSON, 0, len(findings)),
		Count:    len(findings),
	}
	for _, f := range findings {
		report.Findings = append(report.Findings, findingJSON{
			RuleID:   f.RuleID,
			Severity: f.Severity.String(),
			Path:     f.Path,
			Line:     f.Line,
			Column:   f.Column,
			Message:  f.Message,
			Detail:   f.Detail,
			Fix:      f.Fix,
			Docs:     f.Docs,
		})
	}
	enc := json.NewEncoder(command.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
