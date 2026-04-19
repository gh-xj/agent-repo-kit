package skillbuilder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const (
	skillLineWarnThreshold  = 200
	skillLineErrorThreshold = 400
)

var (
	frontmatterNamePattern        = regexp.MustCompile(`(?m)^name:\s*"?([^"\n]+)"?\s*$`)
	frontmatterDescriptionPattern = regexp.MustCompile(`(?m)^description:\s*"?([^"\n]+)"?\s*$`)
	inlinePathPattern             = regexp.MustCompile("`([^`]+)`")
	linkPathPattern               = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	// modulePathPrefixPattern matches classic module path hosts like
	// `github.com/`, `gopkg.in/`, `golang.org/x/`, `example.co.uk/`.
	modulePathPrefixPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-.][a-z0-9]+)*\.[a-z]{2,}/`)
	// goIdentifierPattern matches a token that is a valid Go identifier
	// (letters, digits, underscore; not starting with a digit).
	goIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// knownSkillPathPrefixes are directory names commonly used inside a skill
// directory. A backticked token beginning with one of these is treated as a
// relative repo path, not a package shorthand.
var knownSkillPathPrefixes = []string{
	"references/",
	"scripts/",
	"assets/",
	"cli/",
	"tools/",
	"templates/",
	"examples/",
	"fixtures/",
}

// repoPathExtensions are file-name suffixes that strongly suggest a repo
// path rather than a package path.
var repoPathExtensions = []string{
	".md", ".go", ".sh", ".yml", ".yaml", ".json",
	".tmpl", ".txt", ".toml", ".py", ".rs", ".ts", ".js",
}

type CLIInitializer func(skillDir string, module string) error

type InitConfig struct {
	SkillDir       string
	Name           string
	Description    string
	WithCLI        bool
	CLIModule      string
	CLIInitializer CLIInitializer
}

type InitResult struct {
	SkillDir  string   `json:"skill_dir"`
	SkillPath string   `json:"skill_path"`
	CLIPath   string   `json:"cli_path,omitempty"`
	Created   []string `json:"created"`
}

type AuditConfig struct {
	SkillDir string
}

type AuditResult struct {
	SkillDir    string    `json:"skill_dir"`
	SkillPath   string    `json:"skill_path"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	LineCount   int       `json:"line_count"`
	HasCLI      bool      `json:"has_cli"`
	Referenced  []string  `json:"referenced_files,omitempty"`
	Findings    []Finding `json:"findings,omitempty"`
}

type Finding struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

func (r AuditResult) HasErrors() bool {
	for _, finding := range r.Findings {
		if finding.Level == "error" {
			return true
		}
	}
	return false
}

func InitSkill(cfg InitConfig) (InitResult, error) {
	cfg.SkillDir = strings.TrimSpace(cfg.SkillDir)
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.Description = strings.TrimSpace(cfg.Description)

	if cfg.SkillDir == "" {
		return InitResult{}, fmt.Errorf("skill dir is required")
	}
	if cfg.Name == "" {
		return InitResult{}, fmt.Errorf("name is required")
	}
	if cfg.Description == "" {
		return InitResult{}, fmt.Errorf("description is required")
	}
	if cfg.WithCLI && strings.TrimSpace(cfg.CLIModule) == "" {
		return InitResult{}, fmt.Errorf("cli module is required when --with-cli is set")
	}

	skillDir, err := filepath.Abs(cfg.SkillDir)
	if err != nil {
		return InitResult{}, fmt.Errorf("resolve skill dir: %w", err)
	}
	if err := ensureEmptyOrMissingDir(skillDir); err != nil {
		return InitResult{}, err
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return InitResult{}, fmt.Errorf("create skill dir: %w", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(renderSkillTemplate(cfg.Name, cfg.Description)), 0o644); err != nil {
		return InitResult{}, fmt.Errorf("write SKILL.md: %w", err)
	}

	result := InitResult{
		SkillDir:  skillDir,
		SkillPath: skillPath,
		Created:   []string{skillPath},
	}

	if cfg.WithCLI {
		if cfg.CLIInitializer == nil {
			return InitResult{}, fmt.Errorf("cli initializer is required when --with-cli is set; provide cfg.CLIInitializer")
		}
		if err := cfg.CLIInitializer(skillDir, strings.TrimSpace(cfg.CLIModule)); err != nil {
			return InitResult{}, err
		}
		result.CLIPath = filepath.Join(skillDir, "cli")
		result.Created = append(result.Created, result.CLIPath)
	}

	return result, nil
}

func AuditSkill(cfg AuditConfig) (AuditResult, error) {
	trimmed := strings.TrimSpace(cfg.SkillDir)
	if trimmed == "" {
		return AuditResult{}, fmt.Errorf("skill dir is required")
	}
	skillDir, err := filepath.Abs(trimmed)
	if err != nil {
		return AuditResult{}, fmt.Errorf("resolve skill dir: %w", err)
	}

	result := AuditResult{
		SkillDir:  skillDir,
		SkillPath: filepath.Join(skillDir, "SKILL.md"),
		HasCLI:    dirExists(filepath.Join(skillDir, "cli")),
	}

	data, err := os.ReadFile(result.SkillPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Findings = append(result.Findings, Finding{
				Level:   "error",
				Code:    "missing_skill",
				Path:    result.SkillPath,
				Message: "SKILL.md is required",
			})
			return result, nil
		}
		return result, fmt.Errorf("read SKILL.md: %w", err)
	}

	content := string(data)
	result.LineCount = countLines(content)
	result.Name = extractMatch(frontmatterNamePattern, content)
	result.Description = extractMatch(frontmatterDescriptionPattern, content)

	if result.Name == "" {
		result.Findings = append(result.Findings, Finding{
			Level:   "error",
			Code:    "missing_name",
			Path:    result.SkillPath,
			Message: "frontmatter must include name",
		})
	}
	if result.Description == "" {
		result.Findings = append(result.Findings, Finding{
			Level:   "error",
			Code:    "missing_description",
			Path:    result.SkillPath,
			Message: "frontmatter must include description",
		})
	}
	if result.Description != "" && !triggerOrientedDescription(result.Description) {
		result.Findings = append(result.Findings, Finding{
			Level:   "warning",
			Code:    "description_not_trigger_oriented",
			Path:    result.SkillPath,
			Message: "description should say when to use the skill",
		})
	}
	if result.LineCount > skillLineWarnThreshold {
		level := "warning"
		code := "router_should_extract"
		message := fmt.Sprintf("SKILL.md is %d lines; consider extracting references", result.LineCount)
		if result.LineCount > skillLineErrorThreshold {
			level = "error"
			code = "router_too_large"
			message = fmt.Sprintf("SKILL.md is %d lines; refactor the router now", result.LineCount)
		}
		result.Findings = append(result.Findings, Finding{
			Level:   level,
			Code:    code,
			Path:    result.SkillPath,
			Message: message,
		})
	}

	result.Referenced = referencedRelativePaths(content)
	for _, rel := range result.Referenced {
		target := filepath.Join(skillDir, filepath.Clean(rel))
		if !fileOrDirExists(target) {
			result.Findings = append(result.Findings, Finding{
				Level:   "error",
				Code:    "missing_reference",
				Path:    target,
				Message: fmt.Sprintf("referenced path %q does not exist", rel),
			})
		}
	}

	return result, nil
}

func renderSkillTemplate(name, description string) string {
	title := titleFromSlug(name)
	return fmt.Sprintf(`---
name: %s
description: %q
---

# %s

## Use This For

- Describe the requests that should trigger this skill.

## Router

- State the first decision the agent should make before following deeper references or code.

## References

- Add only the references, scripts, assets, or CLI surfaces this skill actually needs.
`, name, description, title)
}

func ensureEmptyOrMissingDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read skill dir: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("skill dir already exists and is not empty: %s", path)
	}
	return nil
}

func referencedRelativePaths(content string) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, pattern := range []*regexp.Regexp{inlinePathPattern, linkPathPattern} {
		for _, match := range pattern.FindAllStringSubmatch(content, -1) {
			candidate := strings.TrimSpace(match[1])
			if !looksLikeRepoPath(candidate) {
				continue
			}
			candidate = filepath.Clean(candidate)
			if !seen[candidate] {
				seen[candidate] = true
				refs = append(refs, candidate)
			}
		}
	}

	slices.Sort(refs)
	return refs
}

// looksLikeRepoPath reports whether a backticked token from SKILL.md should
// be checked for existence on disk as a path relative to the skill directory.
//
// The extractor skips tokens that are obviously not paths (URLs, placeholder
// markers, command fragments, package shorthands like `samber/lo` or
// `github.com/x/y`). For ambiguous single-segment tokens, only a small set
// of well-known filenames are treated as paths; everything else is skipped
// to keep the signal-to-noise ratio high. When a token with a `/` could be
// either a relative path or a package path, we prefer known skill
// directories and known file extensions as evidence it is a path; absent
// that evidence, we fall back to a conservative default of checking.
func looksLikeRepoPath(candidate string) bool {
	// 0. Structural filters: empty, absolute, URL-like, whitespace, HTML
	//    placeholder markers, trailing slash, or CLI flag fragments.
	if candidate == "" || strings.Contains(candidate, "://") || filepath.IsAbs(candidate) {
		return false
	}
	if strings.ContainsAny(candidate, " \t\n") {
		return false
	}
	if strings.ContainsAny(candidate, "<>") {
		return false
	}
	if strings.HasSuffix(candidate, "/") {
		return false
	}
	if strings.HasPrefix(candidate, "--") {
		return false
	}

	// 1. Explicit relative or absolute path markers: always check.
	if strings.HasPrefix(candidate, "./") || strings.HasPrefix(candidate, "../") {
		return true
	}

	// 2. Single-segment tokens (no `/`): only treat known project filenames
	//    as references; ignore everything else.
	if !strings.Contains(candidate, "/") {
		switch candidate {
		case "SKILL.md", "README.md", "Taskfile.yml", "AGENTS.md", "CLAUDE.md":
			return true
		default:
			return false
		}
	}

	// 3. Known skill subdirectory prefixes: `references/...`, `scripts/...`,
	//    etc. Treat these as repo paths even if both segments happen to look
	//    like Go identifiers.
	for _, prefix := range knownSkillPathPrefixes {
		if strings.HasPrefix(candidate, prefix) {
			return true
		}
	}

	// 4. Classic module path hosts (e.g. `github.com/x/y`, `gopkg.in/x`,
	//    `golang.org/x/tools`): skip.
	if modulePathPrefixPattern.MatchString(candidate) {
		return false
	}

	// 5. `org/name`-shaped shorthand (exactly one `/`, both sides are valid
	//    Go identifiers): skip. Catches `samber/lo`, `spf13/cobra`,
	//    `log/slog`.
	if parts := strings.Split(candidate, "/"); len(parts) == 2 &&
		goIdentifierPattern.MatchString(parts[0]) &&
		goIdentifierPattern.MatchString(parts[1]) {
		return false
	}

	// 6. Known file extensions: treat as a repo path.
	lower := strings.ToLower(candidate)
	for _, ext := range repoPathExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}

	// 7. Conservative default: treat multi-segment tokens as repo paths so
	//    genuine references surface as missing_reference findings when the
	//    author forgets to create the file. Package paths without a host
	//    prefix and outside the `org/name` shape are rare in practice.
	return true
}

func extractMatch(pattern *regexp.Regexp, content string) string {
	match := pattern.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func triggerOrientedDescription(description string) bool {
	lowered := strings.ToLower(description)
	return strings.Contains(lowered, "use when") || strings.Contains(lowered, "use for") || strings.Contains(lowered, "trigger")
}

func countLines(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func titleFromSlug(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}
