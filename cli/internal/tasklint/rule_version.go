package tasklint

import (
	"fmt"
	"regexp"
	"strings"
)

const versionDocsURL = "https://taskfile.dev/reference/schema/#schema"

// versionPattern matches `3` or `3.<digits>` (as required by the rule set).
var versionPattern = regexp.MustCompile(`^3(\.\d+)?$`)

// ruleVersionRequired (rule 1) — top-level `version:` must exist.
func ruleVersionRequired(c *ruleContext) []Finding {
	keyNode, _ := findRootKey(c.rootNode, "version")
	if keyNode != nil {
		return nil
	}
	line, col := 1, 1
	if c.rootNode != nil && c.rootNode.Line > 0 {
		line, col = c.rootNode.Line, c.rootNode.Column
	}
	return []Finding{{
		RuleID:   "version-required",
		Severity: SeverityError,
		Path:     c.reportPath,
		Line:     line,
		Column:   col,
		Message:  "Taskfile is missing the top-level `version:` key",
		Detail:   "Task requires every Taskfile to declare a schema version.",
		Fix:      "Add `version: '3'` at the top of the file.",
		Docs:     versionDocsURL,
	}}
}

// ruleVersionIsThree (rule 2) — version value must match ^3(\.\d+)?$.
func ruleVersionIsThree(c *ruleContext) []Finding {
	keyNode, valueNode := findRootKey(c.rootNode, "version")
	if keyNode == nil || valueNode == nil {
		// rule 1 already reports the missing-version case.
		return nil
	}
	raw := strings.TrimSpace(valueNode.Value)
	if versionPattern.MatchString(raw) {
		return nil
	}
	return []Finding{{
		RuleID:   "version-is-three",
		Severity: SeverityError,
		Path:     c.reportPath,
		Line:     valueNode.Line,
		Column:   valueNode.Column,
		Message:  "Taskfile version must be 3 (or 3.x)",
		Detail:   fmt.Sprintf("found version %q — only version 3 of the Taskfile schema is supported by this linter.", raw),
		Fix:      "Change the `version:` value to `'3'`.",
		Docs:     versionDocsURL,
	}}
}
