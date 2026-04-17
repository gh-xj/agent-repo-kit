package tasklint

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// allowedTopLevelKeys is the full allowlist for the root mapping. Keep
// in sync with the design doc's V1 rule list.
var allowedTopLevelKeys = map[string]struct{}{
	"version":      {},
	"tasks":        {},
	"includes":     {},
	"vars":         {},
	"env":          {},
	"dotenv":       {},
	"output":       {},
	"method":       {},
	"run":          {},
	"set":          {},
	"shopt":        {},
	"interval":     {},
	"silent":       {},
	"requires":     {},
	"experimental": {},
}

// allowedTaskKeys is the allowlist for each task's mapping.
var allowedTaskKeys = map[string]struct{}{
	"desc":         {},
	"summary":      {},
	"aliases":      {},
	"cmds":         {},
	"cmd":          {},
	"deps":         {},
	"preconditions": {},
	"requires":     {},
	"sources":      {},
	"generates":    {},
	"method":       {},
	"status":       {},
	"dir":          {},
	"env":          {},
	"vars":         {},
	"silent":       {},
	"internal":     {},
	"interactive":  {},
	"ignore_error": {},
	"run":          {},
	"platforms":    {},
	"prefix":       {},
	"label":        {},
	"prompt":       {},
	"watch":        {},
	"set":          {},
	"shopt":        {},
	"defer":        {},
	"for":          {},
	"if":           {},
	"output":       {},
}

// commonTypos maps a known-bad root key to its likely intended replacement.
var commonRootTypos = map[string]string{
	"variables": "vars",
	"var":       "vars",
	"task":      "tasks",
	"include":   "includes",
	"dot_env":   "dotenv",
}

const (
	topLevelDocsURL = "https://taskfile.dev/reference/schema/#schema"
	taskKeyDocsURL  = "https://taskfile.dev/reference/schema/#task"
)

// ruleUnknownTopLevelKeys (rule 3) — reject any root key not in the allowlist.
func ruleUnknownTopLevelKeys(c *ruleContext) []Finding {
	var findings []Finding
	iterMapping(c.rootNode, func(key, _ *yaml.Node) {
		if _, ok := allowedTopLevelKeys[key.Value]; ok {
			return
		}
		detail := fmt.Sprintf("found key %q which is not part of the Taskfile schema.", key.Value)
		fix := fmt.Sprintf("Remove `%s:` or rename it to a supported top-level key (%s).", key.Value, allowedListSummary(allowedTopLevelKeys))
		if suggestion, ok := commonRootTypos[key.Value]; ok {
			detail = fmt.Sprintf("found key %q — did you mean `%s:`?", key.Value, suggestion)
			fix = fmt.Sprintf("Rename `%s:` to `%s:`.", key.Value, suggestion)
		}
		findings = append(findings, Finding{
			RuleID:   "unknown-top-level-keys",
			Severity: SeverityError,
			Path:     c.reportPath,
			Line:     key.Line,
			Column:   key.Column,
			Message:  fmt.Sprintf("unknown top-level key `%s:`", key.Value),
			Detail:   detail,
			Fix:      fix,
			Docs:     topLevelDocsURL,
		})
	})
	return findings
}

// ruleUnknownTaskKeys (rule 4) — reject unknown keys inside any task.
func ruleUnknownTaskKeys(c *ruleContext) []Finding {
	_, tasksNode := findRootKey(c.rootNode, "tasks")
	if tasksNode == nil || tasksNode.Kind != yaml.MappingNode {
		return nil
	}
	var findings []Finding
	iterMapping(tasksNode, func(taskKey, taskBody *yaml.Node) {
		// Shortcut-syntax tasks (a scalar command or a sequence of
		// commands) have no keys to validate.
		if taskBody == nil || taskBody.Kind != yaml.MappingNode {
			return
		}
		iterMapping(taskBody, func(key, _ *yaml.Node) {
			if _, ok := allowedTaskKeys[key.Value]; ok {
				return
			}
			findings = append(findings, Finding{
				RuleID:   "unknown-task-keys",
				Severity: SeverityError,
				Path:     c.reportPath,
				Line:     key.Line,
				Column:   key.Column,
				Message:  fmt.Sprintf("unknown task key `%s:` in task %q", key.Value, taskKey.Value),
				Detail:   fmt.Sprintf("`%s` is not a recognised task property.", key.Value),
				Fix:      fmt.Sprintf("Remove `%s:` from task %q or rename it to a supported key.", key.Value, taskKey.Value),
				Docs:     taskKeyDocsURL,
			})
		})
	})
	return findings
}

// allowedListSummary renders the allowlist for human-friendly fix text.
func allowedListSummary(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return joinWithCommas(keys)
}

func joinWithCommas(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += ", " + p
	}
	return out
}
