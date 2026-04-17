// Package tasklint lints Taskfile.yml against structural rules.
//
// V1 ships a fixed 10-rule set focused on structural correctness:
// missing/invalid version, unknown keys, mutually exclusive fields,
// unresolvable includes, flatten name collisions, invalid method enums,
// and gitignore coverage for fingerprint/dotenv files.
//
// Future rule layers (stack detection, per-repo config, --strict mode)
// are intentionally deferred.
package tasklint

import (
	"fmt"
	"os"
	"sort"
)

// Severity describes how strictly a finding should be surfaced.
// V1 emits only SeverityError — every structural rule is an error.
type Severity int

const (
	// SeverityError marks a structural violation. The task is broken or unsafe.
	SeverityError Severity = iota
)

// String renders the severity for human-facing output.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	default:
		return fmt.Sprintf("severity(%d)", int(s))
	}
}

// Finding is a single rule violation reported by Lint.
type Finding struct {
	RuleID   string   // stable kebab-case id, e.g. "version-required"
	Severity Severity // V1 only emits SeverityError
	Path     string   // absolute or repo-relative taskfile path
	Line     int      // 1-based; 0 if unknown
	Column   int      // 1-based; 0 if unknown
	Message  string   // one-line statement of what is wrong
	Detail   string   // context, e.g. "found key 'variables:' — did you mean 'vars:'?"
	Fix      string   // concrete one-line how-to
	Docs     string   // reference URL (upstream or repo-local reference)
}

// Options tunes Lint behaviour.
type Options struct {
	// RepoRoot is used for relative-path resolution (e.g. the Path field
	// in findings) and to locate <repo_root>/.gitignore.
	// If empty, the directory containing the taskfile is used.
	RepoRoot string
}

// Lint runs all V1 rules against the Taskfile at taskfilePath.
// Findings are returned sorted by (path, line, column, ruleID).
//
// Only genuine I/O failures return a non-nil error. Malformed YAML is
// reported as a single "parse-error" finding so callers can treat it
// uniformly with other rule violations.
func Lint(taskfilePath string, opts Options) ([]Finding, error) {
	data, err := os.ReadFile(taskfilePath)
	if err != nil {
		return nil, fmt.Errorf("tasklint: read %s: %w", taskfilePath, err)
	}

	parsed := parseTaskfile(data)
	reportPath := displayPath(taskfilePath, opts.RepoRoot)

	if parsed.parseErr != nil {
		return []Finding{parseErrorFinding(reportPath, parsed.parseErr)}, nil
	}

	ctx := &ruleContext{
		path:         taskfilePath,
		reportPath:   reportPath,
		repoRoot:     effectiveRepoRoot(taskfilePath, opts.RepoRoot),
		astFile:      parsed.ast,
		astParseErr:  parsed.astErr,
		rootNode:     parsed.root,
		documentNode: parsed.document,
	}

	var findings []Finding
	for _, rule := range ruleFuncs {
		findings = append(findings, rule(ctx)...)
	}

	sortFindings(findings)
	return findings, nil
}

// sortFindings orders by (path, line, column, ruleID) for stable output.
func sortFindings(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].Path != fs[j].Path {
			return fs[i].Path < fs[j].Path
		}
		if fs[i].Line != fs[j].Line {
			return fs[i].Line < fs[j].Line
		}
		if fs[i].Column != fs[j].Column {
			return fs[i].Column < fs[j].Column
		}
		return fs[i].RuleID < fs[j].RuleID
	})
}
