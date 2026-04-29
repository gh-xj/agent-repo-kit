package scaffold

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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
		".work":                   true,
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
