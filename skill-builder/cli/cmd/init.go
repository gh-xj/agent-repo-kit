package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/skill-builder/cli/internal/appctx"
	appio "github.com/gh-xj/agent-repo-kit/skill-builder/cli/internal/io"
	skillop "github.com/gh-xj/agent-repo-kit/skill-builder/cli/operator"
)

func InitCommand() command {
	return command{
		Description: "scaffold a skill router",
		Configure: func(command *cobra.Command) {
			command.Flags().String("skill-dir", "", "path to the skill directory to create")
			command.Flags().String("name", "", "skill name for SKILL.md frontmatter")
			command.Flags().String("description", "", "trigger-oriented skill description")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			skillDir, _ := command.Flags().GetString("skill-dir")
			name, _ := command.Flags().GetString("name")
			description, _ := command.Flags().GetString("description")

			result, err := skillop.InitSkill(skillop.InitConfig{
				SkillDir:    skillDir,
				Name:        name,
				Description: description,
			})
			if err != nil {
				return err
			}

			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				return appio.WriteJSON(os.Stdout, map[string]any{
					"command":    "init",
					"ok":         true,
					"skill_dir":  result.SkillDir,
					"skill_path": result.SkillPath,
					"created":    result.Created,
					"cli_path":   result.CLIPath,
				})
			}

			fmt.Fprintf(os.Stdout, "[skill-builder init] created %s\n", result.SkillPath)
			if result.CLIPath != "" {
				fmt.Fprintf(os.Stdout, "[skill-builder init] scaffolded %s\n", result.CLIPath)
			}
			return nil
		},
	}
}
