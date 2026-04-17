package tasklint

import (
	astpkg "github.com/go-task/task/v3/taskfile/ast"
	"gopkg.in/yaml.v3"
)

// ruleContext bundles everything a rule function needs. Constructed
// once per Lint call; rules read from it only.
type ruleContext struct {
	// path is the taskfile path as supplied by the caller (used for
	// include resolution).
	path string
	// reportPath is what we put into Finding.Path — usually repo-relative.
	reportPath string
	// repoRoot is the directory used for .gitignore lookup and
	// include-path resolution when the include path is relative.
	repoRoot string

	astFile     *astpkg.Taskfile // may be nil if upstream decode failed
	astParseErr error            // set if upstream parse failed

	// rootNode is the top-level YAML MappingNode.
	rootNode *yaml.Node
	// documentNode is the enclosing Document (rarely needed).
	documentNode *yaml.Node

	// cached helpers — lazily populated by accessors.
	_ignore *gitignoreMatcher
}

// gitignore returns the (cached) matcher for this context's repo root.
func (c *ruleContext) gitignore() *gitignoreMatcher {
	if c._ignore == nil {
		c._ignore = newGitignoreMatcher(c.repoRoot)
	}
	return c._ignore
}

// ruleFunc is the contract every per-rule implementation satisfies.
type ruleFunc func(c *ruleContext) []Finding

// ruleFuncs lists the rules in execution order. Keep in sync with
// rules_data.go and the per-rule tests.
var ruleFuncs = []ruleFunc{
	ruleSchemaError,             // 0 — upstream AST decode error surfacing
	ruleVersionRequired,         // 1
	ruleVersionIsThree,          // 2
	ruleUnknownTopLevelKeys,     // 3
	ruleUnknownTaskKeys,         // 4
	ruleCmdAndCmdsMutex,         // 5
	ruleIncludesPathsResolvable, // 6
	ruleFlattenNoNameCollision,  // 7
	ruleMethodValidEnum,         // 8
	ruleFingerprintDirGitignored, // 9
	ruleDotenvFilesGitignored,    // 10
}

// ruleSchemaError surfaces upstream AST decode errors (captured by
// the parser) as a single `schema-error` finding. Other rules keep
// running so the user sees every problem at once.
func ruleSchemaError(c *ruleContext) []Finding {
	if c.astParseErr == nil {
		return nil
	}
	return []Finding{schemaErrorFinding(c.reportPath, c.astParseErr)}
}

// findRootKey returns the value yaml.Node for a top-level key, or nil.
func findRootKey(root *yaml.Node, key string) (keyNode, valueNode *yaml.Node) {
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i]
		if k.Value == key {
			return k, root.Content[i+1]
		}
	}
	return nil, nil
}

// iterMapping walks a MappingNode yielding each (key, value) pair.
// Safe to call on nil or non-mapping nodes (yields nothing).
func iterMapping(n *yaml.Node, fn func(key, value *yaml.Node)) {
	if n == nil || n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		fn(n.Content[i], n.Content[i+1])
	}
}
