package adapters

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RunLink wires each manifest link for the named target into place,
// creating (or repairing) the symlink at dstAbs that points at srcAbs.
//
// Semantics mirror install.sh:85-98 exactly:
//   - missing source → error.
//   - existing symlink at dst → remove and recreate.
//   - existing non-symlink at dst → warn and skip.
//   - absent dst → create, making parent dirs as needed.
//
// When dryRun is true, actions are printed to stdout but the filesystem
// is left untouched.
func RunLink(repoRoot, manifestPath, target string, dryRun bool) error {
	return runLink(os.Stdout, os.Stderr, repoRoot, manifestPath, target, dryRun)
}

func runLink(stdout, stderr io.Writer, repoRoot, manifestPath, target string, dryRun bool) error {
	// Step 1: load manifest.
	m, err := Load(manifestPath)
	if err != nil {
		return fmt.Errorf("load manifest %q: %w", manifestPath, err)
	}

	// Step 2: find harness by name.
	var harness *Harness
	for i := range m.Harnesses {
		if m.Harnesses[i].Name == target {
			harness = &m.Harnesses[i]
			break
		}
	}
	if harness == nil {
		return fmt.Errorf("unknown target %q", target)
	}

	// Step 3: expand skill root.
	skillRoot, err := harness.ExpandSkillRoot()
	if err != nil {
		return fmt.Errorf("expand skill_root: %w", err)
	}

	repoAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	// Steps 4 + 5: for each link, verify source and ensure symlink.
	for _, link := range harness.Links {
		srcAbs := filepath.Join(repoAbs, link.Source)
		dstAbs := filepath.Join(skillRoot, link.Dest)
		if _, err := os.Lstat(srcAbs); err != nil {
			return fmt.Errorf("source path missing: %s", srcAbs)
		}
		if err := ensureSymlink(stdout, stderr, srcAbs, dstAbs, dryRun); err != nil {
			return err
		}
	}
	return nil
}

// ensureSymlink mirrors the install.sh `link()` helper semantics.
func ensureSymlink(stdout, stderr io.Writer, src, dst string, dryRun bool) error {
	fi, err := os.Lstat(dst)
	switch {
	case err == nil && fi.Mode()&os.ModeSymlink != 0:
		fmt.Fprintf(stderr, "[adapters] WARN: %s is a symlink — re-linking\n", dst)
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

// ListLinks resolves each link for the named target to absolute paths and
// writes `<srcAbs>\t<dstAbs>` per line to w. It is a read-only helper for
// shell consumers.
func ListLinks(w io.Writer, repoRoot, manifestPath, target string) error {
	m, err := Load(manifestPath)
	if err != nil {
		return fmt.Errorf("load manifest %q: %w", manifestPath, err)
	}
	var harness *Harness
	for i := range m.Harnesses {
		if m.Harnesses[i].Name == target {
			harness = &m.Harnesses[i]
			break
		}
	}
	if harness == nil {
		return fmt.Errorf("unknown target %q", target)
	}
	skillRoot, err := harness.ExpandSkillRoot()
	if err != nil {
		return fmt.Errorf("expand skill_root: %w", err)
	}
	repoAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	for _, link := range harness.Links {
		srcAbs := filepath.Join(repoAbs, link.Source)
		dstAbs := filepath.Join(skillRoot, link.Dest)
		fmt.Fprintf(w, "%s\t%s\n", srcAbs, dstAbs)
	}
	return nil
}
