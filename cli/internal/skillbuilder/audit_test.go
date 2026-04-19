package skillbuilder

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

func TestAuditSkillRejectsEmptySkillDir(t *testing.T) {
	for _, input := range []string{"", "   ", "\t"} {
		_, err := AuditSkill(AuditConfig{SkillDir: input})
		if err == nil {
			t.Fatalf("AuditSkill(%q) returned nil error; expected 'skill dir is required'", input)
		}
		if !strings.Contains(err.Error(), "skill dir is required") {
			t.Fatalf("AuditSkill(%q) = %v; expected 'skill dir is required'", input, err)
		}
	}
}

// TestAuditSkillIgnoresGoPackagePaths is a regression guard for the bug
// where backticked Go package paths like `samber/lo`, `spf13/cobra`, or
// `github.com/alecthomas/kong` were extracted as referenced relative paths
// and reported as missing_reference findings.
func TestAuditSkillIgnoresGoPackagePaths(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "go-package-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		t.Fatalf("mkdir references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "references", "logging.md"), []byte("# Logging\n"), 0o644); err != nil {
		t.Fatalf("write logging: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: go-package-skill
description: Use when evaluating Go package shorthand handling.
---

# Go Package Skill

- Reach for `+"`samber/lo`"+` only as a last resort.
- Prefer stdlib `+"`log/slog`"+` with `+"`lmittmann/tint`"+` colors.
- CLI parsing via `+"`spf13/cobra`"+` or `+"`github.com/alecthomas/kong`"+`.
- See `+"`references/logging.md`"+` for tradeoffs.
`), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	result, err := AuditSkill(AuditConfig{SkillDir: skillDir})
	if err != nil {
		t.Fatalf("AuditSkill returned error: %v", err)
	}
	if result.HasErrors() {
		t.Fatalf("unexpected errors on skill with Go package paths: %#v", result.Findings)
	}
	// The only path the extractor should have picked up is the real
	// references file; package shorthands must be absent.
	if len(result.Referenced) != 1 || result.Referenced[0] != filepath.Clean("references/logging.md") {
		t.Fatalf("Referenced = %v, want [references/logging.md]", result.Referenced)
	}
}

// TestAuditSkillStillFlagsMissingReferenceFiles guards the inverse: a
// genuine-looking relative path (e.g. `references/cobra-patterns.md`,
// `./file.sh`, `../other/README.md`) that is NOT on disk must still surface
// as a missing_reference finding.
func TestAuditSkillStillFlagsMissingReferenceFiles(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "missing-refs-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: missing-refs-skill
description: Use when verifying missing-reference detection still works.
---

# Missing Refs Skill

- See `+"`references/cobra-patterns.md`"+` for usage.
- Run `+"`./file.sh`"+` to bootstrap.
- Also `+"`../other/README.md`"+` from the parent dir.
`), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	result, err := AuditSkill(AuditConfig{SkillDir: skillDir})
	if err != nil {
		t.Fatalf("AuditSkill returned error: %v", err)
	}
	if !result.HasErrors() {
		t.Fatalf("expected missing_reference errors, got %#v", result.Findings)
	}

	wantPaths := []string{
		filepath.Clean("references/cobra-patterns.md"),
		filepath.Clean("file.sh"),
		filepath.Clean("../other/README.md"),
	}
	for _, want := range wantPaths {
		found := false
		for _, ref := range result.Referenced {
			if ref == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Referenced missing %q; got %v", want, result.Referenced)
		}
	}

	// Every referenced path should produce a missing_reference finding
	// (none exist on disk).
	missingCount := 0
	for _, f := range result.Findings {
		if f.Code == "missing_reference" {
			missingCount++
		}
	}
	if missingCount != len(wantPaths) {
		t.Fatalf("expected %d missing_reference findings, got %d: %#v",
			len(wantPaths), missingCount, result.Findings)
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
