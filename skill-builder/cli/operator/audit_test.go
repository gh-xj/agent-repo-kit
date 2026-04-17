package operator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditSkillHappyPath(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "healthy-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		t.Fatalf("mkdir references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "references", "workflow.md"), []byte("# Workflow\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: healthy-skill
description: "Use when healthy skill checks are needed."
---

# Healthy Skill

See `+"`references/workflow.md`"+` for details.
`), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	result, err := AuditSkill(AuditConfig{SkillDir: skillDir})
	if err != nil {
		t.Fatalf("AuditSkill returned error: %v", err)
	}

	if result.HasErrors() {
		t.Fatalf("AuditSkill reported unexpected errors: %#v", result.Findings)
	}
	if result.LineCount == 0 {
		t.Fatalf("LineCount = 0, want > 0")
	}
}

func TestAuditSkillReportsMissingReferencesAndOversizeRouters(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "broken-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}

	var lines []string
	lines = append(lines,
		"---",
		"name: broken-skill",
		`description: "Use when broken skill checks are needed."`,
		"---",
		"",
		"# Broken Skill",
		"",
		"See `references/missing.md` for details.",
	)
	for i := 0; i < 401; i++ {
		lines = append(lines, "line filler")
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	result, err := AuditSkill(AuditConfig{SkillDir: skillDir})
	if err != nil {
		t.Fatalf("AuditSkill returned error: %v", err)
	}

	if !result.HasErrors() {
		t.Fatalf("expected errors, got %#v", result.Findings)
	}
	if !hasFinding(result.Findings, "missing_reference") {
		t.Fatalf("expected missing_reference finding, got %#v", result.Findings)
	}
	if !hasFinding(result.Findings, "router_too_large") {
		t.Fatalf("expected router_too_large finding, got %#v", result.Findings)
	}
}

func TestAuditSkillAcceptsUnquotedDescriptionAndIgnoresPlaceholderPaths(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "portable-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		t.Fatalf("mkdir references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "references", "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatalf("write notes: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: portable-skill
description: Use when validating placeholder path handling.
---

# Portable Skill

- Use `+"`references/notes.md`"+` for real details.
- Mention `+"`tools/<name>`"+` and `+"`scripts/`"+` as examples only.
`), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	result, err := AuditSkill(AuditConfig{SkillDir: skillDir})
	if err != nil {
		t.Fatalf("AuditSkill returned error: %v", err)
	}

	if result.Description == "" {
		t.Fatalf("expected description to be parsed")
	}
	if result.HasErrors() {
		t.Fatalf("unexpected errors: %#v", result.Findings)
	}
}

func hasFinding(findings []Finding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}
