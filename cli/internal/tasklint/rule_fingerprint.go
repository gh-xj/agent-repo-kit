package tasklint

import (
	"gopkg.in/yaml.v3"
)

const fingerprintDocsURL = "https://taskfile.dev/reference/schema/#task"

// ruleFingerprintDirGitignored (rule 9) — if any task declares
// `sources:`, the repo's .gitignore must cover `.task/`.
func ruleFingerprintDirGitignored(c *ruleContext) []Finding {
	_, tasksNode := findRootKey(c.rootNode, "tasks")
	if tasksNode == nil || tasksNode.Kind != yaml.MappingNode {
		return nil
	}
	// Find the first task that uses sources; we anchor the finding there.
	var firstTaskWithSources *yaml.Node
	iterMapping(tasksNode, func(taskKey, taskBody *yaml.Node) {
		if firstTaskWithSources != nil {
			return
		}
		if taskBody == nil || taskBody.Kind != yaml.MappingNode {
			return
		}
		if srcKey, _ := findRootKey(taskBody, "sources"); srcKey != nil {
			firstTaskWithSources = srcKey
		}
	})
	if firstTaskWithSources == nil {
		return nil
	}
	gi := c.gitignore()
	// Accept any of the common literal forms for `.task`.
	if gi.Exists() && (gi.HasLiteralPattern(".task", ".task/", "/.task", "/.task/") || gi.Matches(".task/fingerprint.txt")) {
		return nil
	}
	return []Finding{{
		RuleID:   "fingerprint-dir-gitignored",
		Severity: SeverityError,
		Path:     c.reportPath,
		Line:     firstTaskWithSources.Line,
		Column:   firstTaskWithSources.Column,
		Message:  "`.task/` is not in .gitignore but at least one task uses `sources:`",
		Detail:   "Task writes fingerprints under `<repo>/.task/`. Without a gitignore entry those files will be committed accidentally.",
		Fix:      "Add a line `.task/` to the repository's .gitignore.",
		Docs:     fingerprintDocsURL,
	}}
}
