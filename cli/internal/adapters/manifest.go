// Package adapters loads the adapters/manifest.json contract and provides
// helpers to expand skill-root paths. The manifest is the single source of
// truth for which skills `ark adapters link` wires into each harness.
package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is the top-level shape of adapters/manifest.json.
type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	Harnesses     []Harness `json:"harnesses"`
}

// Harness describes a single harness (e.g. claude-code, codex) and the
// set of skill symlinks to install under its skill root.
type Harness struct {
	Name      string `json:"name"`
	SkillRoot string `json:"skill_root"`
	Links     []Link `json:"links"`
}

// Link is a single source→dest symlink entry. Source is repo-root-relative;
// dest is skill-root-relative.
type Link struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
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
