package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitCommandCreatesSkill(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "demo-skill")

	output, err := runCLI(t,
		"init",
		"--skill-dir", skillDir,
		"--name", "demo-skill",
		"--description", "Use when building or refactoring a demo skill.",
	)
	if err != nil {
		t.Fatalf("init command failed: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Fatalf("expected SKILL.md: %v", err)
	}
	if !strings.Contains(output, "[skill-builder init] created") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestAuditCommandFailsForMissingReferences(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "broken-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: broken-skill
description: "Use when broken skill checks are needed."
---

# Broken Skill

See `+"`references/missing.md`"+` for details.
`), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	output, err := runCLI(t, "audit", "--skill-dir", skillDir)
	if err == nil {
		t.Fatalf("expected audit command to fail\n%s", output)
	}
	if !strings.Contains(output, "missing_reference") {
		t.Fatalf("expected missing_reference in output, got: %s", output)
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
