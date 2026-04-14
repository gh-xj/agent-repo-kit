// Convention Engineering Contract Checker
//
// Sections:
//   - Types & Config       (Config, CheckResult, output structs)
//   - Config Loading        (loadConfig, resolveIncludes)
//   - Check Runners         (runRequiredFileChecks, runTaskfileChecks, runContentChecks,
//                            runCanonicalPointerChecks, runGitExcludeChecks, runInvariantChecks)
//   - Taskfile Parsing      (resolveTaskfileIncludes, extractTaskTokens)
//   - Invariant Parsing     (parseInvariantChecks, evaluateInvariant)
//   - Output & Main         (printResults, printJSON, main)

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const defaultConfigFile = ".convention-engineering.json"

type checkResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

type canonicalPointerConfig struct {
	Name string `json:"name"`
	File string `json:"file"`
	Text string `json:"text"`
}

type contentCheckConfig struct {
	Name            string   `json:"name"`
	File            string   `json:"file"`
	RequiredMarkers []string `json:"required_markers"`
}

type gitExcludeCheckConfig struct {
	Name             string   `json:"name"`
	File             string   `json:"file"`
	RequiredPatterns []string `json:"required_patterns"`
}

type invariantContractConfig struct {
	Required        bool     `json:"required"`
	File            string   `json:"file"`
	RequiredFields  []string `json:"required_fields"`
	ExceptionFields []string `json:"exception_fields"`
}

type repoLocalSkillsPolicy struct {
	Allowed               bool     `json:"allowed"`
	PlacementRoots        []string `json:"placement_roots"`
	AuthoringOwner        string   `json:"authoring_owner"`
	RequiresJustification bool     `json:"requires_justification"`
}

type ownershipPolicyConfig struct {
	PortableSkillAuthoringOwner string                `json:"portable_skill_authoring_owner"`
	DomainKnowledgeOwner        string                `json:"domain_knowledge_owner"`
	RepoLocalSkills             repoLocalSkillsPolicy `json:"repo_local_skills"`
}

type mirrorPolicyConfig struct {
	Mode  string   `json:"mode"`
	Files []string `json:"files"`
}

type evaluationInputsConfig struct {
	RepoRisk string `json:"repo_risk"`
}

type chunkDefinition struct {
	ID                 string   `json:"id"`
	Scope              string   `json:"scope"`
	CompletionCriteria []string `json:"completion_criteria"`
	DependsOn          []string `json:"depends_on"`
}

type chunkPlanConfig struct {
	Enabled bool              `json:"enabled"`
	Chunks  []chunkDefinition `json:"chunks"`
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

type checkConfig struct {
	ContractVersion      int                      `json:"contract_version"`
	Mode                 string                   `json:"mode"`
	Profiles             []string                 `json:"profiles"`
	DocsRoot             string                   `json:"docs_root"`
	OwnershipPolicy      ownershipPolicyConfig    `json:"ownership_policy"`
	MirrorPolicy         mirrorPolicyConfig       `json:"mirror_policy"`
	EvaluationInputs     evaluationInputsConfig   `json:"evaluation_inputs"`
	ChunkPlan            chunkPlanConfig          `json:"chunk_plan"`
	RequiredFiles        []string                 `json:"required_files"`
	TaskfileChecks       map[string][]string      `json:"taskfile_checks"`
	CanonicalPointerMode string                   `json:"canonical_pointer_mode"`
	CanonicalPointers    []canonicalPointerConfig `json:"canonical_pointers"`
	ContentChecks        []contentCheckConfig     `json:"content_checks"`
	GitExcludeChecks     []gitExcludeCheckConfig  `json:"git_exclude_checks"`
	InvariantContract    invariantContractConfig  `json:"invariant_contract"`
}

type jsonReport struct {
	RepoRoot   string        `json:"repo_root"`
	ConfigPath string        `json:"config_path"`
	Failed     int           `json:"failed"`
	Results    []checkResult `json:"results"`
}

func main() {
	repoRoot := flag.String("repo-root", ".", "Path to repository root")
	configPath := flag.String("config", "", "Path to contract checker config JSON (defaults to required tracked contract .convention-engineering.json)")
	jsonMode := flag.Bool("json", false, "Output machine-readable JSON")
	orchestrate := flag.Bool("orchestrate", false, "Write convention handoff artifacts and launch convention evaluation")
	topic := flag.String("topic", "convention-run", "Stable topic label for orchestration artifacts")
	scope := flag.String("scope", orchestrationScopeFinal, "Evaluation scope for orchestration: final or chunk")
	chunkID := flag.String("chunk-id", "", "Chunk id for chunk-scoped orchestration")
	generatedArtifacts := flag.String("generated-artifacts", "", "Comma-separated repo-relative artifact paths under review")
	parentInvocationID := flag.String("parent-invocation-id", "manual", "Parent invocation id for orchestration launch receipts")
	evaluatorPath := flag.String("evaluator-path", "", "Path to convention-evaluator scripts dir or main.go (falls back to $CONVENTION_EVALUATOR_PATH / $EVALUATOR_SCRIPT_PATH, sibling layout, then .claude/skills/convention-evaluator/scripts/main.go)")
	// Portable alias of --evaluator-path for standalone installs. Takes precedence
	// over --evaluator-path if both are supplied, since callers explicitly opting
	// into the portable name signal standalone-kit intent.
	evaluatorScript := flag.String("evaluator-script", "", "Alias for --evaluator-path (portable standalone name)")
	flag.Parse()

	if *orchestrate {
		request := orchestrationRequest{
			Topic:                  *topic,
			ParentInvocationID:     *parentInvocationID,
			RequestedScope:         *scope,
			RequestedChunkID:       *chunkID,
			GeneratedArtifactPaths: parseGeneratedArtifactList(*generatedArtifacts),
		}
		root, err := filepath.Abs(*repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
			os.Exit(2)
		}
		explicitEvaluator := strings.TrimSpace(*evaluatorScript)
		if explicitEvaluator == "" {
			explicitEvaluator = strings.TrimSpace(*evaluatorPath)
		}
		resolvedEvaluator, err := resolveEvaluatorScript(root, explicitEvaluator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve evaluator: %v\n", err)
			os.Exit(2)
		}
		launcher := newProcessEvaluatorLauncher(root, *parentInvocationID, resolvedEvaluator, time.Now)
		os.Exit(runOrchestration(root, *configPath, request, launcher, os.Stdout, os.Stderr, time.Now))
	}

	os.Exit(run(*repoRoot, *configPath, *jsonMode, os.Stdout, os.Stderr))
}

func run(repoRoot, configPath string, jsonMode bool, stdout, stderr io.Writer) int {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve repo root: %v\n", err)
		return 2
	}

	cfg, resolvedConfigPath, err := loadConfig(root, configPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 2
	}

	results := runChecks(root, cfg)
	failed := countFailures(results)

	if jsonMode {
		report := jsonReport{
			RepoRoot:   root,
			ConfigPath: resolvedConfigPath,
			Failed:     failed,
			Results:    results,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(stderr, "failed to encode json: %v\n", err)
			return 2
		}
		if failed > 0 {
			return 1
		}
		return 0
	}

	for _, r := range results {
		if r.Passed {
			fmt.Fprintf(stdout, "[PASS] %s\n", r.Name)
			continue
		}
		fmt.Fprintf(stdout, "[FAIL] %s: %s\n", r.Name, r.Detail)
	}

	if failed > 0 {
		fmt.Fprintf(stdout, "\nConvention contract failed: %d check(s) failed.\n", failed)
		return 1
	}

	fmt.Fprintln(stdout, "\nConvention contract passed.")
	return 0
}

func loadConfig(root, explicitPath string) (checkConfig, string, error) {
	cfg := checkConfig{}
	rawCfg := rawCheckConfig{}
	rawFields := map[string]json.RawMessage{}
	path := explicitPath
	if path == "" {
		path = filepath.Join(root, defaultConfigFile)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return checkConfig{}, "", fmt.Errorf("tracked contract not found: %s", path)
			}
			return checkConfig{}, "", err
		}
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return checkConfig{}, "", err
	}

	decoder := json.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&cfg); err != nil {
		return checkConfig{}, "", err
	}
	if err := json.Unmarshal(content, &rawCfg); err != nil {
		return checkConfig{}, "", err
	}
	if err := json.Unmarshal(content, &rawFields); err != nil {
		return checkConfig{}, "", err
	}
	if err := rejectExplicitNullContractFields(rawFields); err != nil {
		return checkConfig{}, "", err
	}

	applyDefaultConfigFields(&cfg, rawFields)
	if err := validateConfig(cfg, rawCfg, rawFields); err != nil {
		return checkConfig{}, "", err
	}
	return cfg, path, nil
}

func defaultConfig() checkConfig {
	return checkConfig{
		ContractVersion: 1,
		Mode:            "tracked",
		Profiles:        []string{},
		DocsRoot:        "docs",
		// Owner and placement_roots defaults are intentionally empty so this kit
		// is harness-agnostic. Adopters configure these via
		// .convention-engineering.json based on their harness layout. Common
		// placement_roots values are:
		//   - Claude Code harness: ".claude/skills"
		//   - Codex/OpenAI harness: ".codex/skills"
		//   - Custom/mixed: list every directory your harness loads skills from
		// Common owner values are free-form labels identifying the
		// team/role responsible (e.g. "platform-team", "docs-wg", "skill-builder").
		OwnershipPolicy: ownershipPolicyConfig{
			PortableSkillAuthoringOwner: "",
			DomainKnowledgeOwner:        "",
			RepoLocalSkills: repoLocalSkillsPolicy{
				Allowed:               false,
				PlacementRoots:        []string{},
				AuthoringOwner:        "",
				RequiresJustification: true,
			},
		},
		EvaluationInputs:     evaluationInputsConfig{},
		ChunkPlan:            chunkPlanConfig{Chunks: []chunkDefinition{}},
		RequiredFiles:        []string{"Taskfile.yml"},
		TaskfileChecks:       map[string][]string{},
		CanonicalPointerMode: "all",
		CanonicalPointers:    []canonicalPointerConfig{},
		ContentChecks:        []contentCheckConfig{},
		GitExcludeChecks:     []gitExcludeCheckConfig{},
		InvariantContract: invariantContractConfig{
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

func applyDefaultConfigFields(cfg *checkConfig, rawFields map[string]json.RawMessage) {
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

func validateConfig(cfg checkConfig, rawCfg rawCheckConfig, rawFields map[string]json.RawMessage) error {
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

func validateMirrorPolicy(policy mirrorPolicyConfig, present bool) error {
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

func validateCanonicalPointers(pointers []canonicalPointerConfig) error {
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

func validateContentChecks(checks []contentCheckConfig) error {
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

func validateGitExcludeChecks(checks []gitExcludeCheckConfig) error {
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

func validateInvariantContractFields(contract invariantContractConfig) error {
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

func validateEvaluationInputs(inputs evaluationInputsConfig, rawFields map[string]json.RawMessage) error {
	if hasJSONField(rawFields, "repo_risk") && strings.TrimSpace(inputs.RepoRisk) == "" {
		return fmt.Errorf("evaluation_inputs.repo_risk must not be empty")
	}
	return nil
}

var stableChunkIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateChunkPlan(plan chunkPlanConfig, rawPlan rawChunkPlanConfig, present bool) error {
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

func countFailures(results []checkResult) int {
	failed := 0
	for _, r := range results {
		if !r.Passed {
			failed++
		}
	}
	return failed
}

func runChecks(root string, cfg checkConfig) []checkResult {
	results := make([]checkResult, 0)

	results = append(results, runRequiredFileChecks(root, cfg.RequiredFiles)...)
	results = append(results, runTaskfileChecks(root, cfg.TaskfileChecks)...)
	results = append(results, runCanonicalPointerChecks(root, cfg.CanonicalPointerMode, cfg.CanonicalPointers)...)
	results = append(results, runContentChecks(root, cfg.ContentChecks)...)
	results = append(results, runGitExcludeChecks(root, cfg.GitExcludeChecks)...)
	results = append(results, runInvariantChecks(root, cfg.InvariantContract)...)

	return results
}

func runRequiredFileChecks(root string, requiredFiles []string) []checkResult {
	results := make([]checkResult, 0, len(requiredFiles))
	for _, rel := range requiredFiles {
		full := filepath.Join(root, rel)
		if _, err := os.Stat(full); err != nil {
			results = append(results, checkResult{Name: "file:" + rel, Passed: false, Detail: "missing"})
			continue
		}
		results = append(results, checkResult{Name: "file:" + rel, Passed: true})
	}
	return results
}

func runTaskfileChecks(root string, taskfileChecks map[string][]string) []checkResult {
	if len(taskfileChecks) == 0 {
		return []checkResult{{Name: "taskfile-checks", Passed: true, Detail: "no taskfile checks configured"}}
	}

	results := make([]checkResult, 0)
	for rel, tokens := range taskfileChecks {
		aggregateText, visited, err := aggregateTaskfileText(root, rel)
		if err != nil {
			results = append(results, checkResult{Name: "taskfile:" + rel, Passed: false, Detail: err.Error()})
			continue
		}

		results = append(results, checkResult{
			Name:   "taskfile:" + rel + ":includes",
			Passed: true,
			Detail: fmt.Sprintf("resolved %d taskfile(s)", len(visited)),
		})

		for _, token := range tokens {
			name := "task:" + rel + ":" + token
			if strings.Contains(aggregateText, token) {
				results = append(results, checkResult{Name: name, Passed: true})
				continue
			}
			results = append(results, checkResult{Name: name, Passed: false, Detail: "not found in taskfile include graph"})
		}
	}
	return results
}

var taskfilePathLinePattern = regexp.MustCompile(`(?m)^\s*taskfile:\s*(.+?)\s*$`)

func aggregateTaskfileText(root, relativeTaskfile string) (string, []string, error) {
	start := filepath.Join(root, relativeTaskfile)
	visitedOrder := make([]string, 0)
	visited := map[string]bool{}
	builder := strings.Builder{}

	var walk func(string) error
	walk = func(path string) error {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if visited[absPath] {
			return nil
		}
		visited[absPath] = true
		visitedOrder = append(visitedOrder, absPath)

		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		builder.WriteString("\n# file: ")
		builder.WriteString(absPath)
		builder.WriteString("\n")
		builder.Write(content)
		builder.WriteString("\n")

		for _, includeRel := range extractIncludeTaskfiles(string(content)) {
			if strings.Contains(includeRel, "{{") {
				continue
			}
			nextPath := includeRel
			if !filepath.IsAbs(nextPath) {
				nextPath = filepath.Join(filepath.Dir(absPath), includeRel)
			}
			if err := walk(nextPath); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(start); err != nil {
		return "", nil, fmt.Errorf("cannot resolve taskfile include graph from %s: %w", relativeTaskfile, err)
	}
	return builder.String(), visitedOrder, nil
}

func extractIncludeTaskfiles(content string) []string {
	matches := taskfilePathLinePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value := strings.TrimSpace(match[1])
		value = strings.Trim(value, "\"'")
		if value == "" {
			continue
		}
		paths = append(paths, value)
	}
	return paths
}

func runCanonicalPointerChecks(root, mode string, pointers []canonicalPointerConfig) []checkResult {
	if len(pointers) == 0 {
		return []checkResult{{Name: "canonical-pointers", Passed: true, Detail: "no canonical pointer checks configured"}}
	}

	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	if normalizedMode != "any" {
		normalizedMode = "all"
	}

	passed := make([]string, 0)
	failed := make([]string, 0)
	for _, pointer := range pointers {
		name := pointer.Name
		if name == "" {
			name = fmt.Sprintf("pointer:%s", pointer.File)
		}
		fullPath := filepath.Join(root, pointer.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (cannot read %s)", name, pointer.File))
			continue
		}
		if strings.Contains(string(content), pointer.Text) {
			passed = append(passed, name)
			continue
		}
		failed = append(failed, fmt.Sprintf("%s (text missing)", name))
	}

	if normalizedMode == "any" {
		if len(passed) > 0 {
			return []checkResult{{
				Name:   "canonical-pointers:any",
				Passed: true,
				Detail: fmt.Sprintf("passed=%s", strings.Join(passed, ",")),
			}}
		}
		return []checkResult{{
			Name:   "canonical-pointers:any",
			Passed: false,
			Detail: fmt.Sprintf("no pointer matched; failures=%s", strings.Join(failed, "; ")),
		}}
	}

	if len(failed) == 0 {
		return []checkResult{{
			Name:   "canonical-pointers:all",
			Passed: true,
			Detail: fmt.Sprintf("all pointers matched (%d)", len(passed)),
		}}
	}
	return []checkResult{{
		Name:   "canonical-pointers:all",
		Passed: false,
		Detail: fmt.Sprintf("failures=%s", strings.Join(failed, "; ")),
	}}
}

func runContentChecks(root string, checks []contentCheckConfig) []checkResult {
	if len(checks) == 0 {
		return []checkResult{{Name: "content-checks", Passed: true, Detail: "no content checks configured"}}
	}

	results := make([]checkResult, 0, len(checks))
	for _, check := range checks {
		name := check.Name
		if name == "" {
			name = "content:" + check.File
		}
		fullPath := filepath.Join(root, check.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			results = append(results, checkResult{Name: name, Passed: false, Detail: "cannot read " + check.File})
			continue
		}
		text := string(content)
		missing := make([]string, 0)
		for _, marker := range check.RequiredMarkers {
			if strings.Contains(text, marker) {
				continue
			}
			missing = append(missing, marker)
		}
		if len(missing) == 0 {
			results = append(results, checkResult{Name: name, Passed: true})
			continue
		}
		results = append(results, checkResult{Name: name, Passed: false, Detail: "missing markers: " + strings.Join(missing, ",")})
	}
	return results
}

func runGitExcludeChecks(root string, checks []gitExcludeCheckConfig) []checkResult {
	if len(checks) == 0 {
		return []checkResult{{Name: "git-exclude-checks", Passed: true, Detail: "no git exclude checks configured"}}
	}

	results := make([]checkResult, 0, len(checks))
	for _, check := range checks {
		name := strings.TrimSpace(check.Name)
		if name == "" {
			name = "git-exclude:" + check.File
		}

		fullPath := filepath.Join(root, check.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			results = append(results, checkResult{Name: name, Passed: false, Detail: "cannot read " + check.File})
			continue
		}

		available := parseGitExcludePatterns(string(content))
		missing := make([]string, 0)
		for _, pattern := range check.RequiredPatterns {
			trimmed := strings.TrimSpace(pattern)
			if trimmed == "" {
				continue
			}
			if available[trimmed] {
				continue
			}
			missing = append(missing, trimmed)
		}

		if len(missing) == 0 {
			results = append(results, checkResult{Name: name, Passed: true})
			continue
		}
		results = append(results, checkResult{
			Name:   name,
			Passed: false,
			Detail: "missing patterns: " + strings.Join(missing, ","),
		})
	}
	return results
}

func parseGitExcludePatterns(content string) map[string]bool {
	patterns := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns[line] = true
	}
	return patterns
}

type invariantEntry struct {
	Fields map[string]string
	Index  int
}

func runInvariantChecks(root string, cfg invariantContractConfig) []checkResult {
	name := "invariants:" + cfg.File
	fullPath := filepath.Join(root, cfg.File)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if cfg.Required {
			return []checkResult{{Name: name, Passed: false, Detail: "cannot read " + cfg.File}}
		}
		return []checkResult{{Name: name, Passed: true, Detail: "optional file missing"}}
	}

	entries := parseInvariantEntries(string(content))
	if len(entries) == 0 {
		return []checkResult{{Name: name, Passed: false, Detail: "no invariant entries found"}}
	}

	results := make([]checkResult, 0)
	enforceableCount := 0
	for _, entry := range entries {
		if normalizeFieldValue(entry.Fields["enforceability"]) != "enforceable" {
			continue
		}
		enforceableCount++

		entryID := entry.Fields["id"]
		if isEmptyFieldValue(entryID) {
			entryID = fmt.Sprintf("entry-%d", entry.Index)
		}

		missingRequired := missingFields(entry.Fields, cfg.RequiredFields)
		if len(missingRequired) > 0 {
			results = append(results, checkResult{
				Name:   "invariants:" + entryID + ":required-fields",
				Passed: false,
				Detail: "missing field(s): " + strings.Join(missingRequired, ","),
			})
			continue
		}

		status := normalizeFieldValue(entry.Fields["status"])
		if status != "accepted" && status != "enforced" {
			results = append(results, checkResult{
				Name:   "invariants:" + entryID + ":status",
				Passed: false,
				Detail: "invalid status: " + entry.Fields["status"],
			})
			continue
		}

		anyExceptionField := false
		for _, field := range cfg.ExceptionFields {
			if !isEmptyFieldValue(entry.Fields[field]) {
				anyExceptionField = true
				break
			}
		}

		if anyExceptionField {
			missingException := missingFields(entry.Fields, cfg.ExceptionFields)
			if len(missingException) > 0 {
				results = append(results, checkResult{
					Name:   "invariants:" + entryID + ":exception-fields",
					Passed: false,
					Detail: "missing field(s): " + strings.Join(missingException, ","),
				})
				continue
			}

			expiryRaw := strings.TrimSpace(entry.Fields["exception_expires"])
			expiryAt, dateOnly, parseErr := parseExceptionExpiry(expiryRaw)
			if parseErr != nil {
				results = append(results, checkResult{
					Name:   "invariants:" + entryID + ":exception-expiry",
					Passed: false,
					Detail: parseErr.Error(),
				})
				continue
			}

			if exceptionExpired(expiryAt, dateOnly) {
				results = append(results, checkResult{
					Name:   "invariants:" + entryID + ":exception-expiry",
					Passed: false,
					Detail: "expired on " + expiryRaw,
				})
				continue
			}
		}

		results = append(results, checkResult{Name: "invariants:" + entryID + ":contract", Passed: true})
	}

	if enforceableCount == 0 {
		results = append(results, checkResult{Name: "invariants:enforceable", Passed: true, Detail: "no enforceable invariants"})
	}

	entryFailures := countFailures(results)
	summary := checkResult{Name: name, Passed: entryFailures == 0}
	if entryFailures > 0 {
		summary.Detail = fmt.Sprintf("%d check(s) failed", entryFailures)
	}
	return append([]checkResult{summary}, results...)
}

func parseInvariantEntries(content string) []invariantEntry {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var entries []invariantEntry
	itemIndent := -1
	lastKey := ""
	var current invariantEntry
	active := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := leadingSpaces(line)
		if strings.HasPrefix(trimmed, "- ") {
			if itemIndent == -1 {
				itemIndent = indent
			}
			if indent == itemIndent {
				if active {
					entries = append(entries, current)
				}
				active = true
				lastKey = ""
				current = invariantEntry{Fields: map[string]string{}, Index: len(entries) + 1}

				key, value, ok := splitYAMLField(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
				if ok {
					current.Fields[key] = value
					lastKey = key
				}
				continue
			}

			if active && lastKey != "" && isEmptyFieldValue(current.Fields[lastKey]) {
				current.Fields[lastKey] = "(list)"
			}
			continue
		}

		if !active {
			continue
		}

		key, value, ok := splitYAMLField(trimmed)
		if !ok {
			continue
		}
		current.Fields[key] = value
		lastKey = key
	}

	if active {
		entries = append(entries, current)
	}

	return entries
}

func splitYAMLField(line string) (string, string, bool) {
	colon := strings.Index(line, ":")
	if colon <= 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:colon])
	if key == "" || strings.Contains(key, " ") {
		return "", "", false
	}

	value := strings.TrimSpace(line[colon+1:])
	value = strings.Trim(value, `"'`)
	return key, value, true
}

func leadingSpaces(line string) int {
	spaces := 0
	for _, ch := range line {
		if ch != ' ' {
			break
		}
		spaces++
	}
	return spaces
}

func normalizeFieldValue(value string) string {
	normalized := strings.TrimSpace(value)
	normalized = strings.Trim(normalized, `"'`)
	return strings.ToLower(normalized)
}

func isEmptyFieldValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	if trimmed == "[]" || trimmed == "{}" || strings.EqualFold(trimmed, "null") {
		return true
	}
	return false
}

func missingFields(fields map[string]string, required []string) []string {
	missing := make([]string, 0)
	for _, name := range required {
		if isEmptyFieldValue(fields[name]) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

func parseExceptionExpiry(raw string) (time.Time, bool, error) {
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed, layout == "2006-01-02", nil
		}
	}
	return time.Time{}, false, fmt.Errorf("invalid exception_expires: %q", raw)
}

func exceptionExpired(expiryAt time.Time, dateOnly bool) bool {
	nowUTC := time.Now().UTC()
	if dateOnly {
		todayUTC := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)
		return expiryAt.Before(todayUTC)
	}
	return expiryAt.UTC().Before(nowUTC)
}
