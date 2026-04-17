package tasklint

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// testFixture is the shared builder used by every per-rule test. It
// writes a Taskfile (and optional .gitignore + include files) into a
// temp directory and runs Lint against the result.
type testFixture struct {
	t           *testing.T
	taskfile    string // contents of Taskfile.yml
	gitignore   string // contents of .gitignore (if non-empty)
	includeFiles map[string]string // repo-relative path -> contents
}

func (f *testFixture) run() []Finding {
	f.t.Helper()
	dir := f.t.TempDir()
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	if err := os.WriteFile(taskfilePath, []byte(f.taskfile), 0o644); err != nil {
		f.t.Fatalf("write taskfile: %v", err)
	}
	if f.gitignore != "" {
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(f.gitignore), 0o644); err != nil {
			f.t.Fatalf("write gitignore: %v", err)
		}
	}
	for rel, contents := range f.includeFiles {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			f.t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
			f.t.Fatalf("write include %s: %v", rel, err)
		}
	}
	findings, err := Lint(taskfilePath, Options{RepoRoot: dir})
	if err != nil {
		f.t.Fatalf("Lint: %v", err)
	}
	return findings
}

// findingsByRule returns findings whose RuleID equals id.
func findingsByRule(fs []Finding, id string) []Finding {
	var out []Finding
	for _, f := range fs {
		if f.RuleID == id {
			out = append(out, f)
		}
	}
	return out
}

// assertHasRule asserts at least one finding with the given rule id is
// present and that each such finding has non-empty Message/Fix.
func assertHasRule(t *testing.T, findings []Finding, ruleID string) []Finding {
	t.Helper()
	got := findingsByRule(findings, ruleID)
	if len(got) == 0 {
		t.Fatalf("expected at least one %s finding; got %d findings total:\n%s",
			ruleID, len(findings), dumpFindings(findings))
	}
	for _, f := range got {
		if f.Message == "" {
			t.Errorf("%s finding has empty Message", ruleID)
		}
		if f.Fix == "" {
			t.Errorf("%s finding has empty Fix", ruleID)
		}
	}
	return got
}

// assertRuleAbsent asserts no findings with that rule id are present.
func assertRuleAbsent(t *testing.T, findings []Finding, ruleID string) {
	t.Helper()
	if got := findingsByRule(findings, ruleID); len(got) > 0 {
		t.Fatalf("expected zero %s findings, got %d:\n%s",
			ruleID, len(got), dumpFindings(got))
	}
}

func dumpFindings(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.RuleID)
		b.WriteString(": ")
		b.WriteString(f.Message)
		b.WriteString("\n")
	}
	return b.String()
}

// ------------------------------------------------------------------
// End-to-end tests
// ------------------------------------------------------------------

func TestLintParseError(t *testing.T) {
	fx := &testFixture{
		t:        t,
		taskfile: "version: '3'\ntasks:\n  build:\n    cmds:\n  - : : not yaml\n",
	}
	findings := fx.run()
	if len(findings) != 1 {
		t.Fatalf("expected exactly one finding, got %d:\n%s", len(findings), dumpFindings(findings))
	}
	if findings[0].RuleID != "parse-error" {
		t.Fatalf("expected rule parse-error, got %q", findings[0].RuleID)
	}
	if findings[0].Message == "" || findings[0].Detail == "" {
		t.Fatalf("parse-error finding missing Message/Detail: %+v", findings[0])
	}
}

func TestLintRulesCatalogMatchesImplemented(t *testing.T) {
	catalog := Rules()
	if len(catalog) != len(ruleFuncs) {
		t.Fatalf("catalog has %d rules but ruleFuncs has %d", len(catalog), len(ruleFuncs))
	}

	kebab := regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	seen := map[string]bool{}
	for _, r := range catalog {
		if !kebab.MatchString(r.ID) {
			t.Errorf("rule ID %q is not kebab-case", r.ID)
		}
		if seen[r.ID] {
			t.Errorf("duplicate rule ID %q", r.ID)
		}
		seen[r.ID] = true
		if r.Title == "" {
			t.Errorf("rule %q missing Title", r.ID)
		}
		if r.Description == "" {
			t.Errorf("rule %q missing Description", r.ID)
		}
	}

	// The ten documented V1 rule IDs must all be present.
	expected := []string{
		"version-required",
		"version-is-three",
		"unknown-top-level-keys",
		"unknown-task-keys",
		"cmd-and-cmds-mutex",
		"includes-paths-resolvable",
		"flatten-no-name-collision",
		"method-valid-enum",
		"fingerprint-dir-gitignored",
		"dotenv-files-gitignored",
	}
	for _, id := range expected {
		if !seen[id] {
			t.Errorf("catalog missing expected rule %q", id)
		}
	}
}

func TestLintOnRepoOwnCliTaskfile(t *testing.T) {
	// Locate the repo's cli/Taskfile.yml relative to this test file.
	// From cli/internal/tasklint/, that's ../../Taskfile.yml.
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	taskfilePath := filepath.Clean(filepath.Join(pkgDir, "..", "..", "Taskfile.yml"))
	repoRoot := filepath.Clean(filepath.Join(pkgDir, "..", ".."))

	if _, err := os.Stat(taskfilePath); err != nil {
		t.Fatalf("cannot find %s: %v", taskfilePath, err)
	}

	findings, err := Lint(taskfilePath, Options{RepoRoot: repoRoot})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected zero findings on cli/Taskfile.yml, got %d:\n%s",
			len(findings), dumpFindings(findings))
	}
}

func TestLintSortedOutput(t *testing.T) {
	// Craft a file that triggers two rules at known different positions.
	fx := &testFixture{
		t: t,
		taskfile: `version: '3'
tasks:
  build:
    cmd: echo a
    cmds: [echo b]
  test:
    method: banana
`,
	}
	findings := fx.run()
	if len(findings) < 2 {
		t.Fatalf("expected >= 2 findings, got %d:\n%s", len(findings), dumpFindings(findings))
	}
	// Walk the slice and verify the sort key is monotonically non-decreasing.
	for i := 1; i < len(findings); i++ {
		a, b := findings[i-1], findings[i]
		if a.Path > b.Path {
			t.Fatalf("findings[%d].Path > findings[%d].Path", i-1, i)
		}
		if a.Path == b.Path && a.Line > b.Line {
			t.Fatalf("findings[%d].Line > findings[%d].Line", i-1, i)
		}
		if a.Path == b.Path && a.Line == b.Line && a.Column > b.Column {
			t.Fatalf("findings[%d].Column > findings[%d].Column", i-1, i)
		}
		if a.Path == b.Path && a.Line == b.Line && a.Column == b.Column && a.RuleID > b.RuleID {
			t.Fatalf("findings[%d].RuleID > findings[%d].RuleID", i-1, i)
		}
	}
	// Extra: `sort.SliceIsSorted` with the same key.
	ok := sort.SliceIsSorted(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		if findings[i].Column != findings[j].Column {
			return findings[i].Column < findings[j].Column
		}
		return findings[i].RuleID < findings[j].RuleID
	})
	if !ok {
		t.Fatalf("findings not sorted:\n%s", dumpFindings(findings))
	}
}
