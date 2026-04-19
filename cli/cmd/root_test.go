package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

// runArk drives Execute with the given args and returns the resolved
// exit code plus the captured stdout+stderr. It's the kong-era
// replacement for the old cobra-based helper; taskfile_lint_test.go
// and any future test file reuse it.
func runArk(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := execWriters(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestExecuteResolvesExitCodeError(t *testing.T) {
	// Guard: a wrapped *ExitCodeError (code 1) must still resolve to 1
	// via errors.As, matching the pre-migration resolveCode behavior.
	wrapped := errors.Join(errors.New("ctx"), appctx.NewExitError(appctx.ExitError, ""))
	if got := appctx.ResolveExitCode(wrapped); got != appctx.ExitError {
		t.Fatalf("expected %d, got %d", appctx.ExitError, got)
	}
}

func TestExecuteUnknownCommandReturnsUsageCode(t *testing.T) {
	// Kong emits its own usage error on unknown commands; Execute must
	// translate that into ExitUsage (2).
	code, _, _ := runArk(t, "not-a-real-command")
	if code != appctx.ExitUsage {
		t.Fatalf("expected ExitUsage on unknown command, got %d", code)
	}
}

func TestExecuteUnknownFlagReturnsUsageCode(t *testing.T) {
	code, _, _ := runArk(t, "--definitely-not-a-flag")
	if code != appctx.ExitUsage {
		t.Fatalf("expected ExitUsage on unknown flag, got %d", code)
	}
}

func TestExecuteVersionCommand(t *testing.T) {
	code, stdout, _ := runArk(t, "version")
	if code != appctx.ExitSuccess {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !bytes.Contains([]byte(stdout), []byte(binaryName)) {
		t.Fatalf("expected %q in version stdout, got %q", binaryName, stdout)
	}
}
