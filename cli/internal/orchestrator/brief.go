package orchestrator

import (
	"fmt"
	"strings"

	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

// defaultConventionBrief renders a minimal brief body when the caller
// does not supply one, capturing just enough metadata for the evaluator
// to orient itself.
func defaultConventionBrief(topic, contractPath string, cfg contract.Config, request Request) string {
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

// formatCheckerSummary renders a plain-text summary of a contract report
// for storage alongside the structured checker JSON.
func formatCheckerSummary(report contract.Report) string {
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

// checkerExitCode mirrors the semantic exit-code convention used by the
// standalone contract checker binary.
func checkerExitCode(failed int) int {
	if failed > 0 {
		return 1
	}
	return 0
}
