// Package adapters loads the adapters/manifest.json contract and provides
// helpers to expand skill-root paths. The manifest declares which harnesses
// exist and where their skill roots live; the set of skills to link is
// auto-derived from the `skills/` directory at the repo root — every
// subdirectory there that contains a SKILL.md is a skill.
package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// SkillsDir is the repo-root-relative directory that holds skill sources.
// Every immediate subdirectory with a SKILL.md is treated as a skill.
const SkillsDir = "skills"

// Manifest is the top-level shape of adapters/manifest.json.
type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	Harnesses     []Harness `json:"harnesses"`
}

// Harness describes a single harness (e.g. claude-code, codex) and the
// skill root under which skill symlinks are installed.
type Harness struct {
	Name      string `json:"name"`
	SkillRoot string `json:"skill_root"`
}

// Load reads and parses a manifest from disk.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ExpandSkillRoot resolves a leading "~/" in SkillRoot to the current user's
// home directory. Other paths are returned unchanged.
func (h *Harness) ExpandSkillRoot() (string, error) {
	if strings.HasPrefix(h.SkillRoot, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, h.SkillRoot[2:]), nil
	}
	return h.SkillRoot, nil
}
