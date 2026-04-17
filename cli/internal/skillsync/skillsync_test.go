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

// writeManifestFixture builds a manifest plus canonical skill source on disk
// and returns the repo root.
func writeManifestFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	sourceDir := filepath.Join(root, "demo-skill")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(sampleSourceWithFrontmatter), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	manifest := `{
		"schema_version": 1,
		"skills": [
			{
				"id": "demo-skill",
				"source": "demo-skill/SKILL.md",
				"targets": [
					{"adapter": "claude-code", "path": "adapters/claude-code/SKILL.md", "mode": "frontmatter+body"},
					{"adapter": "codex", "path": "adapters/codex/SKILL.md", "mode": "body-only"},
					{"adapter": "cursor", "path": "adapters/cursor/demo-skill.md", "mode": "body-only"}
				]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(root, ManifestDefaultPath), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return root
}

func TestSyncWritesAllTargetsAndIsIdempotent(t *testing.T) {
	root := writeManifestFixture(t)
	m, err := LoadManifest(filepath.Join(root, ManifestDefaultPath))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	if err := Sync(root, m); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	// All target files must exist.
	for _, target := range m.Skills[0].Targets {
		if _, err := os.Stat(filepath.Join(root, target.Path)); err != nil {
			t.Fatalf("target %s not written: %v", target.Adapter, err)
		}
	}

	// Snapshot modification times and content.
	before := make(map[string][]byte)
	for _, target := range m.Skills[0].Targets {
		data, err := os.ReadFile(filepath.Join(root, target.Path))
		if err != nil {
			t.Fatalf("read target: %v", err)
		}
		before[target.Path] = data
	}

	// Second Sync must be a byte-for-byte no-op.
	if err := Sync(root, m); err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	for _, target := range m.Skills[0].Targets {
		data, err := os.ReadFile(filepath.Join(root, target.Path))
		if err != nil {
			t.Fatalf("read target: %v", err)
		}
		if string(data) != string(before[target.Path]) {
			t.Fatalf("second Sync changed %s content", target.Path)
		}
	}

	// Check should report zero drift immediately after Sync.
	drifts, err := Check(root, m)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(drifts) != 0 {
		t.Fatalf("expected no drift, got %d: %+v", len(drifts), drifts)
	}
}

func TestCheckReturnsEmptyWhenInSync(t *testing.T) {
	root := writeManifestFixture(t)
	m, err := LoadManifest(filepath.Join(root, ManifestDefaultPath))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if err := Sync(root, m); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	drifts, err := Check(root, m)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(drifts) != 0 {
		t.Fatalf("expected zero drifts, got %v", drifts)
	}
}

func TestCheckReportsMissingTarget(t *testing.T) {
	root := writeManifestFixture(t)
	m, err := LoadManifest(filepath.Join(root, ManifestDefaultPath))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	drifts, err := Check(root, m)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(drifts) != len(m.Skills[0].Targets) {
		t.Fatalf("expected %d drifts, got %d", len(m.Skills[0].Targets), len(drifts))
	}
	for _, d := range drifts {
		if d.Reason != "missing" {
			t.Fatalf("expected reason 'missing', got %q", d.Reason)
		}
	}
}

func TestCheckReportsStaleTarget(t *testing.T) {
	root := writeManifestFixture(t)
	m, err := LoadManifest(filepath.Join(root, ManifestDefaultPath))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if err := Sync(root, m); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	// Introduce drift in one target.
	staleTarget := m.Skills[0].Targets[0]
	p := filepath.Join(root, staleTarget.Path)
	if err := os.WriteFile(p, []byte("stray edit\n"), 0o644); err != nil {
		t.Fatalf("write stale: %v", err)
	}
	drifts, err := Check(root, m)
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

func TestLoadManifestRejectsUnknownMode(t *testing.T) {
	root := t.TempDir()
	manifest := `{
		"schema_version": 1,
		"skills": [
			{
				"id": "demo",
				"source": "demo/SKILL.md",
				"targets": [
					{"adapter": "claude-code", "path": "out.md", "mode": "garbage"}
				]
			}
		]
	}`
	path := filepath.Join(root, "m.json")
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "unknown mode") {
		t.Fatalf("error should mention unknown mode, got: %v", err)
	}
}

func TestLoadManifestRejectsDuplicateSkillID(t *testing.T) {
	root := t.TempDir()
	manifest := `{
		"schema_version": 1,
		"skills": [
			{"id": "dup", "source": "a/SKILL.md", "targets": [{"adapter":"codex","path":"a.md","mode":"body-only"}]},
			{"id": "dup", "source": "b/SKILL.md", "targets": [{"adapter":"codex","path":"b.md","mode":"body-only"}]}
		]
	}`
	path := filepath.Join(root, "m.json")
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected error for duplicate skill id")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error should mention duplicate, got: %v", err)
	}
}

func TestLoadManifestRejectsUnknownAdapter(t *testing.T) {
	root := t.TempDir()
	manifest := `{
		"schema_version": 1,
		"skills": [
			{"id": "demo", "source": "demo/SKILL.md", "targets": [{"adapter":"atom","path":"a.md","mode":"body-only"}]}
		]
	}`
	path := filepath.Join(root, "m.json")
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected error for unknown adapter")
	}
	if !strings.Contains(err.Error(), "unknown adapter") {
		t.Fatalf("error should mention unknown adapter, got: %v", err)
	}
}

func TestLoadManifestRejectsWrongSchemaVersion(t *testing.T) {
	root := t.TempDir()
	manifest := `{"schema_version": 2, "skills": []}`
	path := filepath.Join(root, "m.json")
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected error for schema_version != 1")
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
