package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunInitScaffoldsRepoAndTaskVerifyPasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := runInit(root, initOptions{
		Profiles:   []string{"go"},
		Operations: []string{"tickets", "wiki"},
		RepoRisk:   "standard",
	}, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected init to succeed, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	checkTaskfile(t, root, "Taskfile.yml", initConventionsTaskfile)
	checkTaskfile(t, root, initConventionsTaskfile, "check:conventions:", "verify:", "../.tickets/Taskfile.yml", "../.wiki/Taskfile.yml")

	cmd := exec.Command("task", "verify")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CONVENTION_ENGINEERING_DIR="+conventionEngineeringRoot(t))
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("expected task verify to pass: %v\n%s", runErr, output)
	}
}

func TestRunInitPreservesExistingAgentDocsAndAvoidsDuplicateManagedBlocks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")
	writeFile(t, root, "AGENTS.md", "# AGENTS.md\n\nExisting guidance.\n")
	writeFile(t, root, "CLAUDE.md", "# CLAUDE.md\n\nExisting guidance.\n")
	writeFile(t, root, "Taskfile.yml", "version: \"3\"\n\ntasks:\n  local:\n    cmds:\n      - echo local\n")

	for i := 0; i < 2; i++ {
		var out bytes.Buffer
		var err bytes.Buffer
		exitCode := runInit(root, initOptions{
			Profiles:   []string{"go"},
			Operations: []string{"tickets", "wiki"},
			RepoRisk:   "standard",
		}, &out, &err)
		if exitCode != 0 {
			t.Fatalf("expected init run %d to succeed, got %d stderr=%s stdout=%s", i+1, exitCode, err.String(), out.String())
		}
	}

	agents := readFileForTest(t, root, "AGENTS.md")
	claude := readFileForTest(t, root, "CLAUDE.md")
	taskfile := readFileForTest(t, root, "Taskfile.yml")

	if !strings.Contains(agents, "Existing guidance.") || !strings.Contains(claude, "Existing guidance.") {
		t.Fatalf("expected existing agent guidance to be preserved")
	}
	if strings.Count(agents, initManagedBlockStart) != 1 || strings.Count(claude, initManagedBlockStart) != 1 {
		t.Fatalf("expected exactly one managed block in agent docs")
	}
	if strings.Count(taskfile, initConventionsTaskfile) != 1 {
		t.Fatalf("expected root Taskfile include to be inserted once, got:\n%s", taskfile)
	}
}

func TestRunInitAutoDetectsProfiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")
	writeFile(t, root, "package.json", "{\n  \"name\": \"init-test\"\n}\n")

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := runInit(root, initOptions{}, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected init to succeed with auto-detected profiles, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	cfg, _, loadErr := loadConfig(root, initConfigPath)
	if loadErr != nil {
		t.Fatalf("expected generated config to load: %v", loadErr)
	}
	if !containsString(cfg.Profiles, "go") || !containsString(cfg.Profiles, "typescript-react") {
		t.Fatalf("expected detected profiles to include go and typescript-react, got %#v", cfg.Profiles)
	}
}

func conventionEngineeringRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func readFileForTest(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func checkTaskfile(t *testing.T, root, rel string, markers ...string) {
	t.Helper()
	content := readFileForTest(t, root, rel)
	for _, marker := range markers {
		if strings.Contains(content, marker) {
			continue
		}
		t.Fatalf("expected %s to contain %q, got:\n%s", rel, marker, content)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
