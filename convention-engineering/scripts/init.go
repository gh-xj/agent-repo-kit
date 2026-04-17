package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	initManagedMarker        = "agent-repo-kit --init"
	initManagedBlockStart    = "<!-- agent-repo-kit:init:start -->"
	initManagedBlockEnd      = "<!-- agent-repo-kit:init:end -->"
	initConventionsTaskfile  = ".convention-engineering/Taskfile.yml"
	initConventionsCheckPath = ".convention-engineering/check.sh"
	initConfigPath           = ".convention-engineering.json"
)

type initOptions struct {
	Profiles   []string
	Operations []string
	RepoRisk   string
}

func runInit(repoRoot string, opts initOptions, stdout, stderr io.Writer) int {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve repo root: %v\n", err)
		return 2
	}

	normalized, err := normalizeInitOptions(root, opts)
	if err != nil {
		fmt.Fprintf(stderr, "failed to prepare init options: %v\n", err)
		return 2
	}

	repoName := filepath.Base(root)
	if err := scaffoldTrackedRepo(root, repoName, normalized); err != nil {
		fmt.Fprintf(stderr, "failed to scaffold repo conventions: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "initialized tracked conventions for %s\n", repoName)
	return run(root, initConfigPath, false, stdout, stderr)
}

func normalizeInitOptions(root string, opts initOptions) (initOptions, error) {
	normalized := initOptions{
		Profiles:   normalizeCSVItems(opts.Profiles),
		Operations: normalizeCSVItems(opts.Operations),
		RepoRisk:   strings.TrimSpace(opts.RepoRisk),
	}
	if normalized.RepoRisk == "" {
		normalized.RepoRisk = "standard"
	}
	if len(normalized.Operations) == 0 {
		normalized.Operations = []string{"tickets", "wiki"}
	}
	if len(normalized.Profiles) == 0 {
		detected, err := detectProfiles(root)
		if err != nil {
			return initOptions{}, err
		}
		normalized.Profiles = detected
	}
	if len(normalized.Profiles) == 0 {
		return initOptions{}, fmt.Errorf("could not detect repo profiles; pass --profiles")
	}

	allowedOps := map[string]bool{"tickets": true, "wiki": true}
	for _, op := range normalized.Operations {
		if !allowedOps[op] {
			return initOptions{}, fmt.Errorf("unsupported operation %q (allowed: tickets,wiki)", op)
		}
	}

	seenProfile := map[string]bool{}
	profiles := make([]string, 0, len(normalized.Profiles))
	for _, profile := range normalized.Profiles {
		if profile == "" || seenProfile[profile] {
			continue
		}
		seenProfile[profile] = true
		profiles = append(profiles, profile)
	}
	normalized.Profiles = profiles

	seenOp := map[string]bool{}
	ops := make([]string, 0, len(normalized.Operations))
	for _, op := range normalized.Operations {
		if op == "" || seenOp[op] {
			continue
		}
		seenOp[op] = true
		ops = append(ops, op)
	}
	sort.Strings(ops)
	normalized.Operations = ops
	return normalized, nil
}

func normalizeCSVItems(items []string) []string {
	normalized := make([]string, 0)
	for _, item := range items {
		for _, part := range strings.Split(item, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" || strings.EqualFold(trimmed, "none") {
				continue
			}
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func detectProfiles(root string) ([]string, error) {
	type detector struct {
		name       string
		files      []string
		extensions []string
	}

	detectors := []detector{
		{name: "go", files: []string{"go.mod"}, extensions: []string{".go"}},
		{name: "typescript-react", files: []string{"package.json", "tsconfig.json"}, extensions: []string{".ts", ".tsx", ".js", ".jsx"}},
		{name: "python", files: []string{"pyproject.toml", "requirements.txt"}, extensions: []string{".py"}},
	}

	found := make([]string, 0, len(detectors))
	for _, detector := range detectors {
		matched, err := repoMatchesDetector(root, detector)
		if err != nil {
			return nil, err
		}
		if matched {
			found = append(found, detector.name)
		}
	}
	return found, nil
}

func repoMatchesDetector(root string, detector struct {
	name       string
	files      []string
	extensions []string
}) (bool, error) {
	for _, rel := range detector.files {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			return true, nil
		}
	}

	skipDirs := map[string]bool{
		".git":                    true,
		".tickets":                true,
		".wiki":                   true,
		".convention-engineering": true,
		"node_modules":            true,
		"vendor":                  true,
	}

	matched := false
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if skipDirs[d.Name()] && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		for _, ext := range detector.extensions {
			if strings.EqualFold(filepath.Ext(d.Name()), ext) {
				matched = true
				return io.EOF
			}
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return false, err
	}
	return matched, nil
}

func scaffoldTrackedRepo(root, repoName string, opts initOptions) error {
	sourceRoot, err := resolveConventionEngineeringRoot()
	if err != nil {
		return err
	}

	configPath := filepath.Join(root, initConfigPath)
	configBytes, err := buildInitConfig(repoName, opts)
	if err != nil {
		return err
	}
	if err := writeManagedJSON(configPath, configBytes); err != nil {
		return err
	}

	if err := ensureDocsReadmes(root, repoName, opts); err != nil {
		return err
	}
	if err := ensureTickets(root, repoName, hasOperation(opts.Operations, "tickets")); err != nil {
		return err
	}
	if err := ensureWiki(root, hasOperation(opts.Operations, "wiki")); err != nil {
		return err
	}
	if err := ensureConventionSupportFiles(root, sourceRoot, opts); err != nil {
		return err
	}
	if err := ensureRootTaskfile(root); err != nil {
		return err
	}
	if err := ensureAgentContractFiles(root, opts); err != nil {
		return err
	}
	return nil
}

func hasOperation(operations []string, target string) bool {
	for _, op := range operations {
		if op == target {
			return true
		}
	}
	return false
}

func buildInitConfig(repoName string, opts initOptions) ([]byte, error) {
	requiredFiles := []string{
		"AGENTS.md",
		"CLAUDE.md",
		"Taskfile.yml",
		initConventionsTaskfile,
		initConventionsCheckPath,
		"docs/README.md",
		"docs/requests/README.md",
		"docs/planning/README.md",
		"docs/plans/README.md",
		"docs/implementation/README.md",
		"docs/taxonomy/README.md",
	}
	if hasOperation(opts.Operations, "tickets") {
		requiredFiles = append(requiredFiles,
			".tickets/.gitignore",
			".tickets/README.md",
			".tickets/Taskfile.yml",
			".tickets/all/.gitkeep",
			".tickets/audit-log.md",
			".tickets/harness/schema.yaml",
			".tickets/harness/taxonomy.yaml",
			".tickets/harness/test-ticket-system.sh",
		)
	}
	if hasOperation(opts.Operations, "wiki") {
		requiredFiles = append(requiredFiles,
			".wiki/RULES.md",
			".wiki/Taskfile.yml",
			".wiki/pages/.gitkeep",
			".wiki/raw/.gitkeep",
			".wiki/scripts/lint.sh",
		)
	}

	taskfileChecks := map[string]any{
		"Taskfile.yml": []string{
			initConventionsTaskfile,
			"flatten: true",
		},
		initConventionsTaskfile: []string{
			"check:conventions:",
			"verify:",
			"check.sh",
		},
	}
	if hasOperation(opts.Operations, "tickets") {
		taskfileChecks[initConventionsTaskfile] = append(taskfileChecks[initConventionsTaskfile].([]string), "../.tickets/Taskfile.yml")
		taskfileChecks[".tickets/Taskfile.yml"] = []string{"init:", "new:", "transition:", "close:", "test:"}
	}
	if hasOperation(opts.Operations, "wiki") {
		taskfileChecks[initConventionsTaskfile] = append(taskfileChecks[initConventionsTaskfile].([]string), "../.wiki/Taskfile.yml")
		taskfileChecks[".wiki/Taskfile.yml"] = []string{"init:", "lint:"}
	}

	conventionMarkers := []string{
		"## Conventions",
		"**Docs**",
		initConfigPath,
		"`task verify`",
	}
	if hasOperation(opts.Operations, "tickets") {
		conventionMarkers = append(conventionMarkers, ".tickets/README.md")
	}
	if hasOperation(opts.Operations, "wiki") {
		conventionMarkers = append(conventionMarkers, ".wiki/RULES.md")
	}

	contentChecks := []map[string]any{
		{
			"name":             "agents-conventions",
			"file":             "AGENTS.md",
			"required_markers": conventionMarkers,
		},
		{
			"name":             "claude-conventions",
			"file":             "CLAUDE.md",
			"required_markers": conventionMarkers,
		},
		{
			"name": "docs-root-explained",
			"file": "docs/README.md",
			"required_markers": []string{
				initConfigPath,
				"requests/",
				"planning/",
				"plans/",
				"implementation/",
				"taxonomy/",
			},
		},
	}
	if hasOperation(opts.Operations, "tickets") {
		contentChecks = append(contentChecks, map[string]any{
			"name": "tickets-taxonomy-customized",
			"file": ".tickets/harness/taxonomy.yaml",
			"required_markers": []string{
				fmt.Sprintf("project: %s", repoName),
			},
		})
	}

	config := map[string]any{
		"generated_by":     initManagedMarker,
		"contract_version": 1,
		"mode":             "tracked",
		"profiles":         opts.Profiles,
		"docs_root":        "docs",
		"ownership_policy": map[string]any{
			"portable_skill_authoring_owner": "repo-maintainers",
			"domain_knowledge_owner":         "repo-maintainers",
			"repo_local_skills": map[string]any{
				"allowed":                false,
				"placement_roots":        []string{".claude/skills", ".agents/skills"},
				"authoring_owner":        "repo-maintainers",
				"requires_justification": true,
			},
		},
		"required_files":  requiredFiles,
		"taskfile_checks": taskfileChecks,
		"mirror_policy": map[string]any{
			"mode":  "mirrored",
			"files": []string{"CLAUDE.md", "AGENTS.md"},
		},
		"canonical_pointer_mode": "all",
		"canonical_pointers":     []any{},
		"content_checks":         contentChecks,
		"git_exclude_checks":     []any{},
		"invariant_contract": map[string]any{
			"required": false,
		},
		"evaluation_inputs": map[string]any{
			"repo_risk": opts.RepoRisk,
		},
		"chunk_plan": map[string]any{
			"enabled": false,
			"chunks":  []any{},
		},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(config); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeManagedJSON(path string, content []byte) error {
	if isManagedJSON(path) {
		return os.WriteFile(path, content, 0o644)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("refusing to overwrite unmanaged file %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func isManagedJSON(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	return strings.TrimSpace(fmt.Sprint(payload["generated_by"])) == initManagedMarker
}

func ensureDocsReadmes(root, repoName string, opts initOptions) error {
	docs := map[string]string{
		"docs/README.md": fmt.Sprintf(`# Docs

Tracked docs root for %q. This repo uses the intent-first taxonomy declared in %q.

- %q — RFIs, asks, scope, and acceptance criteria.
- %q — design decisions and architecture choices.
- %q — executable implementation plans.
- %q — rollout notes, verification notes, and implementation reports.
- %q — stable vocabularies, schemas, and mappings.

External-source-backed material belongs in %q, not %q.

<!-- Generated by %s -->
`, repoName, initConfigPath, "requests/", "planning/", "plans/", "implementation/", "taxonomy/", ".wiki/", "docs/", initManagedMarker),
		"docs/requests/README.md": `# Requests

Store RFI and scoped-request documents here.

Filename contract: ` + "`YYYYMMDD_rfi_<topic>.md`" + `

<!-- Generated by ` + initManagedMarker + ` -->
`,
		"docs/planning/README.md": `# Planning

Store approved design and architecture documents here.

Filename contract: ` + "`YYYY-MM-DD_<topic>_design.md`" + `

<!-- Generated by ` + initManagedMarker + ` -->
`,
		"docs/plans/README.md": `# Plans

Store executable implementation plans here.

Filename contract: ` + "`YYYY-MM-DD-<topic>.md`" + `

<!-- Generated by ` + initManagedMarker + ` -->
`,
		"docs/implementation/README.md": `# Implementation

Store rollout logs, verification notes, and implementation reports here.

Filename contract: ` + "`YYYY-MM-DD_<topic>_impl_report.md`" + `

<!-- Generated by ` + initManagedMarker + ` -->
`,
		"docs/taxonomy/README.md": `# Taxonomy

Store stable domain taxonomies, schemas, and mapping notes here.

Recommended shape: ` + "`taxonomy/<domain>/<subject>.md`" + `

<!-- Generated by ` + initManagedMarker + ` -->
`,
	}

	if !hasOperation(opts.Operations, "wiki") {
		docs["docs/README.md"] = fmt.Sprintf(`# Docs

Tracked docs root for %q. This repo uses the intent-first taxonomy declared in %q.

- %q — RFIs, asks, scope, and acceptance criteria.
- %q — design decisions and architecture choices.
- %q — executable implementation plans.
- %q — rollout notes, verification notes, and implementation reports.
- %q — stable vocabularies, schemas, and mappings.

Use %q for repo-authored documentation.

<!-- Generated by %s -->
`, repoName, initConfigPath, "requests/", "planning/", "plans/", "implementation/", "taxonomy/", "docs/", initManagedMarker)
	}

	for rel, body := range docs {
		if err := writeManagedTextFile(filepath.Join(root, rel), body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func ensureTickets(root, repoName string, enabled bool) error {
	if !enabled {
		return nil
	}
	srcRoot, err := resolveTemplatesRoot()
	if err != nil {
		return err
	}
	src := filepath.Join(srcRoot, "tickets")
	dst := filepath.Join(root, ".tickets")
	if err := copyTemplateTree(src, dst, map[string]func([]byte) []byte{
		"harness/taxonomy.yaml": func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "<repo-name>", repoName))
		},
	}); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dst, "all"), 0o755); err != nil {
		return err
	}
	if err := touchManagedFile(filepath.Join(dst, "all", ".gitkeep"), 0o644); err != nil {
		return err
	}
	if err := writeMissingFile(filepath.Join(dst, "audit-log.md"), []byte("# Audit Log\n\n"), 0o644); err != nil {
		return err
	}
	return nil
}

func ensureWiki(root string, enabled bool) error {
	if !enabled {
		return nil
	}
	srcRoot, err := resolveTemplatesRoot()
	if err != nil {
		return err
	}
	src := filepath.Join(srcRoot, "wiki")
	dst := filepath.Join(root, ".wiki")
	if err := copyTemplateTree(src, dst, nil); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dst, "raw"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dst, "pages"), 0o755); err != nil {
		return err
	}
	if err := touchManagedFile(filepath.Join(dst, "raw", ".gitkeep"), 0o644); err != nil {
		return err
	}
	if err := touchManagedFile(filepath.Join(dst, "pages", ".gitkeep"), 0o644); err != nil {
		return err
	}
	return nil
}

func resolveConventionEngineeringRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve convention-engineering source root")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	if _, err := os.Stat(filepath.Join(root, "scripts", "main.go")); err != nil {
		return "", err
	}
	return root, nil
}

func resolveTemplatesRoot() (string, error) {
	root, err := resolveConventionEngineeringRoot()
	if err != nil {
		return "", err
	}
	templatesRoot := filepath.Join(root, "references", "templates")
	if _, err := os.Stat(templatesRoot); err != nil {
		return "", err
	}
	return templatesRoot, nil
}

func copyTemplateTree(src, dst string, transforms map[string]func([]byte) []byte) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if _, err := os.Stat(target); err == nil {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if transform := transforms[filepath.ToSlash(rel)]; transform != nil {
			content = transform(content)
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

func touchManagedFile(path string, mode fs.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, nil, mode)
}

func writeMissingFile(path string, content []byte, mode fs.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, mode)
}

func writeManagedTextFile(path, content string, mode fs.FileMode) error {
	if data, err := os.ReadFile(path); err == nil && !strings.Contains(string(data), initManagedMarker) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), mode)
}

func ensureConventionSupportFiles(root, sourceRoot string, opts initOptions) error {
	if err := os.MkdirAll(filepath.Join(root, ".convention-engineering"), 0o755); err != nil {
		return err
	}
	if err := writeManagedTextFile(filepath.Join(root, initConventionsCheckPath), renderConventionCheckScript(sourceRoot), 0o755); err != nil {
		return err
	}
	if err := writeManagedTextFile(filepath.Join(root, initConventionsTaskfile), renderConventionTaskfile(opts), 0o644); err != nil {
		return err
	}
	return nil
}

func renderConventionCheckScript(sourceRoot string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

bootstrap_source=%s

resolve_tool_dir() {
  local candidates=()
  if [ -n "${CONVENTION_ENGINEERING_DIR:-}" ]; then
    candidates+=("$CONVENTION_ENGINEERING_DIR")
  fi
  if [ -n "$bootstrap_source" ]; then
    candidates+=("$bootstrap_source")
  fi
  if [ -n "${HOME:-}" ]; then
    candidates+=("$HOME/.claude/skills/convention-engineering")
    candidates+=("$HOME/.agents/skills/convention-engineering")
    candidates+=("$HOME/.codex/skills/convention-engineering")
  fi

  local candidate
  for candidate in "${candidates[@]}"; do
    if [ -f "$candidate/scripts/main.go" ]; then
      printf '%%s\n' "$candidate/scripts"
      return 0
    fi
    if [ -f "$candidate/main.go" ]; then
      printf '%%s\n' "$candidate"
      return 0
    fi
  done

  printf 'convention-engineering tool not found. Set CONVENTION_ENGINEERING_DIR, keep the original bootstrap checkout available, or install agent-repo-kit into ~/.claude/skills or ~/.agents/skills (legacy fallback: ~/.codex/skills).\n' >&2
  return 1
}

tool_dir=$(resolve_tool_dir)
go_files=()
for file in "$tool_dir"/*.go; do
  case "$file" in
    *_test.go) continue ;;
  esac
  go_files+=("$file")
done

exec env GO111MODULE=off go run "${go_files[@]}" "$@"

# Generated by %s
`, shellSingleQuote(sourceRoot), initManagedMarker)
}

func shellSingleQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func renderConventionTaskfile(opts initOptions) string {
	lines := []string{
		`version: "3"`,
		"",
	}

	if hasOperation(opts.Operations, "tickets") || hasOperation(opts.Operations, "wiki") {
		lines = append(lines, "includes:")
		if hasOperation(opts.Operations, "tickets") {
			lines = append(lines,
				"  tickets:",
				"    taskfile: ../.tickets/Taskfile.yml",
				"    dir: ../.tickets",
			)
		}
		if hasOperation(opts.Operations, "wiki") {
			lines = append(lines,
				"  wiki:",
				"    taskfile: ../.wiki/Taskfile.yml",
				"    dir: ../.wiki",
			)
		}
		lines = append(lines, "")
	}

	lines = append(lines,
		"tasks:",
		"  check:conventions:",
		"    desc: Validate the tracked convention contract",
		"    cmds:",
		"      - bash check.sh --repo-root .. --config "+initConfigPath,
		"",
		"  verify:",
		"    desc: Run the repo convention verification gates",
		"    cmds:",
	)
	if hasOperation(opts.Operations, "tickets") {
		lines = append(lines, "      - task: tickets:test")
	}
	if hasOperation(opts.Operations, "wiki") {
		lines = append(lines, "      - task: wiki:lint")
	}
	lines = append(lines,
		"      - task: check:conventions",
		"",
		"# Generated by "+initManagedMarker,
		"",
	)
	return strings.Join(lines, "\n")
}

func ensureRootTaskfile(root string) error {
	path := filepath.Join(root, "Taskfile.yml")
	includeBlock := []string{
		"  conventions:",
		"    taskfile: " + initConventionsTaskfile,
		"    dir: .convention-engineering",
		"    flatten: true",
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		body := []string{
			`version: "3"`,
			"",
			"includes:",
		}
		body = append(body, includeBlock...)
		body = append(body, "")
		return os.WriteFile(path, []byte(strings.Join(body, "\n")), 0o644)
	}

	text := string(content)
	if strings.Contains(text, initConventionsTaskfile) {
		return nil
	}

	lines := strings.Split(text, "\n")
	if !containsTopLevelKey(lines, "version:") {
		lines = append([]string{`version: "3"`, ""}, lines...)
	}

	insertAt := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "includes:" {
			insertAt = i + 1
			break
		}
	}
	if insertAt == -1 {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "includes:")
		insertAt = len(lines)
	}

	lines = append(lines[:insertAt], append(includeBlock, lines[insertAt:]...)...)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func containsTopLevelKey(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func ensureAgentContractFiles(root string, opts initOptions) error {
	block := renderAgentConventionBlock(opts)
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		path := filepath.Join(root, name)
		content, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		header := "# " + name + "\n\n"
		text := header
		if err == nil {
			text = string(content)
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
		}
		updated := upsertManagedBlock(text, block)
		if strings.TrimSpace(updated) == "" {
			updated = header + block
		}
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func renderAgentConventionBlock(opts initOptions) string {
	lines := []string{
		initManagedBlockStart,
		"## Conventions",
		"",
		"- **Docs** — tracked repo docs live under `docs/` using the `requests/`,",
		"  `planning/`, `plans/`, `implementation/`, and `taxonomy/` folders.",
	}
	if hasOperation(opts.Operations, "tickets") {
		lines = append(lines,
			"- **Tickets** — flat-file work tracker at `.tickets/`. Read `.tickets/README.md`",
			"  for the verb surface and `.tickets/harness/schema.yaml` for the state",
			"  machine. Daily commands:",
			"  `task -d .tickets {new|list|transition|close|test}`.",
		)
	}
	if hasOperation(opts.Operations, "wiki") {
		lines = append(lines,
			"- **Wiki** — LLM-maintained knowledge base at `.wiki/`. Read `.wiki/RULES.md`",
			"  for page types, frontmatter, and citation rules. Validate with",
			"  `task wiki:lint` (or `task -d .wiki lint`).",
		)
	}
	lines = append(lines,
		"- **Verification** — run `task verify` from the repo root to execute the",
		"  convention verification gate.",
		"- **Tracked contract** — `.convention-engineering.json` is the",
		"  machine-readable convention contract for this repo.",
		"",
		"Conventions are scaffolded by `agent-repo-kit` under `.convention-engineering/`.",
		initManagedBlockEnd,
		"",
	)
	return strings.Join(lines, "\n")
}

func upsertManagedBlock(existing, block string) string {
	start := strings.Index(existing, initManagedBlockStart)
	end := strings.Index(existing, initManagedBlockEnd)
	if start >= 0 && end >= start {
		end += len(initManagedBlockEnd)
		prefix := strings.TrimRight(existing[:start], "\n")
		suffix := strings.TrimLeft(existing[end:], "\n")
		switch {
		case prefix == "" && suffix == "":
			return block
		case prefix == "":
			return block + "\n" + suffix
		case suffix == "":
			return prefix + "\n\n" + block
		default:
			return prefix + "\n\n" + block + "\n" + suffix
		}
	}
	trimmed := strings.TrimRight(existing, "\n")
	if trimmed == "" {
		return block
	}
	return trimmed + "\n\n" + block
}
