package contract

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunJSONModeRequiresTrackedContractWhenNoConfigProvided(t *testing.T) {
	root := t.TempDir()

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := Run(root, "", true, &out, &err)
	if exitCode != 2 {
		t.Fatalf("expected exit 2, got %d stderr=%s", exitCode, err.String())
	}
	if !strings.Contains(err.String(), ".convention-engineering.json") {
		t.Fatalf("expected tracked contract error, got %q", err.String())
	}
}

func TestRunJSONModeStructuredOutput(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeJSONConfig(t, root, ".convention-engineering.json", baseContract("tracked", "docs"))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := Run(root, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}

	report := Report{}
	if unmarshalErr := json.Unmarshal(out.Bytes(), &report); unmarshalErr != nil {
		t.Fatalf("expected valid json report: %v", unmarshalErr)
	}
	if !strings.HasSuffix(report.ConfigPath, ".convention-engineering.json") {
		t.Fatalf("expected tracked contract config path, got %q", report.ConfigPath)
	}
}

func TestRunJSONModeUsesContractCheckerTerminology(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Taskfile.yml", "version: '3'\n")
	writeJSONConfig(t, root, ".convention-engineering.json", baseContract("tracked", "docs"))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := Run(root, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", exitCode, err.String())
	}
	if strings.Contains(out.String(), "readiness") || strings.Contains(err.String(), "readiness") {
		t.Fatalf("expected contract checker terminology, stdout=%q stderr=%q", out.String(), err.String())
	}
}

func TestSelfHostedRepoRootDefaultInvocationPasses(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, repoRoot, "Taskfile.yml", "version: '3'\n")
	writeJSONConfig(t, repoRoot, ".convention-engineering.json", baseContract("tracked", "docs"))

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := Run(repoRoot, "", true, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected self-hosted repo root invocation to pass, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	report := Report{}
	if unmarshalErr := json.Unmarshal(out.Bytes(), &report); unmarshalErr != nil {
		t.Fatalf("expected valid json report: %v", unmarshalErr)
	}
	if !strings.HasSuffix(report.ConfigPath, ".convention-engineering.json") {
		t.Fatalf("expected tracked contract config path, got %q", report.ConfigPath)
	}
	if report.Failed != 0 {
		t.Fatalf("expected no self-hosted contract failures, got %#v", report)
	}
}

// NOTE: The original main_test.go contained TestSelfHostedPackageDirectoryCLIInvocationPasses
// which shelled out via `go run ./convention-engineering/scripts` to validate the
// legacy main-package binary. The ported package is a library (package contract)
// with no main entry point, and the equivalent binary-level smoke is the
// responsibility of the parent `ark check` cobra wrapper's e2e suite. Dropped
// from this unit-test surface.
