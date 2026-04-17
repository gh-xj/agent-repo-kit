package orchestrator

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gh-xj/agent-repo-kit/cli/internal/evaluator"
)

// EvaluatorLauncher launches the convention-evaluator against a handoff
// manifest and returns a LaunchReceipt describing the launch.
type EvaluatorLauncher interface {
	Launch(repoRoot string, handoffPath string) (LaunchReceipt, error)
}

// ProcessEvaluatorLauncher writes the launch receipt and invokes the
// absorbed in-process evaluator (Stage 5 — formerly a `go run` subprocess
// against convention-evaluator/scripts/main.go). The name is preserved for
// historical reasons and because the receipt still reports
// launch_mode="process" (the enclosing ark process).
type ProcessEvaluatorLauncher struct {
	parentInvocationID string
	now                func() time.Time
}

// NewProcessEvaluatorLauncher builds a ProcessEvaluatorLauncher. The
// repoRoot argument is accepted for API symmetry with the pre-Stage-5
// signature but is no longer used (the evaluator is in-process now).
func NewProcessEvaluatorLauncher(repoRoot, parentInvocationID string, now func() time.Time) ProcessEvaluatorLauncher {
	_ = repoRoot
	return ProcessEvaluatorLauncher{
		parentInvocationID: strings.TrimSpace(parentInvocationID),
		now:                now,
	}
}

// Launch writes the launch receipt next to the handoff manifest and then
// calls the in-process evaluator. Non-zero evaluator exits do not surface
// as a Launch error — the evaluator result JSON is the source of truth for
// semantic outcomes. Only true infrastructure failures (could not write
// receipt, or could not parse handoff) return an error.
func (l ProcessEvaluatorLauncher) Launch(repoRoot string, handoffPath string) (LaunchReceipt, error) {
	manifest := HandoffManifest{}
	if err := loadJSONArtifact(repoRoot, handoffPath, &manifest); err != nil {
		return LaunchReceipt{}, err
	}

	forkContext := false
	receipt := LaunchReceipt{
		ParentInvocationID:    valueOrDefault(l.parentInvocationID, "manual"),
		EvaluatorInvocationID: fmt.Sprintf("convention-evaluator-%d", l.now().UTC().UnixNano()),
		LaunchMode:            "process",
		FreshContext:          true,
		ForkContext:           &forkContext,
		HandoffManifestID:     manifest.ManifestID,
		LaunchedAt:            l.now().UTC().Format(time.RFC3339),
	}
	if err := writeJSONArtifact(repoRoot, manifest.LaunchReceiptPath, receipt); err != nil {
		return LaunchReceipt{}, err
	}

	// In-process evaluation. stdout/stderr are captured to avoid polluting
	// the orchestrator's own stream; the evaluator writes the authoritative
	// result JSON + report markdown to paths declared in the manifest.
	var evalOut, evalErr bytes.Buffer
	_ = evaluator.Run(repoRoot, handoffPath, &evalOut, &evalErr, l.now)

	return receipt, nil
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
