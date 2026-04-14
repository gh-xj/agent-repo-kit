package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunFailsWhenLaunchReceiptIsNotFresh(t *testing.T) {
	fx := newEvalFixture(t)
	fx.launchReceipt["fresh_context"] = false
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "infrastructure_failed" {
		t.Fatalf("expected infrastructure_failed, got %#v", result)
	}
}

func TestRunFailsWhenHandoffManifestIsUnreadable(t *testing.T) {
	repoRoot := t.TempDir()

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(repoRoot, "docs/reviews/missing_handoff.json", &out, &err, fixedNow)
	if exitCode != 2 {
		t.Fatalf("expected exit 2, got %d stderr=%s", exitCode, err.String())
	}
	if !strings.Contains(err.String(), "handoff") {
		t.Fatalf("expected handoff error, got %q", err.String())
	}
}

func TestRunRaisesSoftFailThresholdForHighRiskRepos(t *testing.T) {
	fx := newEvalFixture(t)
	fx.contract["evaluation_inputs"] = map[string]any{"repo_risk": "high"}
	fx.checker["failed"] = 1
	fx.checker["results"] = []map[string]any{
		{
			"name":   "file:README.md",
			"passed": false,
			"detail": "missing README marker",
		},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "semantic_failed" {
		t.Fatalf("expected semantic_failed, got %#v", result)
	}
	if len(result.SoftFailDimensions) != 1 || result.SoftFailDimensions[0] != "legibility" {
		t.Fatalf("expected legibility soft fail, got %#v", result)
	}
}

func TestRunWritesChunkScopedResultWithChunkID(t *testing.T) {
	fx := newEvalFixture(t)
	fx.handoff["requested_scope"] = "chunk"
	fx.handoff["requested_chunk_id"] = "agent-legibility"
	fx.handoff["chunk_state"] = []map[string]any{
		{"id": "repo-bootstrap", "status": "passed", "hard_fail_dimensions": []string{}, "soft_fail_dimensions": []string{}},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Scope != "chunk" || result.ChunkID != "agent-legibility" {
		t.Fatalf("expected chunk result for agent-legibility, got %#v", result)
	}
}

func TestRunChecksRegressionOnCompletedChunks(t *testing.T) {
	fx := newEvalFixture(t)
	fx.handoff["requested_scope"] = "chunk"
	fx.handoff["requested_chunk_id"] = "verification-gates"
	fx.handoff["chunk_state"] = []map[string]any{
		{"id": "repo-bootstrap", "status": "passed", "hard_fail_dimensions": []string{"verification"}, "soft_fail_dimensions": []string{}},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "semantic_failed" || len(result.HardFailDimensions) != 1 || result.HardFailDimensions[0] != "verification" {
		t.Fatalf("expected verification regression failure, got %#v", result)
	}
}

func TestRunIgnoresFuturePendingChunksInChunkScope(t *testing.T) {
	fx := newEvalFixture(t)
	fx.handoff["requested_scope"] = "chunk"
	fx.handoff["requested_chunk_id"] = "verification-gates"
	fx.handoff["chunk_state"] = []map[string]any{
		{"id": "repo-bootstrap", "status": "passed", "hard_fail_dimensions": []string{}, "soft_fail_dimensions": []string{}},
		{"id": "future-polish", "status": "pending", "hard_fail_dimensions": []string{"verification"}, "soft_fail_dimensions": []string{}},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "passed" {
		t.Fatalf("expected passed result, got %#v", result)
	}
}

func TestRunFailsFinalScopeWhenDeferredHardFailRemains(t *testing.T) {
	fx := newEvalFixture(t)
	fx.handoff["requested_scope"] = "final"
	fx.handoff["chunk_state"] = []map[string]any{
		{"id": "verification-gates", "status": "deferred", "hard_fail_dimensions": []string{"verification"}, "soft_fail_dimensions": []string{}},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "semantic_failed" || len(result.HardFailDimensions) != 1 || result.HardFailDimensions[0] != "verification" {
		t.Fatalf("expected deferred verification hard fail, got %#v", result)
	}
}

func TestRunWritesSemanticFailureResultForCheckerFailures(t *testing.T) {
	fx := newEvalFixture(t)
	fx.checker["failed"] = 1
	fx.checker["results"] = []map[string]any{
		{
			"name":   "task:Taskfile.yml:verify:",
			"passed": false,
			"detail": "missing verify task",
		},
	}
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Scope != "final" || result.Status != "semantic_failed" {
		t.Fatalf("expected final semantic failure, got %#v", result)
	}
	if len(result.HardFailDimensions) != 1 || result.HardFailDimensions[0] != "verification" {
		t.Fatalf("expected verification hard fail, got %#v", result)
	}
	if result.ReportPath != fx.reportRel || result.CheckerJSONPath != fx.checkerRel || result.LaunchReceiptPath != fx.receiptRel {
		t.Fatalf("expected artifact paths to match handoff, got %#v", result)
	}
	if result.GeneratedAt != fx.now().UTC().Format(time.RFC3339) {
		t.Fatalf("expected generated_at to use fixed clock, got %#v", result)
	}
}

func TestRunWritesPassedResultForCleanCheckerReport(t *testing.T) {
	fx := newEvalFixture(t)
	fx.writeAll(t)

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := run(fx.repoRoot, fx.handoffRel, &out, &err, fx.now)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}

	result := fx.readResult(t)
	if result.Status != "passed" {
		t.Fatalf("expected passed result, got %#v", result)
	}

	report := fx.readReport(t)
	for _, needle := range []string{"legibility", "enforceability", "verification", "drift_resistance", "ownership_clarity", "Status: passed"} {
		if !strings.Contains(report, needle) {
			t.Fatalf("expected report to contain %q, got:\n%s", needle, report)
		}
	}
}

type evalFixture struct {
	repoRoot      string
	handoffRel    string
	receiptRel    string
	checkerRel    string
	reportRel     string
	resultRel     string
	contract      map[string]any
	handoff       map[string]any
	launchReceipt map[string]any
	checker       map[string]any
	now           func() time.Time
}

func newEvalFixture(t *testing.T) evalFixture {
	t.Helper()
	repoRoot := t.TempDir()
	now := fixedNow

	fx := evalFixture{
		repoRoot:   repoRoot,
		handoffRel: "docs/reviews/handoff.json",
		receiptRel: "docs/reviews/launch-receipt.json",
		checkerRel: "docs/reviews/checker.json",
		reportRel:  "docs/reviews/report.md",
		resultRel:  "docs/reviews/evaluation_result.json",
		now:        now,
	}
	fx.contract = map[string]any{
		"contract_version": 1,
		"evaluation_inputs": map[string]any{
			"repo_risk": "standard",
		},
	}
	fx.handoff = map[string]any{
		"manifest_id":              "manifest-1",
		"contract_path":            "docs/reviews/contract.json",
		"brief_path":               "docs/planning/brief.md",
		"generated_artifact_paths": []string{"README.md"},
		"raw_evidence_bundle_path": "docs/reviews/evidence",
		"launch_receipt_path":      fx.receiptRel,
		"requested_scope":          "final",
		"checker_json_path":        fx.checkerRel,
		"report_path":              fx.reportRel,
		"result_path":              fx.resultRel,
		"chunk_state":              []map[string]any{},
	}
	fx.launchReceipt = map[string]any{
		"parent_invocation_id":    "parent-1",
		"evaluator_invocation_id": "eval-1",
		"launch_mode":             "process",
		"fresh_context":           true,
		"fork_context":            false,
		"handoff_manifest_id":     "manifest-1",
		"launched_at":             now().UTC().Format(time.RFC3339),
	}
	fx.checker = map[string]any{
		"repo_root":   repoRoot,
		"config_path": filepath.Join(repoRoot, ".convention-engineering.json"),
		"failed":      0,
		"results": []map[string]any{
			{"name": "file:README.md", "passed": true},
		},
	}

	return fx
}

func (fx evalFixture) writeAll(t *testing.T) {
	t.Helper()
	writeJSON(t, fx.repoRoot, "docs/reviews/contract.json", fx.contract)
	writeFile(t, fx.repoRoot, "docs/planning/brief.md", "# Brief\n")
	writeFile(t, fx.repoRoot, "docs/reviews/evidence/manifest.json", `{"records":[]}`)
	writeJSON(t, fx.repoRoot, fx.receiptRel, fx.launchReceipt)
	writeJSON(t, fx.repoRoot, fx.checkerRel, fx.checker)
	writeJSON(t, fx.repoRoot, fx.handoffRel, fx.handoff)
}

func (fx evalFixture) readResult(t *testing.T) evaluationResult {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fx.repoRoot, fx.resultRel))
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	result := evaluationResult{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	return result
}

func (fx evalFixture) readReport(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fx.repoRoot, fx.reportRel))
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	return string(data)
}

func writeJSON(t *testing.T, root, rel string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", rel, err)
	}
	writeFile(t, root, rel, string(data))
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
}
