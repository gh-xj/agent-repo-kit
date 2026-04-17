package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/skill-builder/cli/internal/appctx"
	appio "github.com/gh-xj/agent-repo-kit/skill-builder/cli/internal/io"
	skillop "github.com/gh-xj/agent-repo-kit/skill-builder/cli/operator"
)

func AuditCommand() command {
	return command{
		Description: "audit a skill router, references, and local CLI layout",
		Configure: func(command *cobra.Command) {
			command.Flags().String("skill-dir", "", "path to the skill directory to audit")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			skillDir, _ := command.Flags().GetString("skill-dir")

			result, err := skillop.AuditSkill(skillop.AuditConfig{SkillDir: skillDir})
			if err != nil {
				return err
			}

			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				if err := appio.WriteJSON(os.Stdout, map[string]any{
					"command":          "audit",
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
				printAuditResult(result)
			}

			if result.HasErrors() {
				return errors.New("skill audit failed")
			}
			return nil
		},
	}
}

func printAuditResult(result skillop.AuditResult) {
	fmt.Fprintf(os.Stdout, "[skill-builder audit] %s\n", result.SkillPath)
	fmt.Fprintf(os.Stdout, "  line_count: %d\n", result.LineCount)
	fmt.Fprintf(os.Stdout, "  has_cli: %t\n", result.HasCLI)
	if result.Name != "" {
		fmt.Fprintf(os.Stdout, "  name: %s\n", result.Name)
	}
	if result.Description != "" {
		fmt.Fprintf(os.Stdout, "  description: %s\n", result.Description)
	}
	if len(result.Referenced) > 0 {
		fmt.Fprintf(os.Stdout, "  referenced_files:\n")
		for _, ref := range result.Referenced {
			fmt.Fprintf(os.Stdout, "    - %s\n", ref)
		}
	}
	if len(result.Findings) == 0 {
		fmt.Fprintln(os.Stdout, "  findings: none")
		return
	}

	findings := append([]skillop.Finding(nil), result.Findings...)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Level != findings[j].Level {
			return findings[i].Level < findings[j].Level
		}
		return findings[i].Code < findings[j].Code
	})

	fmt.Fprintf(os.Stdout, "  findings:\n")
	for _, finding := range findings {
		if finding.Path != "" {
			fmt.Fprintf(os.Stdout, "    - [%s] %s: %s (%s)\n", finding.Level, finding.Code, finding.Message, finding.Path)
			continue
		}
		fmt.Fprintf(os.Stdout, "    - [%s] %s: %s\n", finding.Level, finding.Code, finding.Message)
	}
}
