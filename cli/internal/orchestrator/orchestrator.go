// Package orchestrator wires the convention contract checker and the
// convention-evaluator together, producing the handoff artifacts (brief,
// checker report, evidence manifest, handoff manifest, launch receipt,
// evaluator result) that drive a single evaluation round.
package orchestrator

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

// Scope constants for Request.RequestedScope.
const (
	ScopeFinal = "final"
	ScopeChunk = "chunk"
)

// Request is the user-facing orchestration input. Fields mirror the CLI
// flags plumbed by the `ark orchestrate` command.
type Request struct {
	Topic                  string
	ParentInvocationID     string
	RequestedScope         string
	RequestedChunkID       string
	GeneratedArtifactPaths []string
	ChunkState             []ChunkState
	BriefBody              string
}

// Outcome captures every artifact path the orchestrator produced plus
// the evaluator's final verdict.
type Outcome struct {
	BriefPath          string
	HandoffPath        string
	LaunchReceiptPath  string
	EvidenceBundlePath string
	EvidenceManifest   string
	CheckerJSONPath    string
	ReportPath         string
	ResultPath         string
	Result             EvaluationResult
}

// Run is the top-level orchestration entry point used by the cobra
// command. It reports semantic failures (exit 1), infrastructure
// failures (exit 2), and successes (exit 0).
func Run(repoRoot, configPath string, req Request, launcher EvaluatorLauncher, stdout, stderr io.Writer, now func() time.Time) int {
	outcome, err := orchestrateEvaluation(repoRoot, configPath, req, launcher, now)
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

// ParseGeneratedArtifactList splits a comma-separated list of paths,
// trimming whitespace and dropping empty entries.
func ParseGeneratedArtifactList(csv string) []string {
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

func orchestrateEvaluation(repoRoot, configPath string, request Request, launcher EvaluatorLauncher, now func() time.Time) (Outcome, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Outcome{}, fmt.Errorf("resolve repo root: %w", err)
	}
	if strings.TrimSpace(request.RequestedScope) == "" {
		request.RequestedScope = ScopeFinal
	}
	if err := validateOrchestrationRequest(request); err != nil {
		return Outcome{}, err
	}

	cfg, resolvedConfigPath, err := contract.LoadConfig(root, configPath)
	if err != nil {
		return Outcome{}, fmt.Errorf("load contract: %w", err)
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
		return Outcome{}, fmt.Errorf("write brief: %w", err)
	}

	checkerStartedAt := now().UTC()
	checkResults := contract.RunChecks(root, cfg)
	checkerFinishedAt := now().UTC()
	checkerReport := contract.Report{
		RepoRoot:   root,
		ConfigPath: resolvedConfigPath,
		Failed:     contract.CountFailures(checkResults),
		Results:    checkResults,
	}
	if err := writeJSONArtifact(root, checkerRel, checkerReport); err != nil {
		return Outcome{}, fmt.Errorf("write checker report: %w", err)
	}
	if err := writeTextArtifact(root, checkerStdoutRel, formatCheckerSummary(checkerReport)); err != nil {
		return Outcome{}, fmt.Errorf("write checker stdout: %w", err)
	}
	if err := writeTextArtifact(root, checkerStderrRel, ""); err != nil {
		return Outcome{}, fmt.Errorf("write checker stderr: %w", err)
	}

	evidence := EvidenceManifest{
		Records: []EvidenceRecord{
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
		return Outcome{}, fmt.Errorf("write evidence manifest: %w", err)
	}

	manifest := HandoffManifest{
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
		return Outcome{}, fmt.Errorf("write handoff: %w", err)
	}

	outcome := Outcome{
		BriefPath:          briefRel,
		HandoffPath:        handoffRel,
		LaunchReceiptPath:  receiptRel,
		EvidenceBundlePath: evidenceDirRel,
		EvidenceManifest:   evidenceManifestRel,
		CheckerJSONPath:    checkerRel,
		ReportPath:         reportRel,
		ResultPath:         resultRel,
	}

	for attempt := 0; attempt < 2; attempt++ {
		_, launchErr := launcher.Launch(root, resolveArtifactPath(root, handoffRel))
		if launchErr != nil {
			if attempt == 0 {
				continue
			}
			// A launch failure on the retry is a genuine infrastructure
			// problem distinct from a prior infrastructure_failed result:
			// the launcher itself would not start. Propagate the error
			// instead of silently masking it as a stale success-like result.
			return Outcome{}, fmt.Errorf("launch evaluator: %w", launchErr)
		}

		result := EvaluationResult{}
		if err := loadJSONArtifact(root, resultRel, &result); err != nil {
			if attempt == 0 {
				continue
			}
			return Outcome{}, fmt.Errorf("read evaluation result: %w", err)
		}

		if result.Status == "infrastructure_failed" && attempt == 0 {
			continue
		}

		outcome.Result = result
		return outcome, nil
	}

	return Outcome{}, fmt.Errorf("evaluation did not produce a result")
}

func validateOrchestrationRequest(request Request) error {
	scope := strings.TrimSpace(request.RequestedScope)
	if scope == "" {
		scope = ScopeFinal
	}
	switch scope {
	case ScopeFinal, ScopeChunk:
	default:
		return fmt.Errorf("requested scope must be final or chunk")
	}
	if scope == ScopeChunk && strings.TrimSpace(request.RequestedChunkID) == "" {
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

func normalizeChunkStates(chunks []ChunkState) []ChunkState {
	normalized := make([]ChunkState, 0, len(chunks))
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
