package tasklint

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	ignore "github.com/sabhiram/go-gitignore"
)

// gitignoreMatcher lazily loads <repoRoot>/.gitignore and answers
// "is this path ignored?" queries. It deliberately swallows "no
// .gitignore present" so callers can treat it as a simple predicate
// while still letting the rule report the missing file.
type gitignoreMatcher struct {
	repoRoot string

	once    sync.Once
	matcher *ignore.GitIgnore
	exists  bool
	lines   []string // raw lines, used for explicit substring checks
}

func newGitignoreMatcher(repoRoot string) *gitignoreMatcher {
	return &gitignoreMatcher{repoRoot: repoRoot}
}

// load reads and compiles the .gitignore. Safe to call many times.
func (g *gitignoreMatcher) load() {
	g.once.Do(func() {
		path := filepath.Join(g.repoRoot, ".gitignore")
		raw, err := os.ReadFile(path)
		if err != nil {
			g.exists = false
			g.matcher = ignore.CompileIgnoreLines()
			return
		}
		g.exists = true
		g.lines = strings.Split(string(raw), "\n")
		g.matcher = ignore.CompileIgnoreLines(g.lines...)
	})
}

// Exists reports whether a .gitignore file was found at the repo root.
func (g *gitignoreMatcher) Exists() bool {
	g.load()
	return g.exists
}

// Matches reports whether the given path (repo-relative or absolute
// within the repo) is covered by the compiled .gitignore.
func (g *gitignoreMatcher) Matches(p string) bool {
	g.load()
	if g.matcher == nil {
		return false
	}
	return g.matcher.MatchesPath(g.relativize(p))
}

// HasLiteralPattern reports whether .gitignore contains a line that
// (after trimming comments/whitespace) equals one of the given
// candidates. This is used for the fingerprint rule, where we want to
// accept any of `.task`, `.task/`, or `/.task` as "covered" without
// worrying about the matcher's edge cases around trailing slashes.
func (g *gitignoreMatcher) HasLiteralPattern(candidates ...string) bool {
	g.load()
	if !g.exists {
		return false
	}
	for _, raw := range g.lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, c := range candidates {
			if line == c {
				return true
			}
		}
	}
	return false
}

// relativize turns an absolute path inside the repo into a
// repo-relative path (preferred by the matcher). Other paths are
// returned unchanged.
func (g *gitignoreMatcher) relativize(p string) string {
	if g.repoRoot == "" || !filepath.IsAbs(p) {
		return p
	}
	rel, err := filepath.Rel(g.repoRoot, p)
	if err != nil {
		return p
	}
	return rel
}
