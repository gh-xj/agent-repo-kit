package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestOrchestratorWritesBriefHandoffAndLaunchReceipt(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "passed"},
		},
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:                  "Convention Launch",
		RequestedScope:         orchestrationScopeFinal,
		GeneratedArtifactPaths: []string{"README.md", "OWNERSHIP.md"},
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected orchestration to succeed, got %v", err)
	}
	if outcome.Result.Status != "passed" {
		t.Fatalf("expected passed result, got %#v", outcome.Result)
	}

	for _, rel := range []string{
		outcome.BriefPath,
		outcome.HandoffPath,
		outcome.LaunchReceiptPath,
		outcome.EvidenceManifest,
	} {
		if _, err := os.Stat(filepath.Join(fx.repoRoot, rel)); err != nil {
			t.Fatalf("expected artifact %s to exist: %v", rel, err)
		}
	}
}

func TestOrchestratorRetriesInfrastructureFailureOnce(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "infrastructure_failed"},
			{resultStatus: "passed"},
		},
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:          "Convention Retry",
		RequestedScope: orchestrationScopeFinal,
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected orchestration retry to succeed, got %v", err)
	}
	if len(launcher.calls) != 2 {
		t.Fatalf("expected one retry, got %d calls", len(launcher.calls))
	}
	if outcome.Result.Status != "passed" {
		t.Fatalf("expected final passed status, got %#v", outcome.Result)
	}
}

func TestOrchestratorPersistsChunkStateForChunkScope(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "passed"},
		},
	}
	request := orchestrationRequest{
		Topic:             "Chunk Scope",
		RequestedScope:    orchestrationScopeChunk,
		RequestedChunkID:  "agent-legibility",
		GeneratedArtifactPaths: []string{"README.md"},
		ChunkState: []orchestratorChunkState{
			{ID: "repo-bootstrap", Status: "passed"},
			{ID: "verification-gates", Status: "pending"},
			{ID: "ownership-cleanup", Status: "deferred", HardFailDimensions: []string{"ownership_clarity"}},
		},
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", request, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected chunk orchestration to succeed, got %v", err)
	}

	manifest := orchestrationHandoffManifest{}
	readJSONForTest(t, fx.repoRoot, outcome.HandoffPath, &manifest)
	if manifest.RequestedScope != orchestrationScopeChunk || manifest.RequestedChunkID != "agent-legibility" {
		t.Fatalf("expected chunk scope handoff, got %#v", manifest)
	}
	if len(manifest.ChunkState) != 3 || manifest.ChunkState[0].Status != "passed" || manifest.ChunkState[2].Status != "deferred" {
		t.Fatalf("expected persisted chunk states, got %#v", manifest.ChunkState)
	}
}

func TestOrchestratorDoesNotPassSessionContextToEvaluatorLaunch(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "passed"},
		},
	}

	_, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:          "Launch Boundary",
		RequestedScope: orchestrationScopeFinal,
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected orchestration to succeed, got %v", err)
	}
	if len(launcher.calls) != 1 {
		t.Fatalf("expected one launch call, got %d", len(launcher.calls))
	}
	if launcher.calls[0].repoRoot != fx.repoRoot {
		t.Fatalf("expected repo root launch arg, got %#v", launcher.calls[0])
	}
	if !strings.HasSuffix(launcher.calls[0].handoffPath, "_handoff.json") {
		t.Fatalf("expected handoff path only, got %#v", launcher.calls[0])
	}
}

func TestOrchestratorWritesFullLaunchReceiptFields(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "passed"},
		},
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:               "Receipt Shape",
		ParentInvocationID:  "parent-invocation",
		RequestedScope:      orchestrationScopeFinal,
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected orchestration to succeed, got %v", err)
	}

	receipt := orchestrationLaunchReceipt{}
	readJSONForTest(t, fx.repoRoot, outcome.LaunchReceiptPath, &receipt)
	if receipt.ParentInvocationID == "" || receipt.EvaluatorInvocationID == "" || receipt.LaunchMode == "" || receipt.HandoffManifestID == "" || receipt.LaunchedAt == "" {
		t.Fatalf("expected full launch receipt, got %#v", receipt)
	}
	if !receipt.FreshContext {
		t.Fatalf("expected fresh_context=true, got %#v", receipt)
	}
	if receipt.ForkContext == nil || *receipt.ForkContext {
		t.Fatalf("expected fork_context=false, got %#v", receipt)
	}
}

func TestProcessEvaluatorLauncherMakesReceiptAvailableToEvaluator(t *testing.T) {
	fx := newOrchestratorFixture(t)
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	evaluatorScript := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "convention-evaluator", "scripts", "main.go"))
	launcher := processEvaluatorLauncher{
		parentInvocationID: "parent-process",
		now:                orchestratorFixedNow,
		evaluatorScript:    evaluatorScript,
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:          "Process Launch",
		RequestedScope: orchestrationScopeFinal,
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected process launch orchestration to succeed, got %v", err)
	}
	if outcome.Result.Status != "passed" {
		t.Fatalf("expected passed process launch result, got %#v", outcome.Result)
	}

	receipt := orchestrationLaunchReceipt{}
	readJSONForTest(t, fx.repoRoot, outcome.LaunchReceiptPath, &receipt)
	if !receipt.FreshContext || receipt.LaunchMode != "process" {
		t.Fatalf("expected process-mode fresh receipt, got %#v", receipt)
	}
}

func TestOrchestratorWritesFullHandoffManifestFields(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "passed"},
		},
	}

	outcome, err := orchestrateEvaluation(fx.repoRoot, "", orchestrationRequest{
		Topic:                  "Handoff Shape",
		RequestedScope:         orchestrationScopeFinal,
		GeneratedArtifactPaths: []string{"README.md", "OWNERSHIP.md"},
	}, launcher, orchestratorFixedNow)
	if err != nil {
		t.Fatalf("expected orchestration to succeed, got %v", err)
	}

	manifest := orchestrationHandoffManifest{}
	readJSONForTest(t, fx.repoRoot, outcome.HandoffPath, &manifest)
	if manifest.ManifestID == "" || manifest.ContractPath == "" || manifest.BriefPath == "" || manifest.RawEvidenceBundlePath == "" || manifest.LaunchReceiptPath == "" || manifest.RequestedScope == "" {
		t.Fatalf("expected full handoff manifest, got %#v", manifest)
	}
	if len(manifest.GeneratedArtifactPaths) != 2 {
		t.Fatalf("expected generated artifact paths to persist, got %#v", manifest)
	}
	if manifest.CheckerJSONPath == "" || manifest.ReportPath == "" || manifest.ResultPath == "" {
		t.Fatalf("expected evaluator artifact paths in handoff, got %#v", manifest)
	}
}

type orchestratorFixture struct {
	repoRoot string
}

func newOrchestratorFixture(t *testing.T) orchestratorFixture {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")

	cfg := baseContract("tracked", "docs")
	cfg["required_files"] = []string{"Taskfile.yml"}
	cfg["taskfile_checks"] = map[string][]string{}
	cfg["canonical_pointers"] = []map[string]any{}
	cfg["content_checks"] = []map[string]any{}
	cfg["git_exclude_checks"] = []map[string]any{}
	cfg["chunk_plan"] = map[string]any{
		"enabled": false,
		"chunks":  []map[string]any{},
	}
	writeJSONConfig(t, root, ".convention-engineering.json", cfg)

	return orchestratorFixture{repoRoot: root}
}

type fakeLaunchAttempt struct {
	resultStatus string
	launchErr    error
}

type fakeLaunchCall struct {
	repoRoot   string
	handoffPath string
}

type fakeEvaluatorLauncher struct {
	attempts []fakeLaunchAttempt
	calls    []fakeLaunchCall
}

func (f *fakeEvaluatorLauncher) Launch(repoRoot string, handoffPath string) (orchestrationLaunchReceipt, error) {
	f.calls = append(f.calls, fakeLaunchCall{repoRoot: repoRoot, handoffPath: handoffPath})
	attempt := fakeLaunchAttempt{resultStatus: "passed"}
	if len(f.attempts) >= len(f.calls) {
		attempt = f.attempts[len(f.calls)-1]
	}

	manifest := orchestrationHandoffManifest{}
	if err := loadJSONArtifact(repoRoot, handoffPath, &manifest); err != nil {
		return orchestrationLaunchReceipt{}, err
	}

	forkContext := false
	receipt := orchestrationLaunchReceipt{
		ParentInvocationID:    "parent-1",
		EvaluatorInvocationID: fmt.Sprintf("eval-%d", len(f.calls)),
		LaunchMode:            "agent",
		FreshContext:          true,
		ForkContext:           &forkContext,
		HandoffManifestID:     manifest.ManifestID,
		LaunchedAt:            orchestratorFixedNow().UTC().Format(time.RFC3339),
	}
	if err := writeJSONForTest(repoRoot, manifest.LaunchReceiptPath, receipt); err != nil {
		return orchestrationLaunchReceipt{}, err
	}

	result := orchestrationEvaluationResult{
		Scope:             manifest.RequestedScope,
		Status:            attempt.resultStatus,
		ChunkID:           manifest.RequestedChunkID,
		ReportPath:        manifest.ReportPath,
		CheckerJSONPath:   manifest.CheckerJSONPath,
		LaunchReceiptPath: manifest.LaunchReceiptPath,
		GeneratedAt:       orchestratorFixedNow().UTC().Format(time.RFC3339),
	}
	if err := writeJSONForTest(repoRoot, manifest.ResultPath, result); err != nil {
		return orchestrationLaunchReceipt{}, err
	}
	if err := writeTextForTest(repoRoot, manifest.ReportPath, "# Convention Evaluation\n"); err != nil {
		return orchestrationLaunchReceipt{}, err
	}

	if attempt.launchErr != nil {
		return receipt, attempt.launchErr
	}
	return receipt, nil
}

func readJSONForTest(t *testing.T, root, rel string, dest any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		t.Fatalf("unmarshal %s: %v", rel, err)
	}
}

func writeJSONForTest(root, rel string, value any) error {
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func writeTextForTest(root, rel, content string) error {
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func orchestratorFixedNow() time.Time {
	return time.Date(2026, time.April, 3, 16, 0, 0, 0, time.UTC)
}

func TestRunOrchestrationReturnsSemanticFailureExitCode(t *testing.T) {
	fx := newOrchestratorFixture(t)
	launcher := &fakeEvaluatorLauncher{
		attempts: []fakeLaunchAttempt{
			{resultStatus: "semantic_failed"},
		},
	}

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := runOrchestration(fx.repoRoot, "", orchestrationRequest{
		Topic:          "CLI Exit",
		RequestedScope: orchestrationScopeFinal,
	}, launcher, &out, &err, orchestratorFixedNow)
	if exitCode != 1 {
		t.Fatalf("expected semantic failure exit 1, got %d stderr=%s", exitCode, err.String())
	}
	if !strings.Contains(out.String(), "status=semantic_failed") {
		t.Fatalf("expected semantic status output, got %q", out.String())
	}
}
