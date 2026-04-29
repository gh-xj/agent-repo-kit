package arkcli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/tasklint"
)

// TaskfileLintCmd runs the fixed structural ruleset on a Taskfile.yml.
// V1 has no --strict or --config flag.
//
// Path can come from either the --path flag or a positional argument.
// The positional wins if both are set, preserving the original
// cobra-era behavior.
type TaskfileLintCmd struct {
	RepoRoot string `name:"repo-root" help:"path to the repository root (for .gitignore lookup)" default:"."`
	Path     string `help:"path to the Taskfile.yml to lint (relative to repo-root or absolute)" default:"Taskfile.yml"`
	PathArg  string `arg:"" optional:"" name:"taskfile" help:"taskfile path (overrides --path when set)"`
}

func (c *TaskfileLintCmd) Run(globals *CLI) error {
	taskfilePath := c.Path
	if c.PathArg != "" {
		taskfilePath = c.PathArg
	}

	absRoot, err := filepath.Abs(c.RepoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo-root: %w", err)
	}

	absPath := taskfilePath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(absRoot, absPath)
	}

	findings, err := tasklint.Lint(absPath, tasklint.Options{RepoRoot: absRoot})
	if err != nil {
		// I/O failures (e.g. file not found) map to ExitUsage (2).
		return appctx.NewExitError(appctx.ExitUsage, err.Error())
	}

	// Emit findings first so both JSON and human modes produce output
	// before the exit-code decision. Keeping emit and exit-code mapping
	// independent means `--json` exits 1 on findings just like human mode.
	out := globals.stdout()
	if globals.JSON {
		if emitErr := emitFindingsJSON(out, findings); emitErr != nil {
			return emitErr
		}
	} else {
		emitFindingsHuman(out, findings)
	}

	if len(findings) == 0 {
		return nil
	}
	// Parse errors (YAML-level failures that short-circuit the rules)
	// surface as ExitUsage to signal "fix your input first"; all other
	// findings are normal lint failures.
	for _, f := range findings {
		if f.RuleID == "parse-error" {
			return appctx.NewExitError(appctx.ExitUsage, "")
		}
	}
	return appctx.NewExitError(appctx.ExitError, "")
}

func emitFindingsHuman(out io.Writer, findings []tasklint.Finding) {
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

func emitFindingsJSON(out io.Writer, findings []tasklint.Finding) error {
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
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
