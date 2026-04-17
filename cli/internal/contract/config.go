package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultConfigFile is the tracked contract file name used by LoadConfig when
// no explicit path is supplied.
const DefaultConfigFile = ".convention-engineering.json"

// CanonicalPointerConfig describes a canonical-pointer check.
type CanonicalPointerConfig struct {
	Name string `json:"name"`
	File string `json:"file"`
	Text string `json:"text"`
}

// ContentCheckConfig describes a content marker check.
type ContentCheckConfig struct {
	Name            string   `json:"name"`
	File            string   `json:"file"`
	RequiredMarkers []string `json:"required_markers"`
}

// GitExcludeCheckConfig describes a git-exclude pattern check.
type GitExcludeCheckConfig struct {
	Name             string   `json:"name"`
	File             string   `json:"file"`
	RequiredPatterns []string `json:"required_patterns"`
}

// InvariantContractConfig describes the invariant contract evaluation.
type InvariantContractConfig struct {
	Required        bool     `json:"required"`
	File            string   `json:"file"`
	RequiredFields  []string `json:"required_fields"`
	ExceptionFields []string `json:"exception_fields"`
}

// RepoLocalSkillsPolicy describes the repo-local skills policy surface.
type RepoLocalSkillsPolicy struct {
	Allowed               bool     `json:"allowed"`
	PlacementRoots        []string `json:"placement_roots"`
	AuthoringOwner        string   `json:"authoring_owner"`
	RequiresJustification bool     `json:"requires_justification"`
}

// OwnershipPolicyConfig describes the ownership policy surface.
type OwnershipPolicyConfig struct {
	PortableSkillAuthoringOwner string                `json:"portable_skill_authoring_owner"`
	DomainKnowledgeOwner        string                `json:"domain_knowledge_owner"`
	RepoLocalSkills             RepoLocalSkillsPolicy `json:"repo_local_skills"`
}

// MirrorPolicyConfig describes the mirror policy surface.
type MirrorPolicyConfig struct {
	Mode  string   `json:"mode"`
	Files []string `json:"files"`
}

// EvaluationInputsConfig describes the evaluation inputs surface.
type EvaluationInputsConfig struct {
	RepoRisk string `json:"repo_risk"`
}

// ChunkDefinition describes a single chunk entry in the plan.
type ChunkDefinition struct {
	ID                 string   `json:"id"`
	Scope              string   `json:"scope"`
	CompletionCriteria []string `json:"completion_criteria"`
	DependsOn          []string `json:"depends_on"`
}

// ChunkPlanConfig describes the chunk plan surface.
type ChunkPlanConfig struct {
	Enabled bool              `json:"enabled"`
	Chunks  []ChunkDefinition `json:"chunks"`
}

type rawRepoLocalSkillsPolicy struct {
	Allowed               *bool    `json:"allowed"`
	PlacementRoots        []string `json:"placement_roots"`
	AuthoringOwner        string   `json:"authoring_owner"`
	RequiresJustification *bool    `json:"requires_justification"`
}

type rawOwnershipPolicyConfig struct {
	PortableSkillAuthoringOwner string                   `json:"portable_skill_authoring_owner"`
	DomainKnowledgeOwner        string                   `json:"domain_knowledge_owner"`
	RepoLocalSkills             rawRepoLocalSkillsPolicy `json:"repo_local_skills"`
}

type rawChunkDefinition struct {
	ID                 *string   `json:"id"`
	Scope              *string   `json:"scope"`
	CompletionCriteria *[]string `json:"completion_criteria"`
	DependsOn          *[]string `json:"depends_on"`
}

type rawChunkPlanConfig struct {
	Enabled *bool                 `json:"enabled"`
	Chunks  *[]rawChunkDefinition `json:"chunks"`
}

type rawCheckConfig struct {
	OwnershipPolicy rawOwnershipPolicyConfig `json:"ownership_policy"`
	ChunkPlan       rawChunkPlanConfig       `json:"chunk_plan"`
}

// Config is the fully validated convention contract surface consumed by the
// checker.
type Config struct {
	ContractVersion      int                      `json:"contract_version"`
	Mode                 string                   `json:"mode"`
	Profiles             []string                 `json:"profiles"`
	DocsRoot             string                   `json:"docs_root"`
	OwnershipPolicy      OwnershipPolicyConfig    `json:"ownership_policy"`
	MirrorPolicy         MirrorPolicyConfig       `json:"mirror_policy"`
	EvaluationInputs     EvaluationInputsConfig   `json:"evaluation_inputs"`
	ChunkPlan            ChunkPlanConfig          `json:"chunk_plan"`
	RequiredFiles        []string                 `json:"required_files"`
	TaskfileChecks       map[string][]string      `json:"taskfile_checks"`
	CanonicalPointerMode string                   `json:"canonical_pointer_mode"`
	CanonicalPointers    []CanonicalPointerConfig `json:"canonical_pointers"`
	ContentChecks        []ContentCheckConfig     `json:"content_checks"`
	GitExcludeChecks     []GitExcludeCheckConfig  `json:"git_exclude_checks"`
	InvariantContract    InvariantContractConfig  `json:"invariant_contract"`
}

// LoadConfig resolves and parses the convention contract config rooted at
// root. When explicitPath is empty, the tracked contract at
// DefaultConfigFile is required. Returns the loaded Config, the resolved
// absolute config path, and any error encountered.
func LoadConfig(root, explicitPath string) (Config, string, error) {
	cfg := Config{}
	rawCfg := rawCheckConfig{}
	rawFields := map[string]json.RawMessage{}
	path := explicitPath
	if path == "" {
		path = filepath.Join(root, DefaultConfigFile)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return Config{}, "", fmt.Errorf("tracked contract not found: %s", path)
			}
			return Config{}, "", err
		}
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, "", err
	}

	decoder := json.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, "", err
	}
	if err := json.Unmarshal(content, &rawCfg); err != nil {
		return Config{}, "", err
	}
	if err := json.Unmarshal(content, &rawFields); err != nil {
		return Config{}, "", err
	}
	if err := rejectExplicitNullContractFields(rawFields); err != nil {
		return Config{}, "", err
	}

	applyDefaultConfigFields(&cfg, rawFields)
	if err := validateConfig(cfg, rawCfg, rawFields); err != nil {
		return Config{}, "", err
	}
	return cfg, path, nil
}

func defaultConfig() Config {
	return Config{
		ContractVersion: 1,
		Mode:            "tracked",
		Profiles:        []string{},
		DocsRoot:        "docs",
		// Owner and placement_roots defaults are intentionally empty so this kit
		// is harness-agnostic. Adopters configure these via
		// .convention-engineering.json based on their harness layout. Common
		// placement_roots values are:
		//   - Claude Code harness: ".claude/skills"
		//   - Codex/OpenAI harness: ".agents/skills"
		//   - Custom/mixed: list every directory your harness loads skills from
		// Common owner values are free-form labels identifying the
		// team/role responsible (e.g. "platform-team", "docs-wg", "skill-builder").
		OwnershipPolicy: OwnershipPolicyConfig{
			PortableSkillAuthoringOwner: "",
			DomainKnowledgeOwner:        "",
			RepoLocalSkills: RepoLocalSkillsPolicy{
				Allowed:               false,
				PlacementRoots:        []string{},
				AuthoringOwner:        "",
				RequiresJustification: true,
			},
		},
		EvaluationInputs:     EvaluationInputsConfig{},
		ChunkPlan:            ChunkPlanConfig{Chunks: []ChunkDefinition{}},
		RequiredFiles:        []string{"Taskfile.yml"},
		TaskfileChecks:       map[string][]string{},
		CanonicalPointerMode: "all",
		CanonicalPointers:    []CanonicalPointerConfig{},
		ContentChecks:        []ContentCheckConfig{},
		GitExcludeChecks:     []GitExcludeCheckConfig{},
		InvariantContract: InvariantContractConfig{
			Required: false,
			File:     ".claude/architecture/invariants.yaml",
			RequiredFields: []string{
				"id",
				"statement",
				"status",
				"evidence",
				"enforceability",
				"owner",
				"last_verified",
			},
			ExceptionFields: []string{
				"exception_reason",
				"exception_owner",
				"exception_expires",
			},
		},
	}
}

func applyDefaultConfigFields(cfg *Config, rawFields map[string]json.RawMessage) {
	defaults := defaultConfig()
	if !hasJSONField(rawFields, "required_files") {
		cfg.RequiredFiles = defaults.RequiredFiles
	}
	if !hasJSONField(rawFields, "taskfile_checks") {
		cfg.TaskfileChecks = defaults.TaskfileChecks
	}
	if !hasJSONField(rawFields, "canonical_pointer_mode") {
		cfg.CanonicalPointerMode = defaults.CanonicalPointerMode
	}
	if !hasJSONField(rawFields, "canonical_pointers") {
		cfg.CanonicalPointers = defaults.CanonicalPointers
	}
	if !hasJSONField(rawFields, "content_checks") {
		cfg.ContentChecks = defaults.ContentChecks
	}
	if !hasJSONField(rawFields, "git_exclude_checks") {
		cfg.GitExcludeChecks = defaults.GitExcludeChecks
	}
	invariantFields := decodeRawObjectFields(rawFields["invariant_contract"])
	if !hasJSONField(rawFields, "invariant_contract") || !hasJSONField(invariantFields, "file") {
		cfg.InvariantContract.File = defaults.InvariantContract.File
	}
	if !hasJSONField(rawFields, "invariant_contract") || !hasJSONField(invariantFields, "required_fields") {
		cfg.InvariantContract.RequiredFields = defaults.InvariantContract.RequiredFields
	}
	if !hasJSONField(rawFields, "invariant_contract") || !hasJSONField(invariantFields, "exception_fields") {
		cfg.InvariantContract.ExceptionFields = defaults.InvariantContract.ExceptionFields
	}
}

func validateConfig(cfg Config, rawCfg rawCheckConfig, rawFields map[string]json.RawMessage) error {
	switch cfg.ContractVersion {
	case 1:
	default:
		if cfg.ContractVersion == 0 {
			return fmt.Errorf("contract_version is required")
		}
		return fmt.Errorf("unsupported contract_version major: %d", cfg.ContractVersion)
	}

	switch cfg.Mode {
	case "tracked", "overlay":
	default:
		return fmt.Errorf("mode must be tracked or overlay")
	}

	switch cfg.DocsRoot {
	case "docs", ".docs":
	default:
		return fmt.Errorf("docs_root must be docs or .docs")
	}

	if cfg.Profiles == nil {
		return fmt.Errorf("profiles is required")
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("profiles must contain at least one entry")
	}
	for i, profile := range cfg.Profiles {
		if strings.TrimSpace(profile) == "" {
			return fmt.Errorf("profiles[%d] must not be empty", i)
		}
	}
	if strings.TrimSpace(cfg.OwnershipPolicy.PortableSkillAuthoringOwner) == "" {
		return fmt.Errorf("ownership_policy.portable_skill_authoring_owner is required")
	}
	if strings.TrimSpace(cfg.OwnershipPolicy.DomainKnowledgeOwner) == "" {
		return fmt.Errorf("ownership_policy.domain_knowledge_owner is required")
	}
	if strings.TrimSpace(cfg.OwnershipPolicy.RepoLocalSkills.AuthoringOwner) == "" {
		return fmt.Errorf("ownership_policy.repo_local_skills.authoring_owner is required")
	}
	if len(cfg.OwnershipPolicy.RepoLocalSkills.PlacementRoots) == 0 {
		return fmt.Errorf("ownership_policy.repo_local_skills.placement_roots is required")
	}
	for i, root := range cfg.OwnershipPolicy.RepoLocalSkills.PlacementRoots {
		if strings.TrimSpace(root) == "" {
			return fmt.Errorf("ownership_policy.repo_local_skills.placement_roots[%d] must not be empty", i)
		}
	}
	if rawCfg.OwnershipPolicy.RepoLocalSkills.Allowed == nil {
		return fmt.Errorf("ownership_policy.repo_local_skills.allowed is required")
	}
	if rawCfg.OwnershipPolicy.RepoLocalSkills.RequiresJustification == nil {
		return fmt.Errorf("ownership_policy.repo_local_skills.requires_justification is required")
	}
	if strings.TrimSpace(cfg.CanonicalPointerMode) == "" {
		return fmt.Errorf("canonical_pointer_mode must not be empty")
	}
	switch cfg.CanonicalPointerMode {
	case "all", "any":
	default:
		return fmt.Errorf("canonical_pointer_mode must be all or any")
	}
	if strings.TrimSpace(cfg.InvariantContract.File) == "" {
		return fmt.Errorf("invariant_contract.file must not be empty")
	}
	if err := validateRequiredFiles(cfg.RequiredFiles); err != nil {
		return err
	}
	if err := validateTaskfileChecks(cfg.TaskfileChecks); err != nil {
		return err
	}
	if err := validateCanonicalPointers(cfg.CanonicalPointers); err != nil {
		return err
	}
	if err := validateContentChecks(cfg.ContentChecks); err != nil {
		return err
	}
	if err := validateGitExcludeChecks(cfg.GitExcludeChecks); err != nil {
		return err
	}
	if err := validateInvariantContractFields(cfg.InvariantContract); err != nil {
		return err
	}
	if !hasJSONField(rawFields, "mirror_policy") {
		return fmt.Errorf("mirror_policy is required")
	}
	if !hasJSONField(rawFields, "evaluation_inputs") {
		return fmt.Errorf("evaluation_inputs is required")
	}
	if !hasJSONField(rawFields, "chunk_plan") {
		return fmt.Errorf("chunk_plan is required")
	}
	if err := validateMirrorPolicy(cfg.MirrorPolicy, hasJSONField(rawFields, "mirror_policy")); err != nil {
		return err
	}
	if err := validateEvaluationInputs(cfg.EvaluationInputs, decodeRawObjectFields(rawFields["evaluation_inputs"])); err != nil {
		return err
	}

	if err := validateChunkPlan(cfg.ChunkPlan, rawCfg.ChunkPlan, hasJSONField(rawFields, "chunk_plan")); err != nil {
		return err
	}
	return nil
}

func rejectExplicitNullContractFields(rawFields map[string]json.RawMessage) error {
	for _, field := range []string{
		"mirror_policy",
		"evaluation_inputs",
		"chunk_plan",
		"required_files",
		"taskfile_checks",
		"canonical_pointers",
		"content_checks",
		"git_exclude_checks",
		"invariant_contract",
	} {
		if isExplicitNullJSON(rawFields[field]) {
			return fmt.Errorf("%s must not be null", field)
		}
	}

	if !hasJSONField(rawFields, "chunk_plan") {
		return nil
	}

	chunkPlanFields := map[string]json.RawMessage{}
	if err := json.Unmarshal(rawFields["chunk_plan"], &chunkPlanFields); err != nil {
		return nil
	}
	if isExplicitNullJSON(chunkPlanFields["chunks"]) {
		return fmt.Errorf("chunk_plan.chunks must not be null")
	}

	invariantFields := decodeRawObjectFields(rawFields["invariant_contract"])
	if isExplicitNullJSON(invariantFields["required_fields"]) {
		return fmt.Errorf("invariant_contract.required_fields must not be null")
	}
	if isExplicitNullJSON(invariantFields["exception_fields"]) {
		return fmt.Errorf("invariant_contract.exception_fields must not be null")
	}
	return nil
}

func hasJSONField(fields map[string]json.RawMessage, key string) bool {
	_, ok := fields[key]
	return ok
}

func isExplicitNullJSON(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

func validateMirrorPolicy(policy MirrorPolicyConfig, present bool) error {
	if !present {
		return nil
	}
	if strings.TrimSpace(policy.Mode) == "" {
		return fmt.Errorf("mirror_policy.mode is required")
	}
	if len(policy.Files) == 0 {
		return fmt.Errorf("mirror_policy.files is required")
	}
	for i, file := range policy.Files {
		if strings.TrimSpace(file) == "" {
			return fmt.Errorf("mirror_policy.files[%d] must not be empty", i)
		}
	}
	return nil
}

func validateRequiredFiles(requiredFiles []string) error {
	for i, rel := range requiredFiles {
		if strings.TrimSpace(rel) == "" {
			return fmt.Errorf("required_files[%d] must not be empty", i)
		}
	}
	return nil
}

func validateTaskfileChecks(taskfileChecks map[string][]string) error {
	for rel, tokens := range taskfileChecks {
		if strings.TrimSpace(rel) == "" {
			return fmt.Errorf("taskfile_checks keys must not be empty")
		}
		for i, token := range tokens {
			if strings.TrimSpace(token) == "" {
				return fmt.Errorf("taskfile_checks[%q][%d] must not be empty", rel, i)
			}
		}
	}
	return nil
}

func validateCanonicalPointers(pointers []CanonicalPointerConfig) error {
	for i, pointer := range pointers {
		if strings.TrimSpace(pointer.Name) == "" {
			return fmt.Errorf("canonical_pointers[%d].name is required", i)
		}
		if strings.TrimSpace(pointer.File) == "" {
			return fmt.Errorf("canonical_pointers[%d].file is required", i)
		}
		if strings.TrimSpace(pointer.Text) == "" {
			return fmt.Errorf("canonical_pointers[%d].text is required", i)
		}
	}
	return nil
}

func validateContentChecks(checks []ContentCheckConfig) error {
	for i, check := range checks {
		if strings.TrimSpace(check.Name) == "" {
			return fmt.Errorf("content_checks[%d].name is required", i)
		}
		if strings.TrimSpace(check.File) == "" {
			return fmt.Errorf("content_checks[%d].file is required", i)
		}
		if len(check.RequiredMarkers) == 0 {
			return fmt.Errorf("content_checks[%d].required_markers is required", i)
		}
		for j, marker := range check.RequiredMarkers {
			if strings.TrimSpace(marker) == "" {
				return fmt.Errorf("content_checks[%d].required_markers[%d] must not be empty", i, j)
			}
		}
	}
	return nil
}

func validateGitExcludeChecks(checks []GitExcludeCheckConfig) error {
	for i, check := range checks {
		if strings.TrimSpace(check.Name) == "" {
			return fmt.Errorf("git_exclude_checks[%d].name is required", i)
		}
		if strings.TrimSpace(check.File) == "" {
			return fmt.Errorf("git_exclude_checks[%d].file is required", i)
		}
		if len(check.RequiredPatterns) == 0 {
			return fmt.Errorf("git_exclude_checks[%d].required_patterns is required", i)
		}
		for j, pattern := range check.RequiredPatterns {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("git_exclude_checks[%d].required_patterns[%d] must not be empty", i, j)
			}
		}
	}
	return nil
}

func validateInvariantContractFields(contract InvariantContractConfig) error {
	for i, field := range contract.RequiredFields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("invariant_contract.required_fields[%d] must not be empty", i)
		}
	}
	for i, field := range contract.ExceptionFields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("invariant_contract.exception_fields[%d] must not be empty", i)
		}
	}
	return nil
}

func validateEvaluationInputs(inputs EvaluationInputsConfig, rawFields map[string]json.RawMessage) error {
	if hasJSONField(rawFields, "repo_risk") && strings.TrimSpace(inputs.RepoRisk) == "" {
		return fmt.Errorf("evaluation_inputs.repo_risk must not be empty")
	}
	return nil
}

var stableChunkIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateChunkPlan(plan ChunkPlanConfig, rawPlan rawChunkPlanConfig, present bool) error {
	if !present {
		return nil
	}
	if rawPlan.Enabled == nil {
		return fmt.Errorf("chunk_plan.enabled is required")
	}
	if rawPlan.Chunks == nil {
		return fmt.Errorf("chunk_plan.chunks is required")
	}

	seen := make(map[string]bool, len(plan.Chunks))
	for i, chunk := range plan.Chunks {
		if i >= len(*rawPlan.Chunks) {
			return fmt.Errorf("chunk_plan.chunks[%d] is required", i)
		}
		rawChunk := (*rawPlan.Chunks)[i]
		if rawChunk.Scope == nil || strings.TrimSpace(chunk.Scope) == "" {
			return fmt.Errorf("chunk_plan.chunks[%d].scope is required", i)
		}
		if rawChunk.CompletionCriteria == nil || len(chunk.CompletionCriteria) == 0 {
			return fmt.Errorf("chunk_plan.chunks[%d].completion_criteria is required", i)
		}
		for j, criterion := range chunk.CompletionCriteria {
			if strings.TrimSpace(criterion) == "" {
				return fmt.Errorf("chunk_plan.chunks[%d].completion_criteria[%d] must not be empty", i, j)
			}
		}
		if rawChunk.DependsOn == nil {
			return fmt.Errorf("chunk_plan.chunks[%d].depends_on is required", i)
		}
		if !stableChunkIDPattern.MatchString(chunk.ID) {
			return fmt.Errorf("chunk_plan.chunks ids must be stable kebab-case: %q", chunk.ID)
		}
		if seen[chunk.ID] {
			return fmt.Errorf("chunk_plan.chunks contains duplicate id %q", chunk.ID)
		}
		for _, dep := range chunk.DependsOn {
			if !seen[dep] {
				return fmt.Errorf("chunk_plan.chunks %q dependency %q must reference prior chunk", chunk.ID, dep)
			}
		}
		seen[chunk.ID] = true
	}
	return nil
}

func decodeRawObjectFields(raw json.RawMessage) map[string]json.RawMessage {
	if len(raw) == 0 || isExplicitNullJSON(raw) {
		return nil
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil
	}
	return fields
}
