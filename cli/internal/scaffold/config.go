package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func buildInitConfig(_ string, opts Options) ([]byte, error) {
	requiredFiles := []string{
		"AGENTS.md",
		"CLAUDE.md",
		"Taskfile.yml",
		ConventionsTaskfile,
		ConventionsCheckPath,
		"docs/README.md",
		"docs/requests/README.md",
		"docs/planning/README.md",
		"docs/plans/README.md",
		"docs/implementation/README.md",
		"docs/taxonomy/README.md",
	}
	if hasOperation(opts.Operations, "work") {
		requiredFiles = append(requiredFiles,
			".work/.gitignore",
			".work/config.yaml",
		)
	}
	if hasOperation(opts.Operations, "wiki") {
		requiredFiles = append(requiredFiles,
			".wiki/RULES.md",
			".wiki/Taskfile.yml",
			".wiki/pages/.gitkeep",
			".wiki/raw/.gitkeep",
			".wiki/scripts/lint.sh",
		)
	}

	taskfileChecks := map[string]any{
		"Taskfile.yml": []string{
			ConventionsTaskfile,
			"flatten: true",
		},
		ConventionsTaskfile: []string{
			"check:conventions:",
			"verify:",
			"check.sh",
		},
	}
	if hasOperation(opts.Operations, "work") {
		taskfileChecks["Taskfile.yml"] = append(taskfileChecks["Taskfile.yml"].([]string), "work:")
		taskfileChecks[ConventionsTaskfile] = append(taskfileChecks[ConventionsTaskfile].([]string), "work:check:")
	}
	if hasOperation(opts.Operations, "wiki") {
		taskfileChecks[ConventionsTaskfile] = append(taskfileChecks[ConventionsTaskfile].([]string), "../.wiki/Taskfile.yml")
		taskfileChecks[".wiki/Taskfile.yml"] = []string{"init:", "lint:"}
	}

	conventionMarkers := []string{
		"## Conventions",
		"**Docs**",
		ConfigPath,
		"`task verify`",
	}
	if hasOperation(opts.Operations, "work") {
		conventionMarkers = append(conventionMarkers, "**Work**", ".work/config.yaml")
	}
	if hasOperation(opts.Operations, "wiki") {
		conventionMarkers = append(conventionMarkers, ".wiki/RULES.md")
	}

	contentChecks := []map[string]any{
		{
			"name":             "agents-conventions",
			"file":             "AGENTS.md",
			"required_markers": conventionMarkers,
		},
		{
			"name":             "claude-conventions",
			"file":             "CLAUDE.md",
			"required_markers": conventionMarkers,
		},
		{
			"name": "docs-root-explained",
			"file": "docs/README.md",
			"required_markers": []string{
				ConfigPath,
				"requests/",
				"planning/",
				"plans/",
				"implementation/",
				"taxonomy/",
			},
		},
	}
	config := map[string]any{
		"generated_by":     ManagedMarker,
		"contract_version": 1,
		"mode":             "tracked",
		"profiles":         opts.Profiles,
		"docs_root":        "docs",
		"ownership_policy": map[string]any{
			"portable_skill_authoring_owner": "repo-maintainers",
			"domain_knowledge_owner":         "repo-maintainers",
			"repo_local_skills": map[string]any{
				"allowed":                false,
				"placement_roots":        []string{".claude/skills", ".agents/skills"},
				"authoring_owner":        "repo-maintainers",
				"requires_justification": true,
			},
		},
		"required_files":  requiredFiles,
		"taskfile_checks": taskfileChecks,
		"mirror_policy": map[string]any{
			"mode":  "mirrored",
			"files": []string{"CLAUDE.md", "AGENTS.md"},
		},
		"canonical_pointer_mode": "all",
		"canonical_pointers":     []any{},
		"content_checks":         contentChecks,
		"git_exclude_checks":     []any{},
		"invariant_contract": map[string]any{
			"required": false,
		},
		"evaluation_inputs": map[string]any{
			"repo_risk": opts.RepoRisk,
		},
		"chunk_plan": map[string]any{
			"enabled": false,
			"chunks":  []any{},
		},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(config); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeManagedJSON(path string, content []byte) error {
	if isManagedJSON(path) {
		return os.WriteFile(path, content, 0o644)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("refusing to overwrite unmanaged file %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func isManagedJSON(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	return strings.TrimSpace(fmt.Sprint(payload["generated_by"])) == ManagedMarker
}
