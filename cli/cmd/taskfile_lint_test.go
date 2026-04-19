package cmd

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

func TestTaskfileLintJSONExitCodeOnFindings(t *testing.T) {
	// Bug 1 regression: `--json` must still exit 1 when findings exist.
	dir := t.TempDir()
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	writeFile(t, taskfilePath, "version: \"3\"\ntasks:\n  foo:\n    cmd: a\n    cmds: [b]\n")

	code, stdout, _ := runArk(t, "--json", "taskfile", "lint", "--repo-root", dir, "--path", taskfilePath)
	if code != appctx.ExitError {
		t.Fatalf("expected exit %d on findings with --json, got %d (stdout=%q)", appctx.ExitError, code, stdout)
	}

	// stdout must still contain the JSON report with the finding.
	var report struct {
		Count    int `json:"count"`
		Findings []struct {
			RuleID string `json:"rule_id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("unmarshal JSON report: %v (stdout=%q)", err, stdout)
	}
	if report.Count == 0 {
		t.Fatalf("expected at least one finding in JSON, got %d (stdout=%q)", report.Count, stdout)
	}
	var seen bool
	for _, f := range report.Findings {
		if f.RuleID == "cmd-and-cmds-mutex" {
			seen = true
		}
	}
	if !seen {
		t.Errorf("expected cmd-and-cmds-mutex rule in report, got %+v", report.Findings)
	}
}

func TestTaskfileLintJSONExitCodeOnClean(t *testing.T) {
	dir := t.TempDir()
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	writeFile(t, taskfilePath, "version: \"3\"\ntasks:\n  foo:\n    cmds: [echo hi]\n")

	code, _, _ := runArk(t, "--json", "taskfile", "lint", "--repo-root", dir, "--path", taskfilePath)
	if code != appctx.ExitSuccess {
		t.Fatalf("expected exit 0 on clean file with --json, got %d", code)
	}
}

func TestTaskfileLintExitCodeOnMissingFile(t *testing.T) {
	// I/O errors (missing file) must exit 2 (ExitUsage), not 1.
	dir := t.TempDir()
	code, _, stderr := runArk(t, "taskfile", "lint", "--repo-root", dir, "--path", filepath.Join(dir, "_nope.yml"))
	if code != appctx.ExitUsage {
		t.Fatalf("expected exit %d on missing file, got %d (stderr=%q)", appctx.ExitUsage, code, stderr)
	}
}

func TestTaskfileLintExitCodeOnParseError(t *testing.T) {
	// YAML parse error → exit 2 (ExitUsage). The user fixes the input.
	dir := t.TempDir()
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	writeFile(t, taskfilePath, "version: '3'\ntasks:\n  build:\n    cmds:\n  - : : not yaml\n")

	code, _, _ := runArk(t, "taskfile", "lint", "--repo-root", dir, "--path", taskfilePath)
	if code != appctx.ExitUsage {
		t.Fatalf("expected exit %d on parse error, got %d", appctx.ExitUsage, code)
	}
}

func TestTaskfileLintExitCodeOnSchemaError(t *testing.T) {
	// Schema-error (dotenv as scalar) must be exit 1 — it's a lint finding,
	// not a YAML parse error. The rule's emitted in addition to other rules.
	dir := t.TempDir()
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	writeFile(t, taskfilePath, "version: \"3\"\ndotenv: \".env\"\ntasks:\n  build:\n    cmds: [echo hi]\n")

	code, stdout, _ := runArk(t, "taskfile", "lint", "--repo-root", dir, "--path", taskfilePath)
	if code != appctx.ExitError {
		t.Fatalf("expected exit %d on schema-error finding, got %d", appctx.ExitError, code)
	}
	if !strings.Contains(stdout, "schema-error") {
		t.Errorf("expected stdout to contain schema-error rule id, got %q", stdout)
	}
}
