package scaffold

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gh-xj/agent-repo-kit/cli/internal/contract"
)

func TestRunInitScaffoldsRepoAndTaskVerifyPasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := RunInit(root, Options{
		Profiles:   []string{"go"},
		Operations: []string{"work"},
		RepoRisk:   "standard",
	}, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected init to succeed, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	checkTaskfile(t, root, "Taskfile.yml", ConventionsTaskfile, "work:")
	checkTaskfile(t, root, ConventionsTaskfile, "check:conventions:", "work:check:", "verify:")
	if _, statErr := os.Stat(filepath.Join(root, ".wiki")); !os.IsNotExist(statErr) {
		t.Fatalf("expected wiki to be opt-in and absent by default, got %v", statErr)
	}

	freshHome := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(freshHome, 0o755); err != nil {
		t.Fatalf("mkdir fresh home: %v", err)
	}
	taskPath, lookErr := exec.LookPath("task")
	if lookErr != nil {
		t.Fatalf("task binary not found: %v", lookErr)
	}
	cmd := exec.Command("task", "verify")
	cmd.Dir = root
	cmd.Env = testEnv(
		map[string]string{
			"HOME": freshHome,
			"PATH": filepath.Dir(taskPath) + string(os.PathListSeparator) + "/bin" + string(os.PathListSeparator) + "/usr/bin",
		},
		"CODEX_HOME",
		"CONVENTION_ENGINEERING_DIR",
	)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("expected task verify to pass: %v\n%s", runErr, output)
	}
}

func TestRunInitEmbedsBootstrapSourceRootAndUpdatedFallbacks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := RunInit(root, Options{
		Profiles:   []string{"go"},
		Operations: []string{"work"},
		RepoRisk:   "standard",
	}, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected init to succeed, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	config := readFileForTest(t, root, ConfigPath)
	if strings.Contains(config, "bootstrap_source_root") {
		t.Fatalf("did not expect machine-local bootstrap source to be written into %s:\n%s", ConfigPath, config)
	}

	script := readFileForTest(t, root, ConventionsCheckPath)
	bootstrapRoot := conventionEngineeringRootForTest(t)
	// New resolution chain (install-v2.md decision 4):
	//   $ARK_BINARY -> ark on $PATH -> `go run -C $repo_kit_root/cli ./cmd/ark` fallback.
	for _, marker := range []string{
		"bootstrap_source=" + shellSingleQuote(bootstrapRoot),
		"${ARK_BINARY:-}",
		`command -v ark`,
		`go run -C "$repo_kit_root/cli" ./cmd/ark`,
	} {
		if !strings.Contains(script, marker) {
			t.Fatalf("expected generated check script to contain %q, got:\n%s", marker, script)
		}
	}
	// Dropped surfaces (decision 4: no candidate-list shim):
	for _, stale := range []string{
		"$HOME/.claude/skills/agent-repo-kit/cli/bin/ark",
		"$HOME/.codex/skills/agent-repo-kit/cli/bin/ark",
		"$HOME/.agents/skills/agent-repo-kit/cli/bin/ark",
		"legacy fallback:",
	} {
		if strings.Contains(script, stale) {
			t.Fatalf("expected dropped fallback %q to be absent, got:\n%s", stale, script)
		}
	}

	envIdx := strings.Index(script, "${ARK_BINARY:-}")
	pathIdx := strings.Index(script, "command -v ark")
	goRunIdx := strings.Index(script, "go run -C")
	if envIdx < 0 || pathIdx < envIdx || goRunIdx < pathIdx {
		t.Fatalf("expected fallback order env -> PATH -> go run, got:\n%s", script)
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
		exitCode := RunInit(root, Options{
			Profiles:   []string{"go"},
			Operations: []string{"work"},
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
	if strings.Count(agents, managedBlockStart) != 1 || strings.Count(claude, managedBlockStart) != 1 {
		t.Fatalf("expected exactly one managed block in agent docs")
	}
	if strings.Count(taskfile, ConventionsTaskfile) != 1 {
		t.Fatalf("expected root Taskfile include to be inserted once, got:\n%s", taskfile)
	}
}

func TestRunInitAutoDetectsProfiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/init-test\n\ngo 1.22\n")
	writeFile(t, root, "package.json", "{\n  \"name\": \"init-test\"\n}\n")

	var out bytes.Buffer
	var err bytes.Buffer
	exitCode := RunInit(root, Options{}, &out, &err)
	if exitCode != 0 {
		t.Fatalf("expected init to succeed with auto-detected profiles, got %d stderr=%s stdout=%s", exitCode, err.String(), out.String())
	}

	cfg, _, loadErr := contract.LoadConfig(root, ConfigPath)
	if loadErr != nil {
		t.Fatalf("expected generated config to load: %v", loadErr)
	}
	if !containsString(cfg.Profiles, "go") || !containsString(cfg.Profiles, "typescript-react") {
		t.Fatalf("expected detected profiles to include go and typescript-react, got %#v", cfg.Profiles)
	}
}

// conventionEngineeringRootForTest resolves the same convention-engineering
// root that scaffold.resolveConventionEngineeringRoot() embeds in the
// generated check.sh. We re-implement the lookup here instead of calling the
// private helper so the test exercises the production path-resolution
// contract directly.
func conventionEngineeringRootForTest(t *testing.T) string {
	t.Helper()
	root, err := resolveConventionEngineeringRoot()
	if err != nil {
		t.Fatalf("failed to resolve convention-engineering root: %v", err)
	}
	return root
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
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

func testEnv(overrides map[string]string, dropKeys ...string) []string {
	drop := map[string]bool{}
	for _, key := range dropKeys {
		drop[key] = true
	}

	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || drop[key] {
			continue
		}
		if _, overridden := overrides[key]; overridden {
			continue
		}
		env = append(env, entry)
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}
