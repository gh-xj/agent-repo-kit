package contract

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckResult is a single pass/fail record produced by RunChecks.
type CheckResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// CountFailures returns the number of failing entries in results.
func CountFailures(results []CheckResult) int {
	failed := 0
	for _, r := range results {
		if !r.Passed {
			failed++
		}
	}
	return failed
}

// RunChecks executes every convention check configured in cfg against root
// and returns their results in configuration order.
func RunChecks(root string, cfg Config) []CheckResult {
	results := make([]CheckResult, 0)

	results = append(results, runRequiredFileChecks(root, cfg.RequiredFiles)...)
	results = append(results, runTaskfileChecks(root, cfg.TaskfileChecks)...)
	results = append(results, runCanonicalPointerChecks(root, cfg.CanonicalPointerMode, cfg.CanonicalPointers)...)
	results = append(results, runContentChecks(root, cfg.ContentChecks)...)
	results = append(results, runGitExcludeChecks(root, cfg.GitExcludeChecks)...)
	results = append(results, runInvariantChecks(root, cfg.InvariantContract)...)

	return results
}

func runRequiredFileChecks(root string, requiredFiles []string) []CheckResult {
	results := make([]CheckResult, 0, len(requiredFiles))
	for _, rel := range requiredFiles {
		full := filepath.Join(root, rel)
		if _, err := os.Stat(full); err != nil {
			results = append(results, CheckResult{Name: "file:" + rel, Passed: false, Detail: "missing"})
			continue
		}
		results = append(results, CheckResult{Name: "file:" + rel, Passed: true})
	}
	return results
}

func runTaskfileChecks(root string, taskfileChecks map[string][]string) []CheckResult {
	if len(taskfileChecks) == 0 {
		return []CheckResult{{Name: "taskfile-checks", Passed: true, Detail: "no taskfile checks configured"}}
	}

	results := make([]CheckResult, 0)
	for rel, tokens := range taskfileChecks {
		aggregateText, visited, err := aggregateTaskfileText(root, rel)
		if err != nil {
			results = append(results, CheckResult{Name: "taskfile:" + rel, Passed: false, Detail: err.Error()})
			continue
		}

		results = append(results, CheckResult{
			Name:   "taskfile:" + rel + ":includes",
			Passed: true,
			Detail: fmt.Sprintf("resolved %d taskfile(s)", len(visited)),
		})

		for _, token := range tokens {
			name := "task:" + rel + ":" + token
			if strings.Contains(aggregateText, token) {
				results = append(results, CheckResult{Name: name, Passed: true})
				continue
			}
			results = append(results, CheckResult{Name: name, Passed: false, Detail: "not found in taskfile include graph"})
		}
	}
	return results
}

func runCanonicalPointerChecks(root, mode string, pointers []CanonicalPointerConfig) []CheckResult {
	if len(pointers) == 0 {
		return []CheckResult{{Name: "canonical-pointers", Passed: true, Detail: "no canonical pointer checks configured"}}
	}

	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	if normalizedMode != "any" {
		normalizedMode = "all"
	}

	passed := make([]string, 0)
	failed := make([]string, 0)
	for _, pointer := range pointers {
		name := pointer.Name
		if name == "" {
			name = fmt.Sprintf("pointer:%s", pointer.File)
		}
		fullPath := filepath.Join(root, pointer.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (cannot read %s)", name, pointer.File))
			continue
		}
		if strings.Contains(string(content), pointer.Text) {
			passed = append(passed, name)
			continue
		}
		failed = append(failed, fmt.Sprintf("%s (text missing)", name))
	}

	if normalizedMode == "any" {
		if len(passed) > 0 {
			return []CheckResult{{
				Name:   "canonical-pointers:any",
				Passed: true,
				Detail: fmt.Sprintf("passed=%s", strings.Join(passed, ",")),
			}}
		}
		return []CheckResult{{
			Name:   "canonical-pointers:any",
			Passed: false,
			Detail: fmt.Sprintf("no pointer matched; failures=%s", strings.Join(failed, "; ")),
		}}
	}

	if len(failed) == 0 {
		return []CheckResult{{
			Name:   "canonical-pointers:all",
			Passed: true,
			Detail: fmt.Sprintf("all pointers matched (%d)", len(passed)),
		}}
	}
	return []CheckResult{{
		Name:   "canonical-pointers:all",
		Passed: false,
		Detail: fmt.Sprintf("failures=%s", strings.Join(failed, "; ")),
	}}
}

func runContentChecks(root string, checks []ContentCheckConfig) []CheckResult {
	if len(checks) == 0 {
		return []CheckResult{{Name: "content-checks", Passed: true, Detail: "no content checks configured"}}
	}

	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		name := check.Name
		if name == "" {
			name = "content:" + check.File
		}
		fullPath := filepath.Join(root, check.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			results = append(results, CheckResult{Name: name, Passed: false, Detail: "cannot read " + check.File})
			continue
		}
		text := string(content)
		missing := make([]string, 0)
		for _, marker := range check.RequiredMarkers {
			if strings.Contains(text, marker) {
				continue
			}
			missing = append(missing, marker)
		}
		if len(missing) == 0 {
			results = append(results, CheckResult{Name: name, Passed: true})
			continue
		}
		results = append(results, CheckResult{Name: name, Passed: false, Detail: "missing markers: " + strings.Join(missing, ",")})
	}
	return results
}

func runGitExcludeChecks(root string, checks []GitExcludeCheckConfig) []CheckResult {
	if len(checks) == 0 {
		return []CheckResult{{Name: "git-exclude-checks", Passed: true, Detail: "no git exclude checks configured"}}
	}

	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		name := strings.TrimSpace(check.Name)
		if name == "" {
			name = "git-exclude:" + check.File
		}

		fullPath := filepath.Join(root, check.File)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			results = append(results, CheckResult{Name: name, Passed: false, Detail: "cannot read " + check.File})
			continue
		}

		available := parseGitExcludePatterns(string(content))
		missing := make([]string, 0)
		for _, pattern := range check.RequiredPatterns {
			trimmed := strings.TrimSpace(pattern)
			if trimmed == "" {
				continue
			}
			if available[trimmed] {
				continue
			}
			missing = append(missing, trimmed)
		}

		if len(missing) == 0 {
			results = append(results, CheckResult{Name: name, Passed: true})
			continue
		}
		results = append(results, CheckResult{
			Name:   name,
			Passed: false,
			Detail: "missing patterns: " + strings.Join(missing, ","),
		})
	}
	return results
}

func parseGitExcludePatterns(content string) map[string]bool {
	patterns := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns[line] = true
	}
	return patterns
}
