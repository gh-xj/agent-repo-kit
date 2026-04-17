package contract

import (
	"strings"
	"testing"
)

func TestLoadConfigRejectsMissingContractVersion(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "contract_version")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "contract_version") {
		t.Fatalf("expected contract_version error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownMajorVersion(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["contract_version"] = 2
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "unsupported contract_version major") {
		t.Fatalf("expected unknown major version error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["mode"] = "shadow"
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mode") {
		t.Fatalf("expected mode error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingProfiles(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "profiles")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles") {
		t.Fatalf("expected profiles error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingMirrorPolicy(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "mirror_policy")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mirror_policy") {
		t.Fatalf("expected mirror_policy error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingEvaluationInputs(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "evaluation_inputs")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "evaluation_inputs") {
		t.Fatalf("expected evaluation_inputs error, got %v", err)
	}
}

func TestLoadConfigRejectsMissingChunkPlan(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	delete(cfg, "chunk_plan")
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan") {
		t.Fatalf("expected chunk_plan error, got %v", err)
	}
}

func TestLoadConfigRejectsProfilesEmptyArray(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["profiles"] = []string{}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles") {
		t.Fatalf("expected profiles non-empty error, got %v", err)
	}
}

func TestLoadConfigRejectsProfileEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["profiles"] = []string{"go", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "profiles[1]") {
		t.Fatalf("expected profiles item error, got %v", err)
	}
}

func TestLoadConfigRejectsDocsRootWithTrailingWhitespace(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["docs_root"] = ".docs "
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "docs_root") {
		t.Fatalf("expected exact docs_root enum error, got %v", err)
	}
}

func TestLoadConfigAcceptsAdditiveTopLevelFieldForV1(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["future_top_level_field"] = map[string]any{"mode": "reserved"}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "ownership_policy.repo_local_skills.placement_roots") {
		t.Fatalf("expected repo_local_skills placement_roots error, got %v", err)
	}
}

func TestLoadConfigRejectsPlacementRootEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["ownership_policy"].(map[string]any)["repo_local_skills"].(map[string]any)["placement_roots"] = []string{".claude/skills", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "placement_roots[1]") {
		t.Fatalf("expected placement_roots item error, got %v", err)
	}
}

func TestLoadConfigAcceptsSpecShapedEvaluationInputsObject(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["evaluation_inputs"] = map[string]any{"repo_risk": "standard"}
	writeJSONConfig(t, root, "config.json", cfg)

	loaded, _, err := LoadConfig(root, "config.json")
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
      "placement_roots": [".claude/skills", ".agents/skills"],
      "authoring_owner": "skill-builder",
      "requires_justification": true
    }
  },
  "evaluation_inputs": null
}`)

	_, _, err := LoadConfig(root, "config.json")
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

			_, _, err := LoadConfig(root, "config.json")
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
				"required": false,
				tt.field:   nil,
			}
			writeJSONConfig(t, root, "config.json", cfg)

			_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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

	loaded, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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
      "placement_roots": [".claude/skills", ".agents/skills"],
      "authoring_owner": "skill-builder",
      "requires_justification": true
    }
  },
  "chunk_plan": {
    "enabled": true,
    "chunks": null
  }
}`)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "chunk_plan.chunks") || !strings.Contains(err.Error(), "null") {
		t.Fatalf("expected chunk_plan.chunks null error, got %v", err)
	}
}

func TestLoadConfigRejectsEmptyChunkPlanObject(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("tracked", "docs")
	cfg["chunk_plan"] = map[string]any{}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
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

	loaded, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "mirror_policy.mode") {
		t.Fatalf("expected mirror_policy mode error, got %v", err)
	}
}

func TestLoadConfigRejectsEmptyCanonicalPointerMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["canonical_pointer_mode"] = ""
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "canonical_pointer_mode") {
		t.Fatalf("expected canonical_pointer_mode error, got %v", err)
	}
}

func TestLoadConfigRejectsUnknownCanonicalPointerMode(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["canonical_pointer_mode"] = "fallback"
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "canonical_pointer_mode must be all or any") {
		t.Fatalf("expected canonical_pointer_mode enum error, got %v", err)
	}
}

func TestLoadConfigRejectsRequiredFilesEmptyString(t *testing.T) {
	root := t.TempDir()
	cfg := baseContract("overlay", ".docs")
	cfg["required_files"] = []string{"Taskfile.yml", ""}
	writeJSONConfig(t, root, "config.json", cfg)

	_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
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

	_, _, err := LoadConfig(root, "config.json")
	if err == nil || !strings.Contains(err.Error(), "invariant_contract.file") {
		t.Fatalf("expected invariant_contract.file error, got %v", err)
	}
}
