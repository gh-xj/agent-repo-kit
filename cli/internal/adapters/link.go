package adapters

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// RunLink symlinks every skill under <repoRoot>/skills/ into the named
// harness's skill root. A skill is any immediate subdirectory of
// <repoRoot>/skills/ that contains a SKILL.md.
//
// Symlink semantics match install.sh's historical link() helper:
//   - missing source        → error.
//   - existing symlink dst  → remove and recreate.
//   - existing non-symlink  → warn and skip.
//   - absent dst            → create, making parent dirs as needed.
//
// When dryRun is true, actions are printed to stdout but the filesystem
// is left untouched.
func RunLink(repoRoot, manifestPath, target string, dryRun bool) error {
	return runLink(os.Stdout, os.Stderr, repoRoot, manifestPath, target, dryRun)
}

func runLink(stdout, stderr io.Writer, repoRoot, manifestPath, target string, dryRun bool) error {
	harness, skillRoot, repoAbs, err := resolveHarness(repoRoot, manifestPath, target)
	if err != nil {
		return err
	}
	_ = harness

	skills, err := discoverSkills(repoAbs)
	if err != nil {
		return err
	}

	for _, skill := range skills {
		srcAbs := filepath.Join(repoAbs, SkillsDir, skill)
		dstAbs := filepath.Join(skillRoot, skill)
		if err := ensureSymlink(stdout, stderr, srcAbs, dstAbs, dryRun); err != nil {
			return err
		}
	}
	return nil
}

// ListLinks writes one `<srcAbs>\t<dstAbs>` line per discovered skill for
// the named target. Read-only helper for shell consumers.
func ListLinks(w io.Writer, repoRoot, manifestPath, target string) error {
	_, skillRoot, repoAbs, err := resolveHarness(repoRoot, manifestPath, target)
	if err != nil {
		return err
	}
	skills, err := discoverSkills(repoAbs)
	if err != nil {
		return err
	}
	for _, skill := range skills {
		srcAbs := filepath.Join(repoAbs, SkillsDir, skill)
		dstAbs := filepath.Join(skillRoot, skill)
		fmt.Fprintf(w, "%s\t%s\n", srcAbs, dstAbs)
	}
	return nil
}

// resolveHarness loads the manifest, finds the named harness, and returns
// the expanded skill root plus absolute repo root.
func resolveHarness(repoRoot, manifestPath, target string) (*Harness, string, string, error) {
	m, err := Load(manifestPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("load manifest %q: %w", manifestPath, err)
	}
	var harness *Harness
	for i := range m.Harnesses {
		if m.Harnesses[i].Name == target {
			harness = &m.Harnesses[i]
			break
		}
	}
	if harness == nil {
		return nil, "", "", fmt.Errorf("unknown target %q", target)
	}
	skillRoot, err := harness.ExpandSkillRoot()
	if err != nil {
		return nil, "", "", fmt.Errorf("expand skill_root: %w", err)
	}
	repoAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, "", "", fmt.Errorf("resolve repo root: %w", err)
	}
	return harness, skillRoot, repoAbs, nil
}

// discoverSkills returns the sorted names of immediate subdirectories of
// <repoAbs>/skills/ that contain a SKILL.md file. Dirs without SKILL.md
// are silently skipped so incidental files (README, scratch) never
// become accidental skill symlinks.
func discoverSkills(repoAbs string) ([]string, error) {
	skillsRoot := filepath.Join(repoAbs, SkillsDir)
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, fmt.Errorf("read skills dir %q: %w", skillsRoot, err)
	}
	skills := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMd := filepath.Join(skillsRoot, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMd); err != nil {
			continue
		}
		skills = append(skills, entry.Name())
	}
	sort.Strings(skills)
	return skills, nil
}

// ensureSymlink installs (or repairs) a symlink dst → src, honoring dryRun.
func ensureSymlink(stdout, stderr io.Writer, src, dst string, dryRun bool) error {
	if _, err := os.Lstat(src); err != nil {
		return fmt.Errorf("source path missing: %s", src)
	}

	fi, err := os.Lstat(dst)
	switch {
	case err == nil && fi.Mode()&os.ModeSymlink != 0:
		fmt.Fprintf(stdout, "[adapters] re-linking %s\n", dst)
		if dryRun {
			fmt.Fprintf(stdout, "DRY-RUN: rm %q\n", dst)
		} else {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("remove existing symlink %q: %w", dst, err)
			}
		}
	case err == nil:
		fmt.Fprintf(stderr, "[adapters] WARN: %s already exists and is not a symlink — skipping (remove it to re-link)\n", dst)
		return nil
	case !os.IsNotExist(err):
		return fmt.Errorf("stat %q: %w", dst, err)
	}

	parent := filepath.Dir(dst)
	if dryRun {
		fmt.Fprintf(stdout, "DRY-RUN: mkdir -p %q\n", parent)
		fmt.Fprintf(stdout, "DRY-RUN: ln -s %q %q\n", src, dst)
		return nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", parent, err)
	}
	fmt.Fprintf(stdout, "[adapters] ln -s %s %s\n", src, dst)
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("symlink %q -> %q: %w", dst, src, err)
	}
	return nil
}
