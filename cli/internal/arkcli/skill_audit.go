package arkcli

import (
	"errors"
	"fmt"
	"io"
	"sort"

	appio "github.com/gh-xj/agent-repo-kit/cli/internal/io"
	skillop "github.com/gh-xj/agent-repo-kit/cli/internal/skillbuilder"
)

// SkillAuditCmd audits a skill router, references, and local CLI layout.
type SkillAuditCmd struct {
	SkillDir string `name:"skill-dir" help:"path to the skill directory to audit"`
}

func (c *SkillAuditCmd) Run(globals *CLI) error {
	result, err := skillop.AuditSkill(skillop.AuditConfig{SkillDir: c.SkillDir})
	if err != nil {
		return err
	}

	out := globals.stdout()
	if globals.JSON {
		if err := appio.WriteJSON(out, map[string]any{
			"command":          "skill audit",
			"ok":               !result.HasErrors(),
			"skill_dir":        result.SkillDir,
			"skill_path":       result.SkillPath,
			"name":             result.Name,
			"description":      result.Description,
			"line_count":       result.LineCount,
			"has_cli":          result.HasCLI,
			"referenced_files": result.Referenced,
			"findings":         result.Findings,
		}); err != nil {
			return err
		}
	} else {
		printAuditResult(out, result)
	}

	if result.HasErrors() {
		return errors.New("skill audit failed")
	}
	return nil
}

func printAuditResult(out io.Writer, result skillop.AuditResult) {
	fmt.Fprintf(out, "[ark skill audit] %s\n", result.SkillPath)
	fmt.Fprintf(out, "  line_count: %d\n", result.LineCount)
	fmt.Fprintf(out, "  has_cli: %t\n", result.HasCLI)
	if result.Name != "" {
		fmt.Fprintf(out, "  name: %s\n", result.Name)
	}
	if result.Description != "" {
		fmt.Fprintf(out, "  description: %s\n", result.Description)
	}
	if len(result.Referenced) > 0 {
		fmt.Fprintf(out, "  referenced_files:\n")
		for _, ref := range result.Referenced {
			fmt.Fprintf(out, "    - %s\n", ref)
		}
	}
	if len(result.Findings) == 0 {
		fmt.Fprintln(out, "  findings: none")
		return
	}

	findings := append([]skillop.Finding(nil), result.Findings...)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Level != findings[j].Level {
			return findings[i].Level < findings[j].Level
		}
		return findings[i].Code < findings[j].Code
	})

	fmt.Fprintf(out, "  findings:\n")
	for _, finding := range findings {
		if finding.Path != "" {
			fmt.Fprintf(out, "    - [%s] %s: %s (%s)\n", finding.Level, finding.Code, finding.Message, finding.Path)
			continue
		}
		fmt.Fprintf(out, "    - [%s] %s: %s\n", finding.Level, finding.Code, finding.Message)
	}
}
