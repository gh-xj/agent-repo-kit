package tasklint

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestRulesDocsDrift ensures the set of rule IDs in
// `taskfile-authoring/references/lint-rules.md` equals the set
// returned by Rules(). If either side has an ID the other lacks,
// the doc or the code is stale.
func TestRulesDocsDrift(t *testing.T) {
	// Resolve the doc path relative to this test file, not the cwd.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	// this file lives at <repo>/cli/internal/tasklint/<...>_test.go.
	// Walk up to <repo>/ and then into taskfile-authoring/...
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	docPath := filepath.Join(repoRoot, "taskfile-authoring", "references", "lint-rules.md")

	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read lint-rules.md at %s: %v", docPath, err)
	}

	doc := string(data)

	// Extract backticked kebab-case IDs that appear in a rule-heading
	// context. We recognise the pattern `## N. ` + ` `` <id> `` ` as
	// the canonical per-rule header.
	headingRE := regexp.MustCompile("(?m)^##\\s+\\d+\\.\\s+`([a-z][a-z0-9]*(?:-[a-z0-9]+)+)`")
	matches := headingRE.FindAllStringSubmatch(doc, -1)
	if len(matches) == 0 {
		t.Fatalf("no rule headings found in %s — regex may be stale", docPath)
	}

	doced := map[string]bool{}
	for _, m := range matches {
		doced[m[1]] = true
	}

	catalog := map[string]bool{}
	for _, r := range Rules() {
		catalog[r.ID] = true
	}

	var missingFromDocs, missingFromCode []string
	for id := range catalog {
		if !doced[id] {
			missingFromDocs = append(missingFromDocs, id)
		}
	}
	for id := range doced {
		if !catalog[id] {
			missingFromCode = append(missingFromCode, id)
		}
	}
	sort.Strings(missingFromDocs)
	sort.Strings(missingFromCode)

	if len(missingFromDocs) > 0 || len(missingFromCode) > 0 {
		var b strings.Builder
		if len(missingFromDocs) > 0 {
			b.WriteString("Rule IDs present in Rules() but missing from lint-rules.md:\n")
			for _, id := range missingFromDocs {
				b.WriteString("  - ")
				b.WriteString(id)
				b.WriteString("\n")
			}
		}
		if len(missingFromCode) > 0 {
			b.WriteString("Rule IDs present in lint-rules.md but missing from Rules():\n")
			for _, id := range missingFromCode {
				b.WriteString("  - ")
				b.WriteString(id)
				b.WriteString("\n")
			}
		}
		t.Fatalf("rule-id drift between code and docs:\n%s", b.String())
	}
}
