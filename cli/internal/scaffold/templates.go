package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// resolveConventionEngineeringRoot locates the `convention-engineering`
// directory that ships alongside this CLI module. The package file lives at
// `cli/internal/scaffold/`, and templates at
// `convention-engineering/references/templates/`, so we walk up from the
// package file looking for a sibling `convention-engineering` tree with a
// `references/templates` directory.
func resolveConventionEngineeringRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve convention-engineering source root")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 12; i++ {
		candidate := filepath.Join(dir, "convention-engineering")
		if stat, err := os.Stat(filepath.Join(candidate, "references", "templates")); err == nil && stat.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("failed to locate convention-engineering root from %s", thisFile)
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
	if data, err := os.ReadFile(path); err == nil && !strings.Contains(string(data), ManagedMarker) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), mode)
}
