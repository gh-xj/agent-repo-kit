package tasklint

import (
	"errors"
	"path/filepath"

	astpkg "github.com/go-task/task/v3/taskfile/ast"
	"gopkg.in/yaml.v3"
)

// parsedTaskfile bundles the two parallel parses used by the rules.
//
// We keep both because the upstream AST gives typed access (task
// locations, parsed semver) but drops unknown keys, and the raw YAML
// node tree retains every key + line/column but is untyped.
type parsedTaskfile struct {
	ast      *astpkg.Taskfile // may be nil if ast parse failed
	astErr   error
	root     *yaml.Node // MappingNode at document root, or nil
	document *yaml.Node // the Document node, or nil
	parseErr error      // set only when we cannot obtain any usable YAML structure
}

// parseTaskfile performs both parses against the same bytes.
// If the raw YAML parse fails, parseErr is set. The upstream AST
// parse is best-effort: a decode failure is kept in astErr (rules fall
// back to the raw YAML node tree for locations).
func parseTaskfile(data []byte) parsedTaskfile {
	var out parsedTaskfile

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		out.parseErr = err
		return out
	}
	out.document = &doc
	if len(doc.Content) > 0 {
		out.root = doc.Content[0]
	}

	// Reject non-mapping roots (e.g. a list or a scalar file). These
	// will never satisfy the rules and produce misleading findings.
	if out.root == nil || out.root.Kind != yaml.MappingNode {
		out.parseErr = errors.New("taskfile root is not a mapping")
		return out
	}

	var tf astpkg.Taskfile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		// Keep the error but continue — rules that rely on the raw
		// YAML tree can still run.
		out.astErr = err
	} else {
		out.ast = &tf
	}

	return out
}

// parseErrorFinding converts a YAML parse error into a single finding.
func parseErrorFinding(path string, err error) Finding {
	line, col := extractYAMLLocation(err)
	return Finding{
		RuleID:   "parse-error",
		Severity: SeverityError,
		Path:     path,
		Line:     line,
		Column:   col,
		Message:  "Taskfile is not valid YAML",
		Detail:   err.Error(),
		Fix:      "Open the file in an editor and fix the YAML syntax error reported above.",
		Docs:     "https://taskfile.dev/reference/schema/",
	}
}

// extractYAMLLocation pulls a line number out of a yaml.TypeError or
// a yaml syntax error when it is formatted as "yaml: line N: ...".
func extractYAMLLocation(err error) (int, int) {
	if err == nil {
		return 0, 0
	}
	msg := err.Error()
	// yaml.v3 formats syntax errors as "yaml: line 3: ..." (1-based).
	const prefix = "yaml: line "
	if idx := indexOf(msg, prefix); idx >= 0 {
		rest := msg[idx+len(prefix):]
		var n int
		for i := 0; i < len(rest); i++ {
			c := rest[i]
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		if n > 0 {
			return n, 0
		}
	}
	return 0, 0
}

func indexOf(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// effectiveRepoRoot returns opts.RepoRoot if set, otherwise the
// directory containing the taskfile.
func effectiveRepoRoot(taskfilePath, repoRoot string) string {
	if repoRoot != "" {
		return repoRoot
	}
	abs, err := filepath.Abs(taskfilePath)
	if err != nil {
		return filepath.Dir(taskfilePath)
	}
	return filepath.Dir(abs)
}

// displayPath turns an absolute taskfile path into a repo-relative one
// when possible, for nicer finding output.
func displayPath(taskfilePath, repoRoot string) string {
	if repoRoot == "" {
		return taskfilePath
	}
	absTask, err := filepath.Abs(taskfilePath)
	if err != nil {
		return taskfilePath
	}
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return taskfilePath
	}
	rel, err := filepath.Rel(absRoot, absTask)
	if err != nil {
		return taskfilePath
	}
	return rel
}
