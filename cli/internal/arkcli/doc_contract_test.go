package arkcli

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

// TestSkillBuilderDocsMatchCLIContract guards that every `ark ...`
// command string mentioned in the skill-builder skill's docs resolves
// to a real kong-defined command. If you add or rename a subcommand,
// update the docs; if you drop a docs mention, drop the command — or
// rethink the contract.
func TestSkillBuilderDocsMatchCLIContract(t *testing.T) {
	skillDir := skillBuilderSkillDir(t)
	skillDoc := mustReadDoc(t, filepath.Join(skillDir, "SKILL.md"))
	cliRefDoc := mustReadDoc(t, filepath.Join(skillDir, "references", "repo-owned-clis.md"))

	for _, required := range []string{
		"`ark skill init`",
		"`ark skill audit`",
		"`ark skill sync`",
		"`ark skill check`",
		"`task ark:skill:init -- ...`",
		"`task ark:skill:audit -- ...`",
		"`task ark:skill:sync`",
		"`task ark:skill:check`",
	} {
		if !strings.Contains(skillDoc+"\n"+cliRefDoc, required) {
			t.Fatalf("skill-builder docs are missing required contract text %q", required)
		}
	}

	documented := documentedCommands(skillDoc)
	supported := kongCommands(t)

	var unsupported []string
	for command := range documented {
		if !supported[normalizeDocumentedCommand(command)] {
			unsupported = append(unsupported, command)
		}
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		t.Fatalf("skill-builder docs mention unsupported commands: %s", strings.Join(unsupported, ", "))
	}
}

// kongCommands walks the kong model tree and returns the set of
// commands expressible as space-joined paths (e.g. "ark skill init").
func kongCommands(t *testing.T) map[string]bool {
	t.Helper()

	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name(binaryName),
		kong.Vars{
			"version": appVersion,
		},
	)
	if err != nil {
		t.Fatalf("build kong parser: %v", err)
	}

	commands := make(map[string]bool)
	var walk func(prefix []string, nodes []*kong.Node)
	walk = func(prefix []string, nodes []*kong.Node) {
		for _, child := range nodes {
			path := append(append([]string{}, prefix...), child.Name)
			commands[strings.Join(path, " ")] = true
			walk(path, child.Children)
		}
	}
	root := []string{binaryName}
	commands[binaryName] = true
	walk(root, parser.Model.Node.Children)
	return commands
}

func skillBuilderSkillDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve runtime caller")
	}
	// filename lives at cli/internal/arkcli/doc_contract_test.go; skill dir is at
	// <repo>/skills/skill-builder.
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "skills", "skill-builder"))
}

func mustReadDoc(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read doc %s: %v", path, err)
	}
	return string(data)
}

func documentedCommands(content string) map[string]bool {
	pattern := regexp.MustCompile("`(ark[^`]+)`")
	commands := make(map[string]bool)
	for _, match := range pattern.FindAllStringSubmatch(content, -1) {
		commands[strings.Join(strings.Fields(match[1]), " ")] = true
	}
	return commands
}

func normalizeDocumentedCommand(command string) string {
	tokens := strings.Fields(command)
	var normalized []string
	for _, token := range tokens {
		if strings.HasPrefix(token, "--") || token == "..." {
			break
		}
		normalized = append(normalized, token)
	}
	return strings.Join(normalized, " ")
}
