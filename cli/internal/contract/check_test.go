package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunChecksPassesOnMinimalLayoutWithOptionalDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	results := RunChecks(root, defaultConfig())
	if failed := CountFailures(results); failed != 0 {
		t.Fatalf("expected 0 failures, got %d; results=%#v", failed, results)
	}
}

func TestTaskfileChecksResolveIncludeGraph(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\nincludes:\n  core:\n    taskfile: ./taskfiles/core.yml\n")
	writeFile(t, root, "taskfiles/core.yml", "version: '3'\ntasks:\n  verify:\n    cmds: [echo ok]\n")

	cfg := defaultConfig()
	cfg.TaskfileChecks = map[string][]string{"Taskfile.yml": {"verify:"}}

	results := RunChecks(root, cfg)
	if hasFailure(results, "task:Taskfile.yml:verify:") {
		t.Fatalf("expected verify token to be found in include graph, got %#v", results)
	}
}

func TestInvariantOptionalMissingPasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	cfg := defaultConfig()
	cfg.InvariantContract.Required = false

	results := RunChecks(root, cfg)
	if hasFailure(results, "invariants:.claude/architecture/invariants.yaml") {
		t.Fatalf("expected optional invariant check to pass when missing, got %#v", results)
	}
}

func TestInvariantRequiredMissingFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	cfg := defaultConfig()
	cfg.InvariantContract.Required = true

	results := RunChecks(root, cfg)
	if !hasFailure(results, "invariants:.claude/architecture/invariants.yaml") {
		t.Fatalf("expected required invariant file failure, got %#v", results)
	}
}

func TestCanonicalPointerAnyModePassesIfOneMatches(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, "AGENTS.md", "mirror contract text\n")
	writeFile(t, root, "CLAUDE.md", "other text\n")

	cfg := defaultConfig()
	cfg.CanonicalPointerMode = "any"
	cfg.CanonicalPointers = []CanonicalPointerConfig{
		{Name: "agents-check", File: "AGENTS.md", Text: "mirror contract text"},
		{Name: "claude-check", File: "CLAUDE.md", Text: "missing marker"},
	}

	results := RunChecks(root, cfg)
	if hasFailure(results, "canonical-pointers:any") {
		t.Fatalf("expected canonical pointers any mode to pass, got %#v", results)
	}
}

func TestContentChecksFailOnMissingMarker(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, "docs/QUALITY_SCORE.md", "# Quality\n## Onboarding Quality Rubric\n")

	cfg := defaultConfig()
	cfg.ContentChecks = []ContentCheckConfig{{
		Name:            "quality-markers",
		File:            "docs/QUALITY_SCORE.md",
		RequiredMarkers: []string{"## Onboarding Quality Rubric", "## Quality Gates"},
	}}

	results := RunChecks(root, cfg)
	if !hasFailure(results, "quality-markers") {
		t.Fatalf("expected missing marker failure, got %#v", results)
	}
}

func TestGitExcludeChecksPassWhenPatternsPresent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, ".git/info/exclude", "# local overlays\n.docs\n.docs/\nCLAUDE.local.md\n")

	cfg := defaultConfig()
	cfg.GitExcludeChecks = []GitExcludeCheckConfig{{
		Name: "oss-local-overlay",
		File: ".git/info/exclude",
		RequiredPatterns: []string{
			".docs",
			".docs/",
			"CLAUDE.local.md",
		},
	}}

	results := RunChecks(root, cfg)
	if hasFailure(results, "oss-local-overlay") {
		t.Fatalf("expected git exclude check to pass, got %#v", results)
	}
}

func TestGitExcludeChecksFailOnMissingPattern(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, ".git/info/exclude", ".docs\n")

	cfg := defaultConfig()
	cfg.GitExcludeChecks = []GitExcludeCheckConfig{{
		Name:             "oss-local-overlay",
		File:             ".git/info/exclude",
		RequiredPatterns: []string{".docs", ".docs/", "CLAUDE.local.md"},
	}}

	results := RunChecks(root, cfg)
	if !hasFailure(results, "oss-local-overlay") {
		t.Fatalf("expected git exclude check to fail, got %#v", results)
	}
}

// --- Shared test helpers ---

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir failed for %s: %v", rel, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed for %s: %v", rel, err)
	}
}

func writeJSONConfig(t *testing.T, root, rel string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal json failed for %s: %v", rel, err)
	}
	writeFile(t, root, rel, string(data))
}

func hasFailure(results []CheckResult, name string) bool {
	for _, r := range results {
		if r.Name == name && !r.Passed {
			return true
		}
	}
	return false
}

func baseContract(mode, docsRoot string) map[string]any {
	return map[string]any{
		"contract_version": 1,
		"mode":             mode,
		"profiles":         []string{"go"},
		"docs_root":        docsRoot,
		"ownership_policy": map[string]any{
			"portable_skill_authoring_owner": "skill-builder",
			"domain_knowledge_owner":         "domain-skills",
			"repo_local_skills": map[string]any{
				"allowed":                false,
				"placement_roots":        []string{".claude/skills", ".agents/skills"},
				"authoring_owner":        "skill-builder",
				"requires_justification": true,
			},
		},
		"mirror_policy": map[string]any{
			"mode":  "mirrored",
			"files": []string{"CLAUDE.md", "AGENTS.md"},
		},
		"evaluation_inputs": map[string]any{
			"repo_risk": "standard",
		},
		"chunk_plan": map[string]any{
			"enabled": false,
			"chunks":  []map[string]any{},
		},
	}
}
