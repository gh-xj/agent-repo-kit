package tasklint

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const dotenvDocsURL = "https://taskfile.dev/reference/schema/#schema"

// ruleDotenvFilesGitignored (rule 10) — every entry in any `dotenv:`
// list must be matched by the repo's .gitignore unless its basename
// matches `*.example` or `*.sample`.
func ruleDotenvFilesGitignored(c *ruleContext) []Finding {
	gi := c.gitignore()
	var findings []Finding

	// Collect (path-string, anchor yaml.Node, origin label).
	type dotenvItem struct {
		value  string
		node   *yaml.Node
		origin string
	}
	var items []dotenvItem

	if _, topDotenv := findRootKey(c.rootNode, "dotenv"); topDotenv != nil && topDotenv.Kind == yaml.SequenceNode {
		for _, item := range topDotenv.Content {
			items = append(items, dotenvItem{value: item.Value, node: item, origin: "top-level"})
		}
	}

	_, tasksNode := findRootKey(c.rootNode, "tasks")
	if tasksNode != nil && tasksNode.Kind == yaml.MappingNode {
		iterMapping(tasksNode, func(taskKey, taskBody *yaml.Node) {
			if taskBody == nil || taskBody.Kind != yaml.MappingNode {
				return
			}
			_, dotenvNode := findRootKey(taskBody, "dotenv")
			if dotenvNode == nil || dotenvNode.Kind != yaml.SequenceNode {
				return
			}
			origin := fmt.Sprintf("task %q", taskKey.Value)
			for _, item := range dotenvNode.Content {
				items = append(items, dotenvItem{value: item.Value, node: item, origin: origin})
			}
		})
	}

	for _, it := range items {
		if isDotenvExample(it.value) {
			continue
		}
		if gi.Exists() && gi.Matches(it.value) {
			continue
		}
		findings = append(findings, Finding{
			RuleID:   "dotenv-files-gitignored",
			Severity: SeverityError,
			Path:     c.reportPath,
			Line:     it.node.Line,
			Column:   it.node.Column,
			Message:  fmt.Sprintf("dotenv file %q is not gitignored", it.value),
			Detail:   fmt.Sprintf("referenced from %s; dotenv files commonly contain secrets and should not be tracked.", it.origin),
			Fix:      fmt.Sprintf("Add `%s` (or a matching pattern) to .gitignore, or rename the file to match `*.example` / `*.sample`.", it.value),
			Docs:     dotenvDocsURL,
		})
	}
	return findings
}

// isDotenvExample returns true if the basename matches `*.example` or
// `*.sample`, the documented opt-out.
func isDotenvExample(p string) bool {
	base := filepath.Base(p)
	// filepath.Ext returns the last ".foo"; we want to check both
	// "foo.example" and "foo.env.example". strings.HasSuffix handles both.
	return strings.HasSuffix(base, ".example") || strings.HasSuffix(base, ".sample")
}
