package cmd

import (
	"fmt"

	appio "github.com/gh-xj/agent-repo-kit/cli/internal/io"
	skillop "github.com/gh-xj/agent-repo-kit/cli/internal/skillbuilder"
)

// SkillInitCmd scaffolds a skill router.
type SkillInitCmd struct {
	SkillDir    string `name:"skill-dir" help:"path to the skill directory to create"`
	Name        string `help:"skill name for SKILL.md frontmatter"`
	Description string `help:"trigger-oriented skill description"`
}

func (c *SkillInitCmd) Run(globals *CLI) error {
	result, err := skillop.InitSkill(skillop.InitConfig{
		SkillDir:    c.SkillDir,
		Name:        c.Name,
		Description: c.Description,
	})
	if err != nil {
		return err
	}

	out := globals.stdout()
	if globals.JSON {
		return appio.WriteJSON(out, map[string]any{
			"command":    "skill init",
			"ok":         true,
			"skill_dir":  result.SkillDir,
			"skill_path": result.SkillPath,
			"created":    result.Created,
			"cli_path":   result.CLIPath,
		})
	}

	fmt.Fprintf(out, "[ark skill init] created %s\n", result.SkillPath)
	if result.CLIPath != "" {
		fmt.Fprintf(out, "[ark skill init] scaffolded %s\n", result.CLIPath)
	}
	return nil
}
