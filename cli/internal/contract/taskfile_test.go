package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAggregateTaskfileTextRejectsEscapeInclude verifies that a Taskfile
// with an include pointing outside the repo root is rejected rather than
// silently reading arbitrary files. Prevents a path-traversal read via
// `taskfile: /etc/passwd` style entries or `../../..` escapes.
func TestAggregateTaskfileTextRejectsEscapeInclude(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.yml")
	if err := os.WriteFile(outside, []byte("version: '3'\ntasks:\n  secret:\n    cmds: [echo leaked]\n"), 0o644); err != nil {
		t.Fatalf("seed outside file: %v", err)
	}
	main := filepath.Join(root, "Taskfile.yml")
	content := "version: '3'\nincludes:\n  escape:\n    taskfile: " + outside + "\n"
	if err := os.WriteFile(main, []byte(content), 0o644); err != nil {
		t.Fatalf("write main taskfile: %v", err)
	}

	text, _, err := aggregateTaskfileText(root, "Taskfile.yml")
	if err == nil {
		t.Fatalf("expected an error for escape include, got aggregated text:\n%s", text)
	}
	if !strings.Contains(err.Error(), "escapes repo root") {
		t.Fatalf("expected escape-root error, got: %v", err)
	}
	if strings.Contains(text, "leaked") {
		t.Fatalf("secret outside-root content must not appear in aggregate, got:\n%s", text)
	}
}

// TestAggregateTaskfileTextAllowsInRootRelativeInclude confirms the fix
// doesn't regress legitimate in-tree includes.
func TestAggregateTaskfileTextAllowsInRootRelativeInclude(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	main := "version: '3'\nincludes:\n  child:\n    taskfile: sub/Taskfile.yml\n"
	child := "version: '3'\ntasks:\n  child-task:\n    cmds: [echo child]\n"
	if err := os.WriteFile(filepath.Join(root, "Taskfile.yml"), []byte(main), 0o644); err != nil {
		t.Fatalf("write main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "Taskfile.yml"), []byte(child), 0o644); err != nil {
		t.Fatalf("write child: %v", err)
	}

	text, order, err := aggregateTaskfileText(root, "Taskfile.yml")
	if err != nil {
		t.Fatalf("unexpected error for in-tree include: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 visited files, got %d: %v", len(order), order)
	}
	if !strings.Contains(text, "child-task") {
		t.Fatalf("expected child contents in aggregate, got:\n%s", text)
	}
}
