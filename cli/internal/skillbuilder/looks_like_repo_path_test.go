package skillbuilder

import "testing"

// TestLooksLikeRepoPath verifies the heuristic used to decide whether a
// backticked token from SKILL.md should be checked for existence on disk.
// Regression guard for the package-path false-positive where tokens like
// `samber/lo` or `github.com/alecthomas/kong` were reported as missing
// references.
func TestLooksLikeRepoPath(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		// --- package-path-shaped tokens (should NOT be checked) ---
		{"org/name shorthand samber/lo", "samber/lo", false},
		{"org/name shorthand spf13/cobra", "spf13/cobra", false},
		{"org/name shorthand lmittmann/tint", "lmittmann/tint", false},
		{"stdlib log/slog", "log/slog", false},
		{"stdlib path/filepath", "path/filepath", false},
		{"module path github.com/x/y", "github.com/alecthomas/kong", false},
		{"module path gopkg.in/yaml.v3", "gopkg.in/yaml.v3", false},
		{"module path golang.org/x/tools", "golang.org/x/tools", false},
		{"URL", "https://example.com/foo", false},
		{"absolute path", "/etc/passwd", false},
		{"placeholder angles", "tools/<name>", false},
		{"trailing slash", "scripts/", false},
		{"flag fragment", "--verbose", false},
		{"single identifier lo", "lo", false},
		{"single identifier cobra", "cobra", false},
		{"contains space", "see references", false},
		{"empty", "", false},

		// --- repo-path-shaped tokens (SHOULD be checked) ---
		{"explicit relative", "./file.sh", true},
		{"explicit parent", "../other/README.md", true},
		{"references markdown", "references/cobra-patterns.md", true},
		{"references logging", "references/logging.md", true},
		{"scripts shell", "scripts/bootstrap.sh", true},
		{"assets yaml", "assets/config.yaml", true},
		{"known router filename", "SKILL.md", true},
		{"known readme", "README.md", true},
		{"known taskfile", "Taskfile.yml", true},
		{"dashed extension file", "references/kong-patterns.md", true},
		{"nested go file", "cli/internal/foo/bar.go", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := looksLikeRepoPath(tc.input)
			if got != tc.want {
				t.Fatalf("looksLikeRepoPath(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
