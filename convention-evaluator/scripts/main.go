package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type handoffManifest struct {
	ManifestID             string       `json:"manifest_id"`
	ContractPath           string       `json:"contract_path"`
	BriefPath              string       `json:"brief_path"`
	GeneratedArtifactPaths []string     `json:"generated_artifact_paths"`
	RawEvidenceBundlePath  string       `json:"raw_evidence_bundle_path"`
	LaunchReceiptPath      string       `json:"launch_receipt_path"`
	RequestedScope         string       `json:"requested_scope"`
	RequestedChunkID       string       `json:"requested_chunk_id"`
	CheckerJSONPath        string       `json:"checker_json_path"`
	ReportPath             string       `json:"report_path"`
	ResultPath             string       `json:"result_path"`
	ChunkState             []chunkState `json:"chunk_state"`
}

type chunkState struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	HardFailDimensions []string `json:"hard_fail_dimensions"`
	SoftFailDimensions []string `json:"soft_fail_dimensions"`
}

type launchReceipt struct {
	ParentInvocationID    string `json:"parent_invocation_id"`
	EvaluatorInvocationID string `json:"evaluator_invocation_id"`
	LaunchMode            string `json:"launch_mode"`
	FreshContext          bool   `json:"fresh_context"`
	ForkContext           *bool  `json:"fork_context,omitempty"`
	HandoffManifestID     string `json:"handoff_manifest_id"`
	LaunchedAt            string `json:"launched_at"`
}

type checkerReport struct {
	RepoRoot   string          `json:"repo_root"`
	ConfigPath string          `json:"config_path"`
	Failed     int             `json:"failed"`
	Results    []checkerResult `json:"results"`
}

type checkerResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

type contractFile struct {
	EvaluationInputs evaluationInputs `json:"evaluation_inputs"`
}

type evaluationInputs struct {
	RepoRisk string `json:"repo_risk"`
}

type evaluationResult struct {
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

var dimensionOrder = []string{
	"legibility",
	"enforceability",
	"verification",
	"drift_resistance",
	"ownership_clarity",
}

var hardFailDimensions = map[string]bool{
	"enforceability":    true,
	"verification":      true,
	"ownership_clarity": true,
}

func main() {
	repoRoot := flag.String("repo-root", ".", "Path to repository root")
	handoff := flag.String("handoff", "", "Path to handoff manifest JSON")
	flag.Parse()

	os.Exit(run(*repoRoot, *handoff, os.Stdout, os.Stderr, time.Now))
}

func run(repoRoot, handoffPath string, stdout, stderr io.Writer, now func() time.Time) int {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve repo root: %v\n", err)
		return 2
	}
	if strings.TrimSpace(handoffPath) == "" {
		fmt.Fprintln(stderr, "failed to read handoff: --handoff is required")
		return 2
	}

	manifest, handoffRel, err := loadHandoff(root, handoffPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read handoff: %v\n", err)
		return 2
	}

	if err := validateHandoff(manifest); err != nil {
		fmt.Fprintf(stderr, "failed to validate handoff: %v\n", err)
		return 2
	}

	reportRel := valueOrDefault(manifest.ReportPath, siblingRelPath(handoffRel, "report.md"))
	resultRel := valueOrDefault(manifest.ResultPath, siblingRelPath(handoffRel, "evaluation_result.json"))
	checkerRel := valueOrDefault(manifest.CheckerJSONPath, siblingRelPath(handoffRel, "checker.json"))

	result := evaluationResult{
		Scope:             manifest.RequestedScope,
		ChunkID:           manifest.RequestedChunkID,
		ReportPath:        reportRel,
		CheckerJSONPath:   checkerRel,
		LaunchReceiptPath: manifest.LaunchReceiptPath,
		GeneratedAt:       now().UTC().Format(time.RFC3339),
	}

	receipt, infraErr := loadLaunchReceipt(root, manifest)
	if infraErr == nil {
		infraErr = validateLaunchReceipt(receipt, manifest)
	}

	briefErr := requirePathExists(root, manifest.BriefPath)
	evidenceErr := requirePathExists(root, manifest.RawEvidenceBundlePath)

	contract := contractFile{}
	contractErr := loadJSON(root, manifest.ContractPath, &contract)
	if contract.EvaluationInputs.RepoRisk == "" {
		contract.EvaluationInputs.RepoRisk = "standard"
	}

	checker := checkerReport{}
	checkerErr := loadJSON(root, checkerRel, &checker)

	if infraErr != nil || briefErr != nil || evidenceErr != nil || contractErr != nil || checkerErr != nil {
		reason := firstError(
			infraErr,
			briefErr,
			evidenceErr,
			contractErr,
			checkerErr,
		)
		result.Status = "infrastructure_failed"
		if err := writeReport(root, reportRel, buildInfrastructureReport(result, reason)); err != nil {
			fmt.Fprintf(stderr, "failed to write report: %v\n", err)
			return 2
		}
		if err := writeResult(root, resultRel, result); err != nil {
			fmt.Fprintf(stderr, "failed to write result: %v\n", err)
			return 2
		}
		fmt.Fprintf(stdout, "status=%s\n", result.Status)
		return 1
	}

	scores := map[string]int{}
	failuresByDimension := map[string][]string{}
	for _, dimension := range dimensionOrder {
		scores[dimension] = 4
	}

	for _, check := range checker.Results {
		if check.Passed {
			continue
		}
		dimension := mapCheckToDimension(check.Name)
		scores[dimension] = max(0, scores[dimension]-2)
		failuresByDimension[dimension] = append(failuresByDimension[dimension], formatCheckFailure(check))
	}

	if manifest.RequestedScope == "chunk" {
		for _, chunk := range manifest.ChunkState {
			if !isPassedChunkState(chunk.Status) {
				continue
			}
			applyChunkFailure(scores, failuresByDimension, chunk, "completed chunk regression")
		}
	} else {
		for _, chunk := range manifest.ChunkState {
			if chunk.Status != "deferred" {
				continue
			}
			applyChunkFailure(scores, failuresByDimension, chunk, "deferred chunk remains unresolved")
		}
	}

	softThreshold := 2
	if strings.EqualFold(contract.EvaluationInputs.RepoRisk, "high") {
		softThreshold = 3
	}

	hardFails := collectFailingDimensions(scores, 3, hardFailDimensions, true)
	softFails := collectFailingDimensions(scores, softThreshold, hardFailDimensions, false)

	result.HardFailDimensions = hardFails
	result.SoftFailDimensions = softFails
	if len(hardFails) > 0 || len(softFails) > 0 {
		result.Status = "semantic_failed"
	} else {
		result.Status = "passed"
	}

	report := buildSemanticReport(result, scores, failuresByDimension, contract.EvaluationInputs.RepoRisk, 3, softThreshold)
	if err := writeReport(root, reportRel, report); err != nil {
		fmt.Fprintf(stderr, "failed to write report: %v\n", err)
		return 2
	}
	if err := writeResult(root, resultRel, result); err != nil {
		fmt.Fprintf(stderr, "failed to write result: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "status=%s\n", result.Status)
	if result.Status == "passed" {
		return 0
	}
	return 1
}

func loadHandoff(root, handoffPath string) (handoffManifest, string, error) {
	rel := handoffPath
	if filepath.IsAbs(handoffPath) {
		var err error
		rel, err = filepath.Rel(root, handoffPath)
		if err != nil {
			rel = handoffPath
		}
	}
	manifest := handoffManifest{}
	if err := loadJSON(root, handoffPath, &manifest); err != nil {
		return handoffManifest{}, "", err
	}
	return manifest, filepath.ToSlash(rel), nil
}

func validateHandoff(manifest handoffManifest) error {
	required := map[string]string{
		"manifest_id":              manifest.ManifestID,
		"contract_path":            manifest.ContractPath,
		"brief_path":               manifest.BriefPath,
		"raw_evidence_bundle_path": manifest.RawEvidenceBundlePath,
		"launch_receipt_path":      manifest.LaunchReceiptPath,
		"requested_scope":          manifest.RequestedScope,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	switch manifest.RequestedScope {
	case "chunk":
		if strings.TrimSpace(manifest.RequestedChunkID) == "" {
			return fmt.Errorf("requested_chunk_id is required for chunk scope")
		}
	case "final":
	default:
		return fmt.Errorf("requested_scope must be chunk or final")
	}

	return nil
}

func loadLaunchReceipt(root string, manifest handoffManifest) (launchReceipt, error) {
	receipt := launchReceipt{}
	return receipt, loadJSON(root, manifest.LaunchReceiptPath, &receipt)
}

func validateLaunchReceipt(receipt launchReceipt, manifest handoffManifest) error {
	required := map[string]string{
		"parent_invocation_id":    receipt.ParentInvocationID,
		"evaluator_invocation_id": receipt.EvaluatorInvocationID,
		"launch_mode":             receipt.LaunchMode,
		"handoff_manifest_id":     receipt.HandoffManifestID,
		"launched_at":             receipt.LaunchedAt,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if receipt.HandoffManifestID != manifest.ManifestID {
		return fmt.Errorf("handoff manifest ids do not match")
	}
	if !receipt.FreshContext {
		return fmt.Errorf("fresh_context must be true")
	}
	if receipt.ForkContext != nil && *receipt.ForkContext {
		return fmt.Errorf("fork_context must be false")
	}
	switch receipt.LaunchMode {
	case "agent", "process":
	default:
		return fmt.Errorf("launch_mode must be agent or process")
	}
	if _, err := time.Parse(time.RFC3339, receipt.LaunchedAt); err != nil {
		return fmt.Errorf("launched_at must be RFC3339: %w", err)
	}
	return nil
}

func loadJSON(root, path string, dest any) error {
	full := resolvePath(root, path)
	data, err := os.ReadFile(full)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func requirePathExists(root, path string) error {
	full := resolvePath(root, path)
	_, err := os.Stat(full)
	return err
}

func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func siblingRelPath(path, sibling string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(filepath.FromSlash(path)), sibling))
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func mapCheckToDimension(name string) string {
	switch {
	case strings.HasPrefix(name, "task:"):
		return "verification"
	case strings.HasPrefix(name, "invariants:"):
		return "enforceability"
	case strings.HasPrefix(name, "file:"):
		return "legibility"
	case strings.Contains(name, "ownership"):
		return "ownership_clarity"
	case strings.Contains(name, "canonical") || strings.Contains(name, "git-exclude") || strings.Contains(name, "pointer"):
		return "drift_resistance"
	default:
		return "legibility"
	}
}

func formatCheckFailure(check checkerResult) string {
	if strings.TrimSpace(check.Detail) == "" {
		return check.Name
	}
	return fmt.Sprintf("%s: %s", check.Name, check.Detail)
}

func isPassedChunkState(status string) bool {
	switch status {
	case "passed", "completed":
		return true
	default:
		return false
	}
}

func applyChunkFailure(scores map[string]int, failuresByDimension map[string][]string, chunk chunkState, reason string) {
	for _, dimension := range chunk.HardFailDimensions {
		if strings.TrimSpace(dimension) == "" {
			continue
		}
		scores[dimension] = min(scores[dimension], 2)
		failuresByDimension[dimension] = append(failuresByDimension[dimension], fmt.Sprintf("%s (%s)", chunk.ID, reason))
	}
	for _, dimension := range chunk.SoftFailDimensions {
		if strings.TrimSpace(dimension) == "" {
			continue
		}
		scores[dimension] = min(scores[dimension], 1)
		failuresByDimension[dimension] = append(failuresByDimension[dimension], fmt.Sprintf("%s (%s)", chunk.ID, reason))
	}
}

func collectFailingDimensions(scores map[string]int, threshold int, hardMap map[string]bool, wantHard bool) []string {
	dimensions := []string{}
	for _, dimension := range dimensionOrder {
		isHard := hardMap[dimension]
		if isHard != wantHard {
			continue
		}
		if scores[dimension] < threshold {
			dimensions = append(dimensions, dimension)
		}
	}
	sort.Strings(dimensions)
	return dimensions
}

func buildInfrastructureReport(result evaluationResult, reason error) string {
	lines := []string{
		"# Convention Evaluation",
		"",
		fmt.Sprintf("Status: %s", result.Status),
		"",
		"## Infrastructure Failure",
		reason.Error(),
	}
	return strings.Join(lines, "\n")
}

func buildSemanticReport(result evaluationResult, scores map[string]int, failuresByDimension map[string][]string, repoRisk string, hardThreshold, softThreshold int) string {
	lines := []string{
		"# Convention Evaluation",
		"",
		fmt.Sprintf("Status: %s", result.Status),
		fmt.Sprintf("Scope: %s", result.Scope),
	}
	if result.ChunkID != "" {
		lines = append(lines, fmt.Sprintf("Chunk ID: %s", result.ChunkID))
	}
	lines = append(lines,
		"",
		"## Thresholds",
		fmt.Sprintf("- repo_risk: %s", valueOrDefault(repoRisk, "standard")),
		fmt.Sprintf("- hard-fail threshold: %d", hardThreshold),
		fmt.Sprintf("- soft-fail threshold: %d", softThreshold),
		"",
		"## Scores",
	)
	for _, dimension := range dimensionOrder {
		lines = append(lines, fmt.Sprintf("- %s: %d/4", dimension, scores[dimension]))
	}
	lines = append(lines, "", "## Failed Checks By Dimension")
	for _, dimension := range dimensionOrder {
		failures := failuresByDimension[dimension]
		if len(failures) == 0 {
			lines = append(lines, fmt.Sprintf("### %s", dimension), "- none")
			continue
		}
		lines = append(lines, fmt.Sprintf("### %s", dimension))
		for _, failure := range failures {
			lines = append(lines, fmt.Sprintf("- %s", failure))
		}
	}
	return strings.Join(lines, "\n")
}

func writeReport(root, relPath, content string) error {
	full := resolvePath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func writeResult(root, relPath string, result evaluationResult) error {
	full := resolvePath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
