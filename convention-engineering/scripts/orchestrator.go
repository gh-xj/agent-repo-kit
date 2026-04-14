package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	orchestrationScopeFinal = "final"
	orchestrationScopeChunk = "chunk"
)

type orchestratorChunkState struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	HardFailDimensions []string `json:"hard_fail_dimensions"`
	SoftFailDimensions []string `json:"soft_fail_dimensions"`
}

type evidenceRecord struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CWD         string `json:"cwd"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at"`
	ExitCode    int    `json:"exit_code"`
	ToolName    string `json:"tool_name"`
	ToolVersion string `json:"tool_version"`
	StdoutPath  string `json:"stdout_path"`
	StderrPath  string `json:"stderr_path"`
}

type evidenceManifest struct {
	Records []evidenceRecord `json:"records"`
}

type orchestrationHandoffManifest struct {
	ManifestID             string                   `json:"manifest_id"`
	ContractPath           string                   `json:"contract_path"`
	BriefPath              string                   `json:"brief_path"`
	GeneratedArtifactPaths []string                 `json:"generated_artifact_paths"`
	RawEvidenceBundlePath  string                   `json:"raw_evidence_bundle_path"`
	LaunchReceiptPath      string                   `json:"launch_receipt_path"`
	RequestedScope         string                   `json:"requested_scope"`
	RequestedChunkID       string                   `json:"requested_chunk_id,omitempty"`
	CheckerJSONPath        string                   `json:"checker_json_path"`
	ReportPath             string                   `json:"report_path"`
	ResultPath             string                   `json:"result_path"`
	ChunkState             []orchestratorChunkState `json:"chunk_state"`
}

type orchestrationLaunchReceipt struct {
	ParentInvocationID    string `json:"parent_invocation_id"`
	EvaluatorInvocationID string `json:"evaluator_invocation_id"`
	LaunchMode            string `json:"launch_mode"`
	FreshContext          bool   `json:"fresh_context"`
	ForkContext           *bool  `json:"fork_context,omitempty"`
	HandoffManifestID     string `json:"handoff_manifest_id"`
	LaunchedAt            string `json:"launched_at"`
}

type orchestrationEvaluationResult struct {
	Scope              string   `json:"scope"`
	Status             string   `json:"status"`
	ChunkID            string   `json:"chunk_id,omitempty"`
	HardFailDimensions []string `json:"hard_fail_dimensions"`
	SoftFailDimensions []string `json:"soft_fail_dimensions"`
	ReportPath         string   `json:"report_path"`
	CheckerJSONPath    string   `json:"checker_json_path"`
	LaunchReceiptPath  string   `json:"launch_receipt_path"`
	GeneratedAt        string   `json:"generated_at"`
}

type orchestrationRequest struct {
	Topic                 string
	ParentInvocationID    string
	RequestedScope        string
	RequestedChunkID      string
	GeneratedArtifactPaths []string
	ChunkState            []orchestratorChunkState
	BriefBody             string
}

type orchestrationOutcome struct {
	BriefPath          string
	HandoffPath        string
	LaunchReceiptPath  string
	EvidenceBundlePath string
	EvidenceManifest   string
	CheckerJSONPath    string
	ReportPath         string
	ResultPath         string
	Result             orchestrationEvaluationResult
}

type EvaluatorLauncher interface {
	Launch(repoRoot string, handoffPath string) (orchestrationLaunchReceipt, error)
}

type processEvaluatorLauncher struct {
	parentInvocationID string
	now                func() time.Time
	evaluatorScript    string
}

func newProcessEvaluatorLauncher(repoRoot, parentInvocationID, evaluatorScript string, now func() time.Time) processEvaluatorLauncher {
	return processEvaluatorLauncher{
		parentInvocationID: strings.TrimSpace(parentInvocationID),
		now:                now,
		evaluatorScript:    evaluatorScript,
	}
}

// resolveEvaluatorScript determines the absolute path to the convention-evaluator
// main.go in a harness-agnostic way. Resolution order:
//  1. Explicit override (e.g. --evaluator-script / --evaluator-path CLI flag)
//  2. EVALUATOR_SCRIPT_PATH environment variable (portable standalone name)
//  3. CONVENTION_EVALUATOR_PATH environment variable (legacy alias; kept for
//     backward compatibility with existing local setups)
//  4. Sibling lookup relative to this binary/source tree: the standalone
//     agent-repo-kit layout places convention-evaluator next to convention-engineering
//  5. Legacy repo-local path: <repoRoot>/.claude/skills/convention-evaluator/scripts/main.go
//
// Returns an absolute path to main.go (or the scripts directory containing it) that
// exists on disk, or an error describing every attempted location.
func resolveEvaluatorScript(repoRoot, explicit string) (string, error) {
	var attempted []string

	normalize := func(candidate string) (string, bool) {
		if strings.TrimSpace(candidate) == "" {
			return "", false
		}
		abs := candidate
		if !filepath.IsAbs(abs) {
			if resolved, err := filepath.Abs(abs); err == nil {
				abs = resolved
			}
		}
		info, err := os.Stat(abs)
		if err != nil {
			attempted = append(attempted, abs)
			return "", false
		}
		if info.IsDir() {
			candidateFile := filepath.Join(abs, "main.go")
			if _, err := os.Stat(candidateFile); err != nil {
				attempted = append(attempted, candidateFile)
				return "", false
			}
			return candidateFile, true
		}
		return abs, true
	}

	if path, ok := normalize(explicit); ok {
		return path, nil
	}
	if path, ok := normalize(os.Getenv("EVALUATOR_SCRIPT_PATH")); ok {
		return path, nil
	}
	if path, ok := normalize(os.Getenv("CONVENTION_EVALUATOR_PATH")); ok {
		return path, nil
	}

	// Sibling lookup from the directory containing this source/binary. The
	// standalone kit layout is:
	//   agent-repo-kit/
	//     convention-engineering/scripts/<this file>
	//     convention-evaluator/scripts/main.go
	if exe, err := os.Executable(); err == nil {
		siblingFromExe := filepath.Join(filepath.Dir(exe), "..", "..", "..", "convention-evaluator", "scripts", "main.go")
		if path, ok := normalize(siblingFromExe); ok {
			return path, nil
		}
	}
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		siblingFromSource := filepath.Join(filepath.Dir(thisFile), "..", "..", "convention-evaluator", "scripts", "main.go")
		if path, ok := normalize(siblingFromSource); ok {
			return path, nil
		}
	}

	// Legacy repo-local layout (backward compat for adopters using the
	// Claude harness skill placement).
	if strings.TrimSpace(repoRoot) != "" {
		legacy := filepath.Join(repoRoot, ".claude", "skills", "convention-evaluator", "scripts", "main.go")
		if path, ok := normalize(legacy); ok {
			return path, nil
		}
	}

	return "", fmt.Errorf(
		"could not locate convention-evaluator; set --evaluator-script (or --evaluator-path), "+
			"EVALUATOR_SCRIPT_PATH, or CONVENTION_EVALUATOR_PATH (attempted: %s)",
		strings.Join(attempted, ", "),
	)
}

func (l processEvaluatorLauncher) Launch(repoRoot string, handoffPath string) (orchestrationLaunchReceipt, error) {
	manifest := orchestrationHandoffManifest{}
	if err := loadJSONArtifact(repoRoot, handoffPath, &manifest); err != nil {
		return orchestrationLaunchReceipt{}, err
	}

	forkContext := false
	receipt := orchestrationLaunchReceipt{
		ParentInvocationID:    valueOrDefault(l.parentInvocationID, "manual"),
		EvaluatorInvocationID: fmt.Sprintf("convention-evaluator-%d", l.now().UTC().UnixNano()),
		LaunchMode:            "process",
		FreshContext:          true,
		ForkContext:           &forkContext,
		HandoffManifestID:     manifest.ManifestID,
		LaunchedAt:            l.now().UTC().Format(time.RFC3339),
	}
	if err := writeJSONArtifact(repoRoot, manifest.LaunchReceiptPath, receipt); err != nil {
		return orchestrationLaunchReceipt{}, err
	}

	cmd := exec.Command("go", "run", l.evaluatorScript, "--repo-root", repoRoot, "--handoff", handoffPath)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return orchestrationLaunchReceipt{}, err
		}
	}

	return receipt, nil
}

func runOrchestration(repoRoot, configPath string, request orchestrationRequest, launcher EvaluatorLauncher, stdout, stderr io.Writer, now func() time.Time) int {
	outcome, err := orchestrateEvaluation(repoRoot, configPath, request, launcher, now)
	if err != nil {
		fmt.Fprintf(stderr, "failed to orchestrate evaluation: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "status=%s\n", outcome.Result.Status)
	if outcome.Result.Status == "passed" {
		return 0
	}
	return 1
}

func orchestrateEvaluation(repoRoot, configPath string, request orchestrationRequest, launcher EvaluatorLauncher, now func() time.Time) (orchestrationOutcome, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return orchestrationOutcome{}, fmt.Errorf("resolve repo root: %w", err)
	}
	if strings.TrimSpace(request.RequestedScope) == "" {
		request.RequestedScope = orchestrationScopeFinal
	}
	if err := validateOrchestrationRequest(request); err != nil {
		return orchestrationOutcome{}, err
	}

	cfg, resolvedConfigPath, err := loadConfig(root, configPath)
	if err != nil {
		return orchestrationOutcome{}, fmt.Errorf("load contract: %w", err)
	}

	topic := sanitizeTopic(request.Topic)
	datePrefix := now().Format("2006-01-02")
	briefRel := filepath.ToSlash(filepath.Join(cfg.DocsRoot, "planning", fmt.Sprintf("%s_%s_design.md", datePrefix, topic)))
	reviewPrefix := filepath.ToSlash(filepath.Join(cfg.DocsRoot, "reviews", fmt.Sprintf("%s_%s", datePrefix, topic)))
	handoffRel := reviewPrefix + "_handoff.json"
	receiptRel := reviewPrefix + "_launch-receipt.json"
	evidenceDirRel := reviewPrefix + "_evidence"
	evidenceManifestRel := filepath.ToSlash(filepath.Join(evidenceDirRel, "manifest.json"))
	checkerRel := filepath.ToSlash(filepath.Join(evidenceDirRel, "checker.json"))
	checkerStdoutRel := filepath.ToSlash(filepath.Join(evidenceDirRel, "contract-checker.stdout.txt"))
	checkerStderrRel := filepath.ToSlash(filepath.Join(evidenceDirRel, "contract-checker.stderr.txt"))
	reportRel := reviewPrefix + "_convention_eval.md"
	resultRel := reviewPrefix + "_evaluation_result.json"
	contractRel := normalizeRepoPath(root, resolvedConfigPath)
	generatedArtifacts := normalizeGeneratedArtifacts(root, request.GeneratedArtifactPaths)

	briefBody := request.BriefBody
	if strings.TrimSpace(briefBody) == "" {
		briefBody = defaultConventionBrief(topic, contractRel, cfg, request)
	}
	if err := writeTextArtifact(root, briefRel, briefBody); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write brief: %w", err)
	}

	checkerStartedAt := now().UTC()
	checkResults := runChecks(root, cfg)
	checkerFinishedAt := now().UTC()
	checkerReport := jsonReport{
		RepoRoot:   root,
		ConfigPath: resolvedConfigPath,
		Failed:     countFailures(checkResults),
		Results:    checkResults,
	}
	if err := writeJSONArtifact(root, checkerRel, checkerReport); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write checker report: %w", err)
	}
	if err := writeTextArtifact(root, checkerStdoutRel, formatCheckerSummary(checkerReport)); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write checker stdout: %w", err)
	}
	if err := writeTextArtifact(root, checkerStderrRel, ""); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write checker stderr: %w", err)
	}

	evidence := evidenceManifest{
		Records: []evidenceRecord{
			{
				Name:        "contract-checker",
				Command:     "internal:convention-contract-checker",
				CWD:         ".",
				StartedAt:   checkerStartedAt.Format(time.RFC3339),
				FinishedAt:  checkerFinishedAt.Format(time.RFC3339),
				ExitCode:    checkerExitCode(checkerReport.Failed),
				ToolName:    "convention-engineering",
				ToolVersion: "contract-checker-v1",
				StdoutPath:  checkerStdoutRel,
				StderrPath:  checkerStderrRel,
			},
		},
	}
	if err := writeJSONArtifact(root, evidenceManifestRel, evidence); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write evidence manifest: %w", err)
	}

	manifest := orchestrationHandoffManifest{
		ManifestID:             fmt.Sprintf("%s-%d", topic, now().UTC().UnixNano()),
		ContractPath:           contractRel,
		BriefPath:              briefRel,
		GeneratedArtifactPaths: generatedArtifacts,
		RawEvidenceBundlePath:  evidenceDirRel,
		LaunchReceiptPath:      receiptRel,
		RequestedScope:         request.RequestedScope,
		RequestedChunkID:       strings.TrimSpace(request.RequestedChunkID),
		CheckerJSONPath:        checkerRel,
		ReportPath:             reportRel,
		ResultPath:             resultRel,
		ChunkState:             normalizeChunkStates(request.ChunkState),
	}
	if err := writeJSONArtifact(root, handoffRel, manifest); err != nil {
		return orchestrationOutcome{}, fmt.Errorf("write handoff: %w", err)
	}

	outcome := orchestrationOutcome{
		BriefPath:          briefRel,
		HandoffPath:        handoffRel,
		LaunchReceiptPath:  receiptRel,
		EvidenceBundlePath: evidenceDirRel,
		EvidenceManifest:   evidenceManifestRel,
		CheckerJSONPath:    checkerRel,
		ReportPath:         reportRel,
		ResultPath:         resultRel,
	}

	var lastInfraResult *orchestrationEvaluationResult
	for attempt := 0; attempt < 2; attempt++ {
		_, launchErr := launcher.Launch(root, resolveArtifactPath(root, handoffRel))
		if launchErr != nil {
			if attempt == 0 {
				continue
			}
			if lastInfraResult != nil {
				outcome.Result = *lastInfraResult
				return outcome, nil
			}
			return orchestrationOutcome{}, fmt.Errorf("launch evaluator: %w", launchErr)
		}

		result := orchestrationEvaluationResult{}
		if err := loadJSONArtifact(root, resultRel, &result); err != nil {
			if attempt == 0 {
				continue
			}
			if lastInfraResult != nil {
				outcome.Result = *lastInfraResult
				return outcome, nil
			}
			return orchestrationOutcome{}, fmt.Errorf("read evaluation result: %w", err)
		}

		if result.Status == "infrastructure_failed" {
			lastInfraResult = &result
			if attempt == 0 {
				continue
			}
		}

		outcome.Result = result
		return outcome, nil
	}

	if lastInfraResult != nil {
		outcome.Result = *lastInfraResult
		return outcome, nil
	}
	return orchestrationOutcome{}, fmt.Errorf("evaluation did not produce a result")
}

func validateOrchestrationRequest(request orchestrationRequest) error {
	scope := strings.TrimSpace(request.RequestedScope)
	if scope == "" {
		scope = orchestrationScopeFinal
	}
	switch scope {
	case orchestrationScopeFinal, orchestrationScopeChunk:
	default:
		return fmt.Errorf("requested scope must be final or chunk")
	}
	if scope == orchestrationScopeChunk && strings.TrimSpace(request.RequestedChunkID) == "" {
		return fmt.Errorf("requested chunk id is required for chunk scope")
	}

	allowedStates := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"passed":      true,
		"rework":      true,
		"deferred":    true,
	}
	for i, chunk := range request.ChunkState {
		if strings.TrimSpace(chunk.ID) == "" {
			return fmt.Errorf("chunk_state[%d].id is required", i)
		}
		if !allowedStates[chunk.Status] {
			return fmt.Errorf("chunk_state[%d].status must be one of pending, in_progress, passed, rework, deferred", i)
		}
	}
	return nil
}

func normalizeGeneratedArtifacts(root string, paths []string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		rel := normalizeRepoPath(root, path)
		if rel == "" || seen[rel] {
			continue
		}
		seen[rel] = true
		normalized = append(normalized, rel)
	}
	sort.Strings(normalized)
	return normalized
}

func normalizeChunkStates(chunks []orchestratorChunkState) []orchestratorChunkState {
	normalized := make([]orchestratorChunkState, 0, len(chunks))
	for _, chunk := range chunks {
		copyChunk := chunk
		copyChunk.ID = strings.TrimSpace(copyChunk.ID)
		copyChunk.Status = strings.TrimSpace(copyChunk.Status)
		copyChunk.HardFailDimensions = append([]string{}, chunk.HardFailDimensions...)
		copyChunk.SoftFailDimensions = append([]string{}, chunk.SoftFailDimensions...)
		normalized = append(normalized, copyChunk)
	}
	return normalized
}

func sanitizeTopic(topic string) string {
	topic = strings.ToLower(strings.TrimSpace(topic))
	if topic == "" {
		return "convention-run"
	}
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ".", "-", ":", "-", ",", "-")
	topic = replacer.Replace(topic)
	parts := strings.FieldsFunc(topic, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-')
	})
	if len(parts) == 0 {
		return "convention-run"
	}
	slug := strings.Join(parts, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "convention-run"
	}
	return slug
}

func parseGeneratedArtifactList(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return []string{}
	}
	parts := strings.Split(csv, ",")
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			paths = append(paths, trimmed)
		}
	}
	return paths
}

func normalizeRepoPath(root string, path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func defaultConventionBrief(topic, contractPath string, cfg checkConfig, request orchestrationRequest) string {
	lines := []string{
		"# Convention Brief",
		"",
		fmt.Sprintf("- topic: %s", topic),
		fmt.Sprintf("- contract_path: %s", contractPath),
		fmt.Sprintf("- docs_root: %s", cfg.DocsRoot),
		fmt.Sprintf("- mode: %s", cfg.Mode),
		fmt.Sprintf("- requested_scope: %s", request.RequestedScope),
	}
	if strings.TrimSpace(request.RequestedChunkID) != "" {
		lines = append(lines, fmt.Sprintf("- requested_chunk_id: %s", request.RequestedChunkID))
	}
	if len(request.GeneratedArtifactPaths) > 0 {
		lines = append(lines, fmt.Sprintf("- generated_artifact_count: %d", len(request.GeneratedArtifactPaths)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func formatCheckerSummary(report jsonReport) string {
	lines := []string{
		fmt.Sprintf("failed=%d", report.Failed),
	}
	for _, result := range report.Results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		line := fmt.Sprintf("[%s] %s", status, result.Name)
		if strings.TrimSpace(result.Detail) != "" {
			line += ": " + result.Detail
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}

func checkerExitCode(failed int) int {
	if failed > 0 {
		return 1
	}
	return 0
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func loadJSONArtifact(root, path string, dest any) error {
	full := resolveArtifactPath(root, path)
	data, err := os.ReadFile(full)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func resolveArtifactPath(root, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(root, filepath.FromSlash(relPath))
}

func writeJSONArtifact(root, relPath string, value any) error {
	full := resolveArtifactPath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func writeTextArtifact(root, relPath, content string) error {
	full := resolveArtifactPath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}
