package operator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitSkillCreatesSkeleton(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "demo-skill")

	result, err := InitSkill(InitConfig{
		SkillDir:    skillDir,
		Name:        "demo-skill",
		Description: "Use when building or refactoring a demo skill.",
	})
	if err != nil {
		t.Fatalf("InitSkill returned error: %v", err)
	}

	if result.SkillDir != skillDir {
		t.Fatalf("SkillDir = %q, want %q", result.SkillDir, skillDir)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{
		"name: demo-skill",
		`description: "Use when building or refactoring a demo skill."`,
		"# Demo Skill",
		"## Use This For",
		"## Router",
		"## References",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("SKILL.md missing %q\n%s", want, content)
		}
	}
}

func TestInitSkillWithCLIUsesInitializer(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "cli-skill")
	var gotDir string
	var gotModule string

	result, err := InitSkill(InitConfig{
		SkillDir:    skillDir,
		Name:        "cli-skill",
		Description: "Use when a skill should own deterministic CLI operations.",
		WithCLI:     true,
		CLIModule:   "example.com/cli-skill/cli",
		CLIInitializer: func(dir string, module string) error {
			gotDir = dir
			gotModule = module
			if err := os.MkdirAll(filepath.Join(dir, "cli"), 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "cli", "go.mod"), []byte("module "+module+"\n"), 0o644)
		},
	})
	if err != nil {
		t.Fatalf("InitSkill returned error: %v", err)
	}

	if gotDir != skillDir {
		t.Fatalf("CLI initializer dir = %q, want %q", gotDir, skillDir)
	}
	if gotModule != "example.com/cli-skill/cli" {
		t.Fatalf("CLI initializer module = %q", gotModule)
	}
	if result.CLIPath != filepath.Join(skillDir, "cli") {
		t.Fatalf("CLIPath = %q, want %q", result.CLIPath, filepath.Join(skillDir, "cli"))
	}
	if _, err := os.Stat(filepath.Join(skillDir, "cli", "go.mod")); err != nil {
		t.Fatalf("expected generated cli/go.mod: %v", err)
	}
}
