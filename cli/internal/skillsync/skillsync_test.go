package skillsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSourceWithFrontmatter = `---
name: demo-skill
description: demo description
---

# Demo Skill

Body line one.
Body line two.
`

const sampleSourceNoFrontmatter = `# Demo Skill

Body line one.
Body line two.
`

func TestSplitFrontmatterRejectsMidBodyTripleDash(t *testing.T) {
	// A markdown horizontal rule ("\n---" with no trailing newline, mid-body)
	// must NOT be treated as the closing fence. Without the fix this picked
	// up the HR and produced a corrupt frontmatter/body split.
	source := []byte("---\nname: demo\ndescription: d\n---\n\n# Heading\n\n---bad-hr\n\nmore body\n")
	fm, body := splitFrontmatter(source)
	if !strings.HasPrefix(string(fm), "---\nname: demo") {
		t.Fatalf("frontmatter should still capture the real YAML header, got %q", fm)
	}
	if !strings.Contains(string(body), "---bad-hr") {
		t.Fatalf("body should retain the mid-body '---bad-hr' marker, got %q", body)
	}
	// And the split position must be at the real closing fence, not the HR.
	if !strings.HasPrefix(string(body), "\n# Heading") {
		t.Fatalf("body should begin right after real fence, got %q", body)
	}
}

func TestSplitFrontmatterEOFFenceAccepted(t *testing.T) {
	// Files that end with a closing fence and no trailing newline are valid;
	// the whole thing is frontmatter, body is empty.
	source := []byte("---\nname: eof\ndescription: d\n---")
	fm, body := splitFrontmatter(source)
	if string(fm) != string(source) {
		t.Fatalf("expected whole source to be frontmatter, got fm=%q body=%q", fm, body)
	}
	if len(body) != 0 {
		t.Fatalf("expected empty body, got %q", body)
	}
}

func TestSplitFrontmatterNoCloseFenceReturnsOriginal(t *testing.T) {
	// Opening fence but no proper close anywhere → treat entire file as body.
	source := []byte("---\nname: broken\nno closing fence\n")
	fm, body := splitFrontmatter(source)
	if fm != nil {
		t.Fatalf("expected nil frontmatter when no closing fence, got %q", fm)
	}
	if string(body) != string(source) {
		t.Fatalf("expected body to equal source, got %q", body)
	}
}

func TestRenderFrontmatterBodyPreservesFrontmatter(t *testing.T) {
	out, err := Render([]byte(sampleSourceWithFrontmatter), Target{
		Adapter: "claude-code",
		Path:    "unused",
		Mode:    ModeFrontmatterBody,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	s := string(out)
	if !strings.HasPrefix(s, "---\nname: demo-skill\n") {
		t.Fatalf("frontmatter was not preserved at top; got %q", s[:minInt(80, len(s))])
	}
	if !strings.Contains(s, SyncSentinel) {
		t.Fatal("output missing SyncSentinel")
	}
	// Sentinel must land after the frontmatter closing fence.
	fenceIdx := strings.Index(s, "\n---\n")
	sentinelIdx := strings.Index(s, SyncSentinel)
	if fenceIdx < 0 || sentinelIdx < 0 || sentinelIdx < fenceIdx {
		t.Fatalf("sentinel should follow frontmatter: fence=%d sentinel=%d", fenceIdx, sentinelIdx)
	}
	if !strings.Contains(s, "# Demo Skill") {
		t.Fatal("body heading missing from output")
	}
}

func TestRenderBodyOnlyStripsFrontmatter(t *testing.T) {
	out, err := Render([]byte(sampleSourceWithFrontmatter), Target{
		Adapter: "codex",
		Path:    "unused",
		Mode:    ModeBodyOnly,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "name: demo-skill") {
		t.Fatalf("body-only mode leaked frontmatter: %q", s)
	}
	if !strings.HasPrefix(s, SyncSentinel+"\n\n") {
		t.Fatalf("body-only output must start with sentinel + blank line; got %q", s[:minInt(100, len(s))])
	}
	if !strings.Contains(s, "# Demo Skill") {
		t.Fatal("body heading missing from output")
	}
}

func TestRenderHandlesSourceWithoutFrontmatter(t *testing.T) {
	cases := []struct {
		name string
		mode string
	}{
		{"frontmatter+body", ModeFrontmatterBody},
		{"body-only", ModeBodyOnly},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Render([]byte(sampleSourceNoFrontmatter), Target{
				Adapter: "cursor",
				Path:    "unused",
				Mode:    tc.mode,
			})
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			s := string(out)
			if !strings.HasPrefix(s, SyncSentinel+"\n\n") {
				t.Fatalf("expected sentinel at top when no frontmatter; got %q", s[:minInt(100, len(s))])
			}
			if !strings.Contains(s, "# Demo Skill") {
				t.Fatal("body heading missing from output")
			}
			if strings.Contains(s, "---") {
				t.Fatalf("output should not contain YAML fence: %q", s)
			}
		})
	}
}

func TestRenderRejectsUnknownMode(t *testing.T) {
	_, err := Render([]byte(sampleSourceWithFrontmatter), Target{
		Adapter: "claude-code",
		Path:    "unused",
		Mode:    "not-a-mode",
	})
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}

// writeRepoFixture builds a repo-root layout with one "primary" skill and
// one ordinary skill, plus a v2 config naming the primary.
func writeRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	writeSkill := func(name string) {
		dir := filepath.Join(root, SkillsDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, SkillRouterFilename), []byte(sampleSourceWithFrontmatter), 0o644); err != nil {
			t.Fatalf("write skill %s: %v", name, err)
		}
	}
	writeSkill("primary-skill")
	writeSkill("other-skill")

	cfg := `{"schema_version": 2, "primary_skill": "primary-skill"}`
	if err := os.WriteFile(filepath.Join(root, ConfigDefaultPath), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return root
}

func loadPlan(t *testing.T, root string) Plan {
	t.Helper()
	cfg, err := LoadConfig(filepath.Join(root, ConfigDefaultPath))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	plan, err := BuildPlan(root, cfg)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return plan
}

func TestBuildPlanAutoDiscoversSkillsAndAdapters(t *testing.T) {
	root := writeRepoFixture(t)
	plan := loadPlan(t, root)

	if len(plan.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(plan.Skills))
	}

	// Skills come back sorted alphabetically.
	if plan.Skills[0].ID != "other-skill" || plan.Skills[1].ID != "primary-skill" {
		t.Fatalf("unexpected skill order: %s, %s", plan.Skills[0].ID, plan.Skills[1].ID)
	}

	want := map[string]map[string]struct {
		path string
		mode string
	}{
		"primary-skill": {
			"claude-code": {"adapters/claude-code/SKILL.md", ModeFrontmatterBody},
			"codex":       {"adapters/codex/SKILL.md", ModeBodyOnly},
			"cursor":      {"adapters/cursor/primary-skill.md", ModeBodyOnly},
		},
		"other-skill": {
			"claude-code": {"adapters/claude-code/other-skill/SKILL.md", ModeFrontmatterBody},
			"codex":       {"adapters/codex/other-skill/SKILL.md", ModeBodyOnly},
			"cursor":      {"adapters/cursor/other-skill.md", ModeBodyOnly},
		},
	}
	for _, skill := range plan.Skills {
		for _, target := range skill.Targets {
			exp, ok := want[skill.ID][target.Adapter]
			if !ok {
				t.Fatalf("unexpected adapter %q on skill %q", target.Adapter, skill.ID)
			}
			if filepath.ToSlash(target.Path) != exp.path {
				t.Fatalf("skill %q adapter %q: want path %q, got %q",
					skill.ID, target.Adapter, exp.path, target.Path)
			}
			if target.Mode != exp.mode {
				t.Fatalf("skill %q adapter %q: want mode %q, got %q",
					skill.ID, target.Adapter, exp.mode, target.Mode)
			}
		}
	}
}

func TestSyncWritesAllTargetsAndIsIdempotent(t *testing.T) {
	root := writeRepoFixture(t)
	plan := loadPlan(t, root)

	if err := Sync(root, plan); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	// All target files must exist.
	for _, skill := range plan.Skills {
		for _, target := range skill.Targets {
			if _, err := os.Stat(filepath.Join(root, target.Path)); err != nil {
				t.Fatalf("skill %s target %s not written: %v", skill.ID, target.Adapter, err)
			}
		}
	}

	// Snapshot content.
	before := make(map[string][]byte)
	for _, skill := range plan.Skills {
		for _, target := range skill.Targets {
			data, err := os.ReadFile(filepath.Join(root, target.Path))
			if err != nil {
				t.Fatalf("read target: %v", err)
			}
			before[target.Path] = data
		}
	}

	// Second Sync must be a byte-for-byte no-op.
	if err := Sync(root, plan); err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	for _, skill := range plan.Skills {
		for _, target := range skill.Targets {
			data, err := os.ReadFile(filepath.Join(root, target.Path))
			if err != nil {
				t.Fatalf("read target: %v", err)
			}
			if string(data) != string(before[target.Path]) {
				t.Fatalf("second Sync changed %s content", target.Path)
			}
		}
	}

	// Check should report zero drift immediately after Sync.
	drifts, err := Check(root, plan)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(drifts) != 0 {
		t.Fatalf("expected no drift, got %d: %+v", len(drifts), drifts)
	}
}

func TestCheckReportsMissingTarget(t *testing.T) {
	root := writeRepoFixture(t)
	plan := loadPlan(t, root)
	drifts, err := Check(root, plan)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	// Every target missing until we run Sync.
	wantTargets := 0
	for _, skill := range plan.Skills {
		wantTargets += len(skill.Targets)
	}
	if len(drifts) != wantTargets {
		t.Fatalf("expected %d drifts, got %d", wantTargets, len(drifts))
	}
	for _, d := range drifts {
		if d.Reason != "missing" {
			t.Fatalf("expected reason 'missing', got %q", d.Reason)
		}
	}
}

func TestCheckReportsStaleTarget(t *testing.T) {
	root := writeRepoFixture(t)
	plan := loadPlan(t, root)
	if err := Sync(root, plan); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	// Introduce drift in one target.
	staleTarget := plan.Skills[0].Targets[0]
	p := filepath.Join(root, staleTarget.Path)
	if err := os.WriteFile(p, []byte("stray edit\n"), 0o644); err != nil {
		t.Fatalf("write stale: %v", err)
	}
	drifts, err := Check(root, plan)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(drifts) != 1 {
		t.Fatalf("expected exactly 1 drift, got %d: %+v", len(drifts), drifts)
	}
	if drifts[0].Reason != "stale" {
		t.Fatalf("expected reason 'stale', got %q", drifts[0].Reason)
	}
	if drifts[0].Diff == "" {
		t.Fatal("expected diff summary on stale drift")
	}
	if !strings.Contains(drifts[0].Diff, "first differing line") {
		t.Fatalf("diff summary missing 'first differing line': %q", drifts[0].Diff)
	}
}

func TestLoadConfigRejectsWrongSchemaVersion(t *testing.T) {
	root := t.TempDir()
	cfg := `{"schema_version": 1, "primary_skill": "x"}`
	path := filepath.Join(root, ConfigDefaultPath)
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for schema_version != 2")
	}
	if !strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("error should mention schema_version, got: %v", err)
	}
}

func TestLoadConfigRejectsEmptyPrimarySkill(t *testing.T) {
	root := t.TempDir()
	cfg := `{"schema_version": 2}`
	path := filepath.Join(root, ConfigDefaultPath)
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error when primary_skill is empty")
	}
	if !strings.Contains(err.Error(), "primary_skill") {
		t.Fatalf("error should mention primary_skill, got: %v", err)
	}
}

func TestBuildPlanFailsWhenPrimarySkillMissing(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, SkillsDir, "only-skill")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, SkillRouterFilename), []byte(sampleSourceWithFrontmatter), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	_, err := BuildPlan(root, Config{SchemaVersion: 2, PrimarySkill: "nonexistent"})
	if err == nil {
		t.Fatal("expected error when primary_skill is not discoverable")
	}
	if !strings.Contains(err.Error(), "primary_skill") {
		t.Fatalf("error should mention primary_skill, got: %v", err)
	}
}

func TestBuildPlanFailsWhenNoSkillsDiscovered(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, SkillsDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := BuildPlan(root, Config{SchemaVersion: 2, PrimarySkill: "anything"})
	if err == nil {
		t.Fatal("expected error when skills dir is empty")
	}
	if !strings.Contains(err.Error(), "no skills") {
		t.Fatalf("error should mention 'no skills', got: %v", err)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
