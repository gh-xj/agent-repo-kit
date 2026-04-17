package tasklint

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const methodDocsURL = "https://taskfile.dev/reference/schema/#task"

// validMethods lists the accepted values for both the top-level default
// and per-task `method:`.
var validMethods = map[string]struct{}{
	"checksum":  {},
	"timestamp": {},
	"none":      {},
}

// ruleMethodValidEnum (rule 8) — validates every `method:` occurrence.
func ruleMethodValidEnum(c *ruleContext) []Finding {
	var findings []Finding

	// Top-level default.
	if _, topMethod := findRootKey(c.rootNode, "method"); topMethod != nil {
		if f, ok := checkMethod(c.reportPath, topMethod, "top-level default"); !ok {
			findings = append(findings, f)
		}
	}

	// Per-task overrides.
	_, tasksNode := findRootKey(c.rootNode, "tasks")
	if tasksNode != nil && tasksNode.Kind == yaml.MappingNode {
		iterMapping(tasksNode, func(taskKey, taskBody *yaml.Node) {
			if taskBody == nil || taskBody.Kind != yaml.MappingNode {
				return
			}
			_, mNode := findRootKey(taskBody, "method")
			if mNode == nil {
				return
			}
			if f, ok := checkMethod(c.reportPath, mNode, fmt.Sprintf("task %q", taskKey.Value)); !ok {
				findings = append(findings, f)
			}
		})
	}
	return findings
}

// checkMethod returns (finding, valid=true) if the value is ok.
func checkMethod(path string, node *yaml.Node, where string) (Finding, bool) {
	raw := node.Value
	if _, ok := validMethods[raw]; ok {
		return Finding{}, true
	}
	return Finding{
		RuleID:   "method-valid-enum",
		Severity: SeverityError,
		Path:     path,
		Line:     node.Line,
		Column:   node.Column,
		Message:  fmt.Sprintf("invalid `method:` value %q at %s", raw, where),
		Detail:   "`method` must be one of: checksum, timestamp, none.",
		Fix:      "Change the value to `checksum`, `timestamp`, or `none`.",
		Docs:     methodDocsURL,
	}, false
}
