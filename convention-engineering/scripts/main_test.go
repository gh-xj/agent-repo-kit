package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunChecksPassesOnMinimalLayoutWithOptionalDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	results := runChecks(root, defaultConfig())
	if failed := countFailures(results); failed != 0 {
		t.Fatalf("expected 0 failures, got %d; results=%#v", failed, results)
	}
}

func TestTaskfileChecksResolveIncludeGraph(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\nincludes:\n  core:\n    taskfile: ./taskfiles/core.yml\n")
	writeFile(t, root, "taskfiles/core.yml", "version: '3'\ntasks:\n  verify:\n    cmds: [echo ok]\n")

	cfg := defaultConfig()
	cfg.TaskfileChecks = map[string][]string{"Taskfile.yml": {"verify:"}}

	results := runChecks(root, cfg)
	if hasFailure(results, "task:Taskfile.yml:verify:") {
		t.Fatalf("expected verify token to be found in include graph, got %#v", results)
	}
}

func TestInvariantOptionalMissingPasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	cfg := defaultConfig()
	cfg.InvariantContract.Required = false

	results := runChecks(root, cfg)
	if hasFailure(results, "invariants:.claude/architecture/invariants.yaml") {
		t.Fatalf("expected optional invariant check to pass when missing, got %#v", results)
	}
}

func TestInvariantRequiredMissingFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	cfg := defaultConfig()
	cfg.InvariantContract.Required = true

	results := runChecks(root, cfg)
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
	cfg.CanonicalPointers = []canonicalPointerConfig{
		{Name: "agents-check", File: "AGENTS.md", Text: "mirror contract text"},
		{Name: "claude-check", File: "CLAUDE.md", Text: "missing marker"},
	}

	results := runChecks(root, cfg)
	if hasFailure(results, "canonical-pointers:any") {
		t.Fatalf("expected canonical pointers any mode to pass, got %#v", results)
	}
}

func TestContentChecksFailOnMissingMarker(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, "docs/QUALITY_SCORE.md", "# Quality\n## Onboarding Quality Rubric\n")

	cfg := defaultConfig()
	cfg.ContentChecks = []contentCheckConfig{{
		Name:            "quality-markers",
		File:            "docs/QUALITY_SCORE.md",
		RequiredMarkers: []string{"## Onboarding Quality Rubric", "## Quality Gates"},
	}}

	results := runChecks(root, cfg)
	if !hasFailure(results, "quality-markers") {
		t.Fatalf("expected missing marker failure, got %#v", results)
	}
}

func TestGitExcludeChecksPassWhenPatternsPresent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, ".git/info/exclude", "# local overlays\n.docs\n.docs/\nCLAUDE.local.md\n")

	cfg := defaultConfig()
	cfg.GitExcludeChecks = []gitExcludeCheckConfig{{
		Name: "oss-local-overlay",
		File: ".git/info/exclude",
		RequiredPatterns: []string{
			".docs",
			".docs/",
			"CLAUDE.local.md",
		},
	}}

	results := runChecks(root, cfg)
	if hasFailure(results, "oss-local-overlay") {
		t.Fatalf("expected git exclude check to pass, got %#v", results)
	}
}

func TestGitExcludeChecksFailOnMissingPattern(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeFile(t, root, ".git/info/exclude", ".docs\n")

	cfg := defaultConfig()
	cfg.GitExcludeChecks = []gitExcludeCheckConfig{{
		Name:             "oss-local-overlay",
		File:             ".git/info/exclude",
		RequiredPatterns: []string{".docs", ".docs/", "CLAUDE.local.md"},
	}}

	results := runChecks(root, cfg)
	if !hasFailure(results, "oss-local-overlay") {
		t.Fatalf("expected git exclude check to fail, got %#v", results)
	}
}

func TestRunJSONModeRequiresTrackedContractWhenNoConfigProvided(t *testing.T) {
	root := t.TempDir()

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(root, "", true, &out, &err)
	if exitCode != 2 {
		t.Fatalf("expected exit 2, got %d stderr=%s", exitCode, err.String())
	}
	if !strings.Contains(err.String(), ".convention-engineering.json") {
		t.Fatalf("expected tracked contract error, got %q", err.String())
	}
}

func TestRunJSONModeStructuredOutput(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeJSONConfig(t, root, ".convention-engineering.json", baseContract("tracked", "docs"))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(root, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}

	report := jsonReport{}
	if unmarshalErr := json.Unmarshal(out.Bytes(), &report); unmarshalErr != nil {
		t.Fatalf("expected valid json report: %v", unmarshalErr)
	}
	if !strings.HasSuffix(report.ConfigPath, ".convention-engineering.json") {
		t.Fatalf("expected tracked contract config path, got %q", report.ConfigPath)
	}
}

func TestLoadConfigRejectsMissingContractVersion(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "contract_version")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "contract_version") {
		t.Fatalf("expected contract_version error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownMajorVersion(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["contract_version"] = 2
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "unsupported contract_version major") {
		t.Fatalf("expected unknown major version error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["mode"] = "shadow"
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mode") {
		t.Fatalf("expected mode error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingProfiles(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "profiles")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles") {
		t.Fatalf("expected profiles error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingMirrorPolicy(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "mirror_policy")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mirror_policy") {
		t.Fatalf("expected mirror_policy error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingEvaluationInputs(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "evaluation_inputs")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "evaluation_inputs") {
		t.Fatalf("expected evaluation_inputs error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingChunkPlan(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "chunk_plan")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan") {
		t.Fatalf("expected chunk_plan error, got %v", err)
	}
}

func TestLoadConfigRejectsProfilesEmptyArray(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["profiles"] = []string{}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles") {
		t.Fatalf("expected profiles non-empty error, got %v", err)
	}
}

func TestLoadConfigRejectsProfileEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["profiles"] = []string{"go", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles[1]") {
		t.Fatalf("expected profiles item error, got %v", err)
	}
}

func TestLoadConfigRejectsDocsRootWithTrailingWhitespace(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["docs_root"] = ".docs "
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "docs_root") {
		t.Fatalf("expected exact docs_root enum error, got %v", err)
	}
}

func TestLoadConfigAcceptsAdditiveTopLevelFieldForV1(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["future_top_level_field"] = map[string]any{"mode": "reserved"}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := loadConfig(root, "config.json")
	if err != nil {
		t.Fatalf("expected additive v1 field to load, got %v", err)
	}
	if loaded.Mode != "overlay" {
		t.Fatalf("expected overlay mode, got %q", loaded.Mode)
	}
}

func TestLoadConfigRejectsRepoLocalSkillsMissingAllowed(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	repoLocal := cfg["ownership_policy"].(map[string]any)["repo_local_skills"].(map[string]any)
	delete(repoLocal, "allowed")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "ownership_policy.repo_local_skills.allowed") {
		t.Fatalf("expected repo_local_skills allowed error, got %v", err)
	}
}

func TestLoadConfigRejectsRepoLocalSkillsMissingRequiresJustification(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	repoLocal := cfg["ownership_policy"].(map[string]any)["repo_local_skills"].(map[string]any)
	delete(repoLocal, "requires_justification")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "ownership_policy.repo_local_skills.requires_justification") {
		t.Fatalf("expected repo_local_skills requires_justification error, got %v", err)
	}
}

func TestLoadConfigRejectsRepoLocalSkillsMissingPlacementRoots(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	repoLocal := cfg["ownership_policy"].(map[string]any)["repo_local_skills"].(map[string]any)
	delete(repoLocal, "placement_roots")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "ownership_policy.repo_local_skills.placement_roots") {
		t.Fatalf("expected repo_local_skills placement_roots error, got %v", err)
	}
}

func TestLoadConfigRejectsPlacementRootEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["ownership_policy"].(map[string]any)["repo_local_skills"].(map[string]any)["placement_roots"] = []string{".claude/skills", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "placement_roots[1]") {
		t.Fatalf("expected placement_roots item error, got %v", err)
	}
}

func TestLoadConfigAcceptsSpecShapedEvaluationInputsObject(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["evaluation_inputs"] = map[string]any{"repo_risk": "standard"}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := loadConfig(root, "config.json")
	if err != nil {
		t.Fatalf("expected spec-shaped evaluation_inputs object to load, got %v", err)
	}
	if loaded.EvaluationInputs.RepoRisk != "standard" {
		t.Fatalf("expected repo_risk=standard, got %#v", loaded.EvaluationInputs)
	}
}

func TestLoadConfigRejectsEvaluationInputsNull(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "config.json", `{
  "contract_version": 1,
  "mode": "overlay",
  "profiles": ["go"],
  "docs_root": ".docs",
  "ownership_policy": {
    "portable_skill_authoring_owner": "skill-builder",
    "domain_knowledge_owner": "domain-skills",
    "repo_local_skills": {
      "allowed": false,
      "placement_roots": [".claude/skills", ".codex/skills"],
      "authoring_owner": "skill-builder",
      "requires_justification": true
    }
  },
  "evaluation_inputs": null
}`)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "evaluation_inputs") || !strings.Contains(err.Error(), "null") {
		t.Fatalf("expected evaluation_inputs null error, got %v", err)
	}
}

func TestLoadConfigRejectsNullOptionalContractCollections(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     any
		wantError string
	}{
		{name: "required_files", field: "required_files", value: nil, wantError: "required_files"},
		{name: "taskfile_checks", field: "taskfile_checks", value: nil, wantError: "taskfile_checks"},
		{name: "canonical_pointers", field: "canonical_pointers", value: nil, wantError: "canonical_pointers"},
		{name: "content_checks", field: "content_checks", value: nil, wantError: "content_checks"},
		{name: "git_exclude_checks", field: "git_exclude_checks", value: nil, wantError: "git_exclude_checks"},
		{name: "invariant_contract", field: "invariant_contract", value: nil, wantError: "invariant_contract"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			cfg := baseContract("overlay", ".docs")
			cfg[tt.field] = tt.value
			writeJSONConfig(t, root, "config.json", cfg)

			_, _, err := loadConfig(root, "config.json")
			if err == nil || !strings.Contains(err.Error(), tt.wantError) || !strings.Contains(err.Error(), "null") {
				t.Fatalf("expected %s null error, got %v", tt.wantError, err)
			}
		})
	}
}

func TestLoadConfigRejectsNullInvariantContractArrays(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		wantError string
	}{
		{name: "required_fields", field: "required_fields", wantError: "invariant_contract.required_fields"},
		{name: "exception_fields", field: "exception_fields", wantError: "invariant_contract.exception_fields"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			cfg := baseContract("overlay", ".docs")
			cfg["invariant_contract"] = map[string]any{
				"required":      false,
				tt.field:        nil,
			}
			writeJSONConfig(t, root, "config.json", cfg)

			_, _, err := loadConfig(root, "config.json")
			if err == nil || !strings.Contains(err.Error(), tt.wantError) || !strings.Contains(err.Error(), "null") {
				t.Fatalf("expected %s null error, got %v", tt.wantError, err)
			}
		})
	}
}

func TestLoadConfigRejectsEvaluationInputsEmptyRepoRisk(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["evaluation_inputs"] = map[string]any{"repo_risk": ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "evaluation_inputs.repo_risk") {
		t.Fatalf("expected repo_risk error, got %v", err)
	}
}

func TestLoadConfigAcceptsSpecShapedChunkPlan(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("tracked", "docs")
	cfg["chunk_plan"] = map[string]any{
		"enabled": true,
		"chunks": []map[string]any{
			{
				"id":                  "checker-config-contract",
				"scope":               "contract surface",
				"completion_criteria": []string{"spec-shaped loader and schema align"},
				"depends_on":          []string{},
			},
			{
				"id":                  "schema-example-sync",
				"scope":               "references",
				"completion_criteria": []string{"example and schema match the contract"},
				"depends_on":          []string{"checker-config-contract"},
			},
		},
	}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := loadConfig(root, "config.json")
	if err != nil {
		t.Fatalf("expected spec-shaped chunk plan to load, got %v", err)
	}
	if !loaded.ChunkPlan.Enabled {
		t.Fatalf("expected chunk plan enabled, got %#v", loaded.ChunkPlan)
	}
	if loaded.ChunkPlan.Chunks[0].CompletionCriteria[0] != "spec-shaped loader and schema align" {
		t.Fatalf("expected completion criteria list, got %#v", loaded.ChunkPlan.Chunks[0].CompletionCriteria)
	}
	if loaded.ChunkPlan.Chunks[1].DependsOn[0] != "checker-config-contract" {
		t.Fatalf("expected explicit depends_on, got %#v", loaded.ChunkPlan.Chunks[1].DependsOn)
	}
}

func TestLoadConfigRejectsChunkPlanMissingDependsOn(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("tracked", "docs")
	cfg["chunk_plan"] = map[string]any{
		"enabled": true,
		"chunks": []map[string]any{
			{
				"id":                  "checker-config-contract",
				"scope":               "contract surface",
				"completion_criteria": []string{"spec-shaped loader and schema align"},
			},
		},
	}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan.chunks[0].depends_on") {
		t.Fatalf("expected chunk depends_on error, got %v", err)
	}
}

func TestLoadConfigRejectsChunkPlanForwardDependency(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("tracked", "docs")
	cfg["chunk_plan"] = map[string]any{
		"enabled": true,
		"chunks": []map[string]any{
			{
				"id":                  "schema-example-sync",
				"scope":               "references",
				"completion_criteria": []string{"example and schema match the contract"},
				"depends_on":          []string{"checker-config-contract"},
			},
			{
				"id":                  "checker-config-contract",
				"scope":               "contract surface",
				"completion_criteria": []string{"spec-shaped loader and schema align"},
				"depends_on":          []string{},
			},
		},
	}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "must reference prior chunk") {
		t.Fatalf("expected prior dependency error, got %v", err)
	}
}

func TestLoadConfigRejectsChunkPlanChunksNull(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "config.json", `{
  "contract_version": 1,
  "mode": "tracked",
  "profiles": ["go"],
  "docs_root": "docs",
  "ownership_policy": {
    "portable_skill_authoring_owner": "skill-builder",
    "domain_knowledge_owner": "domain-skills",
    "repo_local_skills": {
      "allowed": false,
      "placement_roots": [".claude/skills", ".codex/skills"],
      "authoring_owner": "skill-builder",
      "requires_justification": true
    }
  },
  "chunk_plan": {
    "enabled": true,
    "chunks": null
  }
}`)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan.chunks") || !strings.Contains(err.Error(), "null") {
		t.Fatalf("expected chunk_plan.chunks null error, got %v", err)
	}
}

func TestLoadConfigRejectsEmptyChunkPlanObject(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("tracked", "docs")
	cfg["chunk_plan"] = map[string]any{}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan.enabled") {
		t.Fatalf("expected chunk_plan enabled error, got %v", err)
	}
}

func TestLoadConfigAcceptsOverlayContract(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["mirror_policy"] = map[string]any{
		"mode":  "mirrored",
		"files": []string{"CLAUDE.md", "AGENTS.md"},
	}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := loadConfig(root, "config.json")
	if err != nil {
		t.Fatalf("expected overlay contract to load, got %v", err)
	}
	if loaded.Mode != "overlay" {
		t.Fatalf("expected overlay mode, got %q", loaded.Mode)
	}
	if loaded.DocsRoot != ".docs" {
		t.Fatalf("expected .docs root, got %q", loaded.DocsRoot)
	}
	if loaded.MirrorPolicy.Mode != "mirrored" {
		t.Fatalf("expected mirror policy to be loaded, got %#v", loaded.MirrorPolicy)
	}
}

func TestLoadConfigRejectsEmptyMirrorPolicyObject(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["mirror_policy"] = map[string]any{}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mirror_policy.mode") {
		t.Fatalf("expected mirror_policy mode error, got %v", err)
	}
}

func TestLoadConfigRejectsEmptyCanonicalPointerMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["canonical_pointer_mode"] = ""
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "canonical_pointer_mode") {
		t.Fatalf("expected canonical_pointer_mode error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownCanonicalPointerMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["canonical_pointer_mode"] = "fallback"
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "canonical_pointer_mode must be all or any") {
		t.Fatalf("expected canonical_pointer_mode enum error, got %v", err)
	}
}

func TestLoadConfigRejectsRequiredFilesEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["required_files"] = []string{"Taskfile.yml", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "required_files[1]") {
		t.Fatalf("expected required_files item error, got %v", err)
	}
}

func TestLoadConfigRejectsCanonicalPointerEmptyText(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["canonical_pointers"] = []map[string]any{
		{
			"name": "claude-mirror-check",
			"file": "CLAUDE.md",
			"text": "",
		},
	}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "canonical_pointers[0].text") {
		t.Fatalf("expected canonical_pointers text error, got %v", err)
	}
}

func TestLoadConfigRejectsEmptyInvariantContractFile(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["invariant_contract"] = map[string]any{
		"required": false,
		"file":     "",
	}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := loadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "invariant_contract.file") {
		t.Fatalf("expected invariant_contract.file error, got %v", err)
	}
}

func TestRunJSONModeUsesContractCheckerTerminology(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeJSONConfig(t, root, ".convention-engineering.json", baseContract("tracked", "docs"))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(root, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}
	if strings.Contains(out.String(), "readiness") || strings.Contains(err.String(), "readiness") {
		t.Fatalf("expected contract checker terminology, stdout=%q stderr=%q", out.String(), err.String())
	}
}

func TestSelfHostedRepoRootDefaultInvocationPasses(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(repoRoot, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected self-hosted repo root invocation to pass, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	report := jsonReport{}
	if unmarshalErr := json.Unmarshal(out.Bytes(), &report); unmarshalErr != nil {
		t.Fatalf("expected valid json report: %v", unmarshalErr)
	}
	if !strings.HasSuffix(report.ConfigPath, ".convention-engineering.json") {
		t.Fatalf("expected tracked contract config path, got %q", report.ConfigPath)
	}
	if report.Failed != 0 {
		t.Fatalf("expected no self-hosted contract failures, got %#v", report)
	}
}

func TestSelfHostedPackageDirectoryCLIInvocationPasses(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))

	cmd := exec.Command("go", "run", "./.claude/skills/convention-engineering/scripts", "--repo-root", repoRoot, "--json")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected package-directory CLI invocation to pass, got %v output=%s", err, string(output))
	}

	report := jsonReport{}
	if unmarshalErr := json.Unmarshal(output, &report); unmarshalErr != nil {
		t.Fatalf("expected valid json report from package-directory CLI, got %v output=%s", unmarshalErr, string(output))
	}
	if !strings.HasSuffix(report.ConfigPath, ".convention-engineering.json") || report.Failed != 0 {
		t.Fatalf("expected self-hosted package CLI success, got %#v", report)
	}
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
				"placement_roots":        []string{".claude/skills", ".codex/skills"},
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

func hasFailure(results []checkResult, name string) bool {
	for _, r := range results {
		if r.Name == name && !r.Passed {
			return true
		}
	}
	return false
}
