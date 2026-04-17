package tasklint

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const cmdDocsURL = "https://taskfile.dev/reference/schema/#task"

// ruleCmdAndCmdsMutex (rule 5) — a task may define either `cmd:` or
// `cmds:`, never both. The upstream parser also rejects this but we
// want a friendlier message that survives missing-version cases.
func ruleCmdAndCmdsMutex(c *ruleContext) []Finding {
	_, tasksNode := findRootKey(c.rootNode, "tasks")
	if tasksNode == nil || tasksNode.Kind != yaml.MappingNode {
		return nil
	}
	var findings []Finding
	iterMapping(tasksNode, func(taskKey, taskBody *yaml.Node) {
		if taskBody == nil || taskBody.Kind != yaml.MappingNode {
			return
		}
		var cmdKey, cmdsKey *yaml.Node
		iterMapping(taskBody, func(key, _ *yaml.Node) {
			switch key.Value {
			case "cmd":
				cmdKey = key
			case "cmds":
				cmdsKey = key
			}
		})
		if cmdKey == nil || cmdsKey == nil {
			return
		}
		// Anchor the finding to whichever key appears second — that's
		// the more helpful location for the user to delete.
		anchor := cmdKey
		if cmdsKey.Line > cmdKey.Line || (cmdsKey.Line == cmdKey.Line && cmdsKey.Column > cmdKey.Column) {
			anchor = cmdsKey
		}
		findings = append(findings, Finding{
			RuleID:   "cmd-and-cmds-mutex",
			Severity: SeverityError,
			Path:     c.reportPath,
			Line:     anchor.Line,
			Column:   anchor.Column,
			Message:  fmt.Sprintf("task %q defines both `cmd:` and `cmds:`", taskKey.Value),
			Detail:   "`cmd:` is the shortcut form for a single command; `cmds:` is the list form. Only one may be set.",
			Fix:      fmt.Sprintf("Remove either `cmd:` or `cmds:` from task %q.", taskKey.Value),
			Docs:     cmdDocsURL,
		})
	})
	return findings
}
