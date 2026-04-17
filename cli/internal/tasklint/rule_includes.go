package tasklint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const includesDocsURL = "https://taskfile.dev/reference/schema/#include"

// remoteSchemes are include paths we don't validate (upstream remote
// taskfiles are experimental and resolution requires network).
var remoteSchemes = []string{
	"http://",
	"https://",
	"git+https://",
	"git+ssh://",
}

// includeEntry is a parsed view of a single `includes:` entry.
type includeEntry struct {
	namespace string
	taskfile  string // may be empty for shortcut "include: path"
	optional  bool
	flatten   bool
	excludes  []string
	nameNode  *yaml.Node // the namespace key node (used for locations)
	pathNode  *yaml.Node // the node carrying the taskfile path
}

// ruleIncludesPathsResolvable (rule 6) — every include's `taskfile:`
// path must exist on disk unless the include is `optional:` or uses a
// remote scheme.
func ruleIncludesPathsResolvable(c *ruleContext) []Finding {
	entries := collectIncludes(c.rootNode)
	if len(entries) == 0 {
		return nil
	}
	taskfileDir := filepath.Dir(c.path)

	var findings []Finding
	for _, e := range entries {
		if e.taskfile == "" {
			continue // shortcut form with no path — nothing to check
		}
		if e.optional {
			continue
		}
		if isRemoteInclude(e.taskfile) {
			continue
		}
		resolved := e.taskfile
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(taskfileDir, resolved)
		}
		if info, err := os.Stat(resolved); err == nil {
			if info.IsDir() {
				// Task auto-resolves a directory to its Taskfile.yml.
				if _, err := os.Stat(filepath.Join(resolved, "Taskfile.yml")); err == nil {
					continue
				}
				if _, err := os.Stat(filepath.Join(resolved, "Taskfile.yaml")); err == nil {
					continue
				}
			} else {
				continue
			}
		}
		anchor := e.pathNode
		if anchor == nil {
			anchor = e.nameNode
		}
		findings = append(findings, Finding{
			RuleID:   "includes-paths-resolvable",
			Severity: SeverityError,
			Path:     c.reportPath,
			Line:     anchor.Line,
			Column:   anchor.Column,
			Message:  fmt.Sprintf("include %q points at %q which does not exist", e.namespace, e.taskfile),
			Detail:   fmt.Sprintf("resolved to %s; no such file or directory", resolved),
			Fix:      fmt.Sprintf("Fix the path, delete the include, or set `optional: true` under `%s:`.", e.namespace),
			Docs:     includesDocsURL,
		})
	}
	return findings
}

// ruleFlattenNoNameCollision (rule 7) — when an include sets
// `flatten: true`, none of the included file's task names may collide
// with root-level tasks or with other flattened tasks.
func ruleFlattenNoNameCollision(c *ruleContext) []Finding {
	entries := collectIncludes(c.rootNode)
	if len(entries) == 0 {
		return nil
	}
	taskfileDir := filepath.Dir(c.path)

	// Root-level task name set.
	rootNames := map[string]bool{}
	if _, tasksNode := findRootKey(c.rootNode, "tasks"); tasksNode != nil && tasksNode.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(tasksNode.Content); i += 2 {
			rootNames[tasksNode.Content[i].Value] = true
		}
	}

	// Tracks who owns each flattened name, so we can report a clear
	// "namespace A and namespace B both introduce foo" message.
	flattenedOwner := map[string]string{}

	var findings []Finding
	for _, e := range entries {
		if !e.flatten {
			continue
		}
		if e.taskfile == "" || isRemoteInclude(e.taskfile) {
			continue
		}
		resolved := e.taskfile
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(taskfileDir, resolved)
		}
		if info, err := os.Stat(resolved); err == nil && info.IsDir() {
			// Prefer Taskfile.yml, fall back to Taskfile.yaml.
			if _, err := os.Stat(filepath.Join(resolved, "Taskfile.yml")); err == nil {
				resolved = filepath.Join(resolved, "Taskfile.yml")
			} else if _, err := os.Stat(filepath.Join(resolved, "Taskfile.yaml")); err == nil {
				resolved = filepath.Join(resolved, "Taskfile.yaml")
			}
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			// rule 6 surfaces the unresolved-path case; skip here.
			continue
		}
		includedNames, err := extractTaskNames(data)
		if err != nil {
			continue
		}
		excludeSet := map[string]bool{}
		for _, x := range e.excludes {
			excludeSet[x] = true
		}
		for _, name := range includedNames {
			if excludeSet[name] {
				continue
			}
			var collidesWith string
			switch {
			case rootNames[name]:
				collidesWith = fmt.Sprintf("root task `%s:`", name)
			case flattenedOwner[name] != "":
				collidesWith = fmt.Sprintf("flattened include `%s:` (which also provides task `%s`)", flattenedOwner[name], name)
			}
			if collidesWith == "" {
				flattenedOwner[name] = e.namespace
				continue
			}
			anchor := e.nameNode
			findings = append(findings, Finding{
				RuleID:   "flatten-no-name-collision",
				Severity: SeverityError,
				Path:     c.reportPath,
				Line:     anchor.Line,
				Column:   anchor.Column,
				Message:  fmt.Sprintf("flattened include %q introduces task %q that collides with %s", e.namespace, name, collidesWith),
				Detail:   fmt.Sprintf("`flatten: true` merges included tasks without a namespace prefix, so every name must be unique."),
				Fix:      fmt.Sprintf("Rename the colliding task, drop `flatten: true` on `%s:`, or add `%s` to that include's `excludes:` list.", e.namespace, name),
				Docs:     includesDocsURL,
			})
		}
	}
	return findings
}

// collectIncludes walks the `includes:` mapping and returns a parsed
// entry per namespace, retaining yaml.Node pointers for location info.
func collectIncludes(root *yaml.Node) []includeEntry {
	_, includesNode := findRootKey(root, "includes")
	if includesNode == nil || includesNode.Kind != yaml.MappingNode {
		return nil
	}
	var out []includeEntry
	iterMapping(includesNode, func(key, value *yaml.Node) {
		entry := includeEntry{namespace: key.Value, nameNode: key}
		switch value.Kind {
		case yaml.ScalarNode:
			entry.taskfile = value.Value
			entry.pathNode = value
		case yaml.MappingNode:
			iterMapping(value, func(k, v *yaml.Node) {
				switch k.Value {
				case "taskfile":
					entry.taskfile = v.Value
					entry.pathNode = v
				case "optional":
					entry.optional = v.Value == "true"
				case "flatten":
					entry.flatten = v.Value == "true"
				case "excludes":
					if v.Kind == yaml.SequenceNode {
						for _, item := range v.Content {
							entry.excludes = append(entry.excludes, item.Value)
						}
					}
				}
			})
		default:
			// Unknown include shape — let the upstream AST catch it.
			return
		}
		out = append(out, entry)
	})
	return out
}

// isRemoteInclude reports whether the path uses a remote scheme that
// we intentionally don't validate.
func isRemoteInclude(p string) bool {
	for _, scheme := range remoteSchemes {
		if strings.HasPrefix(p, scheme) {
			return true
		}
	}
	return false
}

// extractTaskNames pulls just the task names from a taskfile's raw
// bytes, without requiring the file to be semantically valid.
func extractTaskNames(data []byte) ([]string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if len(root.Content) == 0 {
		return nil, nil
	}
	doc := root.Content[0]
	_, tasks := findRootKey(doc, "tasks")
	if tasks == nil || tasks.Kind != yaml.MappingNode {
		return nil, nil
	}
	var names []string
	for i := 0; i+1 < len(tasks.Content); i += 2 {
		names = append(names, tasks.Content[i].Value)
	}
	return names, nil
}
