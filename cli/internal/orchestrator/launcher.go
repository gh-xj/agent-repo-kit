package orchestrator

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// EvaluatorLauncher launches the convention-evaluator against a handoff
// manifest and returns a LaunchReceipt describing the launch.
type EvaluatorLauncher interface {
	Launch(repoRoot string, handoffPath string) (LaunchReceipt, error)
}

// ProcessEvaluatorLauncher shells out to `go run` against the resolved
// evaluator script.
type ProcessEvaluatorLauncher struct {
	parentInvocationID string
	now                func() time.Time
	evaluatorScript    string
}

// NewProcessEvaluatorLauncher builds a ProcessEvaluatorLauncher. The
// repoRoot argument is accepted for API symmetry and future use (the
// launcher currently relies on the repoRoot passed to Launch).
func NewProcessEvaluatorLauncher(repoRoot, parentInvocationID, evaluatorScript string, now func() time.Time) ProcessEvaluatorLauncher {
	_ = repoRoot
	return ProcessEvaluatorLauncher{
		parentInvocationID: strings.TrimSpace(parentInvocationID),
		now:                now,
		evaluatorScript:    evaluatorScript,
	}
}

// Launch writes the launch receipt next to the handoff manifest and then
// runs the evaluator as a child process. Non-zero evaluator exits do not
// produce a Launch error — the evaluator result JSON is the source of
// truth for semantic outcomes.
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

	cmd := exec.Command("go", "run", l.evaluatorScript, "--repo-root", repoRoot, "--handoff", handoffPath)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return LaunchReceipt{}, err
		}
	}

	return receipt, nil
}

// ResolveEvaluatorScript determines the absolute path to the
// convention-evaluator main.go in a harness-agnostic way. Resolution
// order:
//  1. Explicit override (e.g. --evaluator-script / --evaluator-path).
//  2. EVALUATOR_SCRIPT_PATH environment variable (portable name).
//  3. CONVENTION_EVALUATOR_PATH environment variable (legacy alias).
//  4. Sibling lookup relative to this package's source tree, walking up
//     three directories (cli/internal/orchestrator → repo root) and then
//     into convention-evaluator/scripts/main.go. A binary-sibling probe is
//     also attempted so callers running from an installed `ark` binary
//     next to a bundled convention-evaluator directory are handled.
//  5. Repo-local skill paths under <repoRoot>/.claude/skills/,
//     <repoRoot>/.agents/skills/, or legacy <repoRoot>/.codex/skills/.
//
// Returns an absolute path to main.go that exists on disk, or an error
// describing every attempted location.
//
// Future note: if convention-evaluator is itself migrated into cli/ and
// dropped as a separate `go run` target, this resolver should collapse
// into a no-op (the in-process evaluator call would replace it). For now
// the process-launch model is preserved.
func ResolveEvaluatorScript(repoRoot, explicit string) (string, error) {
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

	// Sibling lookup when running from an installed binary. Expected layout:
	//   <install-root>/bin/ark
	//   <install-root>/convention-evaluator/scripts/main.go
	if exe, err := os.Executable(); err == nil {
		siblingFromExe := filepath.Join(filepath.Dir(exe), "..", "convention-evaluator", "scripts", "main.go")
		if path, ok := normalize(siblingFromExe); ok {
			return path, nil
		}
	}

	// Sibling lookup from this source file. The standalone kit layout is:
	//   agent-repo-kit/
	//     cli/internal/orchestrator/<this file>
	//     convention-evaluator/scripts/main.go
	// Walk up three levels (orchestrator → internal → cli) to the repo root.
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		siblingFromSource := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "convention-evaluator", "scripts", "main.go")
		if path, ok := normalize(siblingFromSource); ok {
			return path, nil
		}
	}

	// Repo-local skill placement fallbacks.
	if strings.TrimSpace(repoRoot) != "" {
		repoLocalCandidates := []string{
			filepath.Join(repoRoot, ".claude", "skills", "convention-evaluator", "scripts", "main.go"),
			filepath.Join(repoRoot, ".agents", "skills", "convention-evaluator", "scripts", "main.go"),
			filepath.Join(repoRoot, ".codex", "skills", "convention-evaluator", "scripts", "main.go"),
		}
		for _, candidate := range repoLocalCandidates {
			if path, ok := normalize(candidate); ok {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf(
		"could not locate convention-evaluator; set --evaluator-script (or --evaluator-path), "+
			"EVALUATOR_SCRIPT_PATH, or CONVENTION_EVALUATOR_PATH (attempted: %s)",
		strings.Join(attempted, ", "),
	)
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
