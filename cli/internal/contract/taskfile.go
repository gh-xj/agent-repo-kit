package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var taskfilePathLinePattern = regexp.MustCompile(`(?m)^\s*taskfile:\s+(.+?)\s*$`)

func aggregateTaskfileText(root, relativeTaskfile string) (string, []string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, fmt.Errorf("resolve root: %w", err)
	}
	start := filepath.Join(absRoot, relativeTaskfile)
	visitedOrder := make([]string, 0)
	visited := map[string]bool{}
	builder := strings.Builder{}

	var walk func(string) error
	walk = func(path string) error {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		// Reject paths outside the repo root. A Taskfile with
		// `taskfile: /etc/passwd` or `taskfile: ../../escape.yml` would
		// otherwise cause the checker to read arbitrary files on the host.
		rel, relErr := filepath.Rel(absRoot, absPath)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("taskfile include path escapes repo root: %s", absPath)
		}
		if visited[absPath] {
			return nil
		}
		visited[absPath] = true
		visitedOrder = append(visitedOrder, absPath)

		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		builder.WriteString("\n# file: ")
		builder.WriteString(absPath)
		builder.WriteString("\n")
		builder.Write(content)
		builder.WriteString("\n")

		for _, includeRel := range extractIncludeTaskfiles(string(content)) {
			if strings.Contains(includeRel, "{{") {
				continue
			}
			nextPath := includeRel
			if !filepath.IsAbs(nextPath) {
				nextPath = filepath.Join(filepath.Dir(absPath), includeRel)
			}
			if err := walk(nextPath); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(start); err != nil {
		return "", nil, fmt.Errorf("cannot resolve taskfile include graph from %s: %w", relativeTaskfile, err)
	}
	return builder.String(), visitedOrder, nil
}

func extractIncludeTaskfiles(content string) []string {
	matches := taskfilePathLinePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value := strings.TrimSpace(match[1])
		value = strings.Trim(value, "\"'")
		if value == "" {
			continue
		}
		paths = append(paths, value)
	}
	return paths
}
