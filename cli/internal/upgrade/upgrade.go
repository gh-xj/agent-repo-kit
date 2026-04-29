// Package upgrade implements `ark upgrade` for two binary flavors:
//   - Clone: the binary is inside a git checkout of agent-repo-kit. Upgrade
//     = git pull + go build + os.Rename swap.
//   - Prebuilt: the binary was installed from a release archive. Upgrade =
//     download new archive, verify SHA-256, atomic swap with .old recovery.
package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gh-xj/agent-repo-kit/cli/internal/adapters"
)

// Flavor indicates which upgrade path to run.
type Flavor int

const (
	// FlavorClone means selfPath lives inside a git checkout.
	FlavorClone Flavor = iota
	// FlavorPrebuilt means selfPath was installed from a release archive.
	FlavorPrebuilt
)

// releaseOwner / releaseRepo match the URL shape in install-v2.md §1.
const (
	releaseOwner = "gh-xj"
	releaseRepo  = "agent-repo-kit"
)

// DetectFlavor walks up from selfPath looking for a `.git` directory. If
// found, the binary is a clone build; otherwise it is a prebuilt binary.
func DetectFlavor(selfPath string) Flavor {
	dir := filepath.Dir(selfPath)
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return FlavorClone
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return FlavorPrebuilt
}

// findGitRoot returns the nearest ancestor of selfPath that contains a
// `.git` directory. Returns "" if none found.
func findGitRoot(selfPath string) string {
	dir := filepath.Dir(selfPath)
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// RunUpgrade performs an in-place upgrade of the ark/work binaries.
// target selects the harness to re-link after upgrade (empty = skip link).
// prefix is the directory holding the current binary; empty uses os.Executable.
func RunUpgrade(target, prefix string, dryRun bool) error {
	return runUpgrade(os.Stdout, os.Stderr, target, prefix, dryRun)
}

func runUpgrade(stdout, stderr io.Writer, target, prefix string, dryRun bool) error {
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve self path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(selfPath); err == nil {
		selfPath = resolved
	}
	if prefix == "" {
		prefix = filepath.Dir(selfPath)
	} else {
		selfPath = filepath.Join(prefix, filepath.Base(selfPath))
	}

	// Interrupted-upgrade recovery.
	if err := recoverInterrupted(stdout, stderr, selfPath, dryRun); err != nil {
		return err
	}

	flavor := DetectFlavor(selfPath)
	switch flavor {
	case FlavorClone:
		return runCloneUpgrade(stdout, stderr, selfPath, target, dryRun)
	case FlavorPrebuilt:
		return runPrebuiltUpgrade(stdout, stderr, selfPath, target, dryRun)
	default:
		return fmt.Errorf("unknown flavor %d", flavor)
	}
}

func recoverInterrupted(stdout, stderr io.Writer, selfPath string, dryRun bool) error {
	newPath := selfPath + ".new"
	oldPath := selfPath + ".old"

	// If a .new exists, a prior upgrade got interrupted before the final
	// rename. Resume by promoting .new → selfPath.
	if _, err := os.Stat(newPath); err == nil {
		fmt.Fprintf(stderr, "[upgrade] recovering interrupted upgrade: promoting %s -> %s\n", newPath, selfPath)
		if dryRun {
			fmt.Fprintf(stdout, "DRY-RUN: mv %q %q\n", newPath, selfPath)
		} else if err := os.Rename(newPath, selfPath); err != nil {
			return fmt.Errorf("recover .new: %w", err)
		}
	}
	// If only .old exists (selfPath missing), restore .old → selfPath and
	// tell the user to re-run upgrade.
	if _, err := os.Stat(selfPath); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(oldPath); err == nil {
			if dryRun {
				fmt.Fprintf(stdout, "DRY-RUN: mv %q %q\n", oldPath, selfPath)
				return fmt.Errorf("found orphan %s with missing selfPath; re-run `ark upgrade`", oldPath)
			}
			if err := os.Rename(oldPath, selfPath); err != nil {
				return fmt.Errorf("restore .old: %w", err)
			}
			return fmt.Errorf("recovered %s from %s; re-run `ark upgrade`", selfPath, oldPath)
		}
	}
	return nil
}

func runCloneUpgrade(stdout, stderr io.Writer, selfPath, target string, dryRun bool) error {
	repoRoot := findGitRoot(selfPath)
	if repoRoot == "" {
		return fmt.Errorf("clone flavor but no .git ancestor of %s", selfPath)
	}
	cliDir := filepath.Join(repoRoot, "cli")
	newBin := selfPath + ".new"

	fmt.Fprintf(stdout, "[upgrade] clone flavor, repo=%s\n", repoRoot)

	steps := []struct {
		label string
		cmd   []string
		dir   string
	}{
		{"git pull --ff-only", []string{"git", "-C", repoRoot, "pull", "--ff-only"}, ""},
		{"go build ark", []string{"go", "build", "-o", newBin, "./cmd/ark"}, cliDir},
		{"go build work", []string{"go", "build", "-o", filepath.Join(filepath.Dir(selfPath), "work.new"), "./cmd/work"}, cliDir},
	}

	for _, s := range steps {
		if dryRun {
			fmt.Fprintf(stdout, "DRY-RUN: %s (cwd=%s) %s\n", s.label, s.dir, strings.Join(s.cmd, " "))
			continue
		}
		fmt.Fprintf(stdout, "[upgrade] %s\n", s.label)
		cmd := exec.Command(s.cmd[0], s.cmd[1:]...)
		if s.dir != "" {
			cmd.Dir = s.dir
		}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s failed: %w", s.label, err)
		}
	}

	if dryRun {
		fmt.Fprintf(stdout, "DRY-RUN: mv %q %q\n", newBin, selfPath)
		fmt.Fprintf(stdout, "DRY-RUN: mv %q %q\n", filepath.Join(filepath.Dir(selfPath), "work.new"), filepath.Join(filepath.Dir(selfPath), "work"))
	} else {
		if err := os.Rename(newBin, selfPath); err != nil {
			return fmt.Errorf("swap binary: %w", err)
		}
		if err := os.Rename(filepath.Join(filepath.Dir(selfPath), "work.new"), filepath.Join(filepath.Dir(selfPath), "work")); err != nil {
			return fmt.Errorf("swap work binary: %w", err)
		}
	}

	if target == "" {
		return nil
	}
	manifestPath := filepath.Join(repoRoot, "adapters", "manifest.json")
	if dryRun {
		fmt.Fprintf(stdout, "DRY-RUN: ark adapters link --target %s --manifest %s --repo-root %s\n", target, manifestPath, repoRoot)
		return nil
	}
	return adapters.RunLink(repoRoot, manifestPath, target, false)
}

func runPrebuiltUpgrade(stdout, stderr io.Writer, selfPath, target string, dryRun bool) error {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	if osName != "darwin" && osName != "linux" {
		return fmt.Errorf("prebuilt upgrade unsupported on %s", osName)
	}

	installDir := filepath.Dir(selfPath)
	workPath := filepath.Join(installDir, "work")
	fmt.Fprintf(stdout, "[upgrade] prebuilt flavor, os=%s arch=%s\n", osName, archName)

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("fetch latest version: %w", err)
	}
	fmt.Fprintf(stdout, "[upgrade] latest release: v%s\n", latest)

	archive := fmt.Sprintf("ark-%s-%s-%s.tar.gz", latest, osName, archName)
	archiveURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s",
		releaseOwner, releaseRepo, latest, archive)
	checksumURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/checksums.txt",
		releaseOwner, releaseRepo, latest)

	if dryRun {
		fmt.Fprintf(stdout, "DRY-RUN: download %s\n", archiveURL)
		fmt.Fprintf(stdout, "DRY-RUN: download %s\n", checksumURL)
		fmt.Fprintf(stdout, "DRY-RUN: verify SHA-256\n")
		fmt.Fprintf(stdout, "DRY-RUN: mv %q %q; extract ark/work; mv .new binaries into %s\n", selfPath, selfPath+".old", installDir)
		if target != "" {
			fmt.Fprintf(stdout, "DRY-RUN: ark adapters link --target %s\n", target)
		}
		return nil
	}

	tmp, err := os.MkdirTemp("", "ark-upgrade-")
	if err != nil {
		return fmt.Errorf("mktemp: %w", err)
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, archive)
	checksumPath := filepath.Join(tmp, "checksums.txt")
	if err := downloadFile(archiveURL, archivePath); err != nil {
		return fmt.Errorf("download archive: %w", err)
	}
	if err := downloadFile(checksumURL, checksumPath); err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	if err := verifyChecksum(checksumPath, archivePath, archive); err != nil {
		return fmt.Errorf("verify checksum: %w", err)
	}

	extractedArk := filepath.Join(tmp, "ark")
	if err := extractBinaryFromTarGz(archivePath, "ark", extractedArk); err != nil {
		return fmt.Errorf("extract ark: %w", err)
	}
	extractedWork := filepath.Join(tmp, "work")
	if err := extractBinaryFromTarGz(archivePath, "work", extractedWork); err != nil {
		return fmt.Errorf("extract work: %w", err)
	}

	oldPath := selfPath + ".old"
	newPath := selfPath + ".new"
	workOldPath := workPath + ".old"
	workNewPath := workPath + ".new"

	if err := os.Rename(extractedArk, newPath); err != nil {
		return fmt.Errorf("stage new ark binary: %w", err)
	}
	if err := os.Chmod(newPath, 0o755); err != nil {
		return fmt.Errorf("chmod new ark binary: %w", err)
	}
	if err := os.Rename(extractedWork, workNewPath); err != nil {
		_ = os.Remove(newPath)
		return fmt.Errorf("stage new work binary: %w", err)
	}
	if err := os.Chmod(workNewPath, 0o755); err != nil {
		_ = os.Remove(newPath)
		_ = os.Remove(workNewPath)
		return fmt.Errorf("chmod new work binary: %w", err)
	}

	// Move current binaries to .old as recovery points.
	if err := os.Rename(selfPath, oldPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}
	hadWork := false
	if _, err := os.Stat(workPath); err == nil {
		hadWork = true
		if err := os.Rename(workPath, workOldPath); err != nil {
			_ = os.Rename(oldPath, selfPath)
			return fmt.Errorf("backup current work binary: %w", err)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Rename(oldPath, selfPath)
		return fmt.Errorf("stat work binary: %w", err)
	}

	if err := os.Rename(newPath, selfPath); err != nil {
		// leave .old in place so operator can recover
		return fmt.Errorf("swap binary: %w; recover with: mv %s %s", err, oldPath, selfPath)
	}
	if err := os.Rename(workNewPath, workPath); err != nil {
		_ = os.Rename(oldPath, selfPath)
		if hadWork {
			_ = os.Rename(workOldPath, workPath)
		}
		return fmt.Errorf("swap work binary: %w; restored ark/work from backups", err)
	}

	if err := os.Remove(oldPath); err != nil {
		fmt.Fprintf(stderr, "[upgrade] WARN: failed to remove %s: %v\n", oldPath, err)
	}
	if hadWork {
		if err := os.Remove(workOldPath); err != nil {
			fmt.Fprintf(stderr, "[upgrade] WARN: failed to remove %s: %v\n", workOldPath, err)
		}
	}

	if target == "" {
		return nil
	}
	// Prebuilt installs don't know their repo root. Without a checkout the
	// link step requires an explicit manifest and repo root from the
	// operator; skip silently here and defer to `ark adapters link`.
	fmt.Fprintf(stdout, "[upgrade] prebuilt install: run `ark adapters link --target %s --repo-root <checkout>` separately\n", target)
	return nil
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", releaseOwner, releaseRepo)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API status %d", resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	tag := strings.TrimPrefix(payload.TagName, "v")
	if tag == "" {
		return "", fmt.Errorf("empty tag_name in API response")
	}
	return tag, nil
}

func downloadFile(url, dst string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(checksumFile, archivePath, archiveName string) error {
	data, err := os.ReadFile(checksumFile)
	if err != nil {
		return err
	}
	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// sha256sum format: "<hash>  <filename>" — filename may have a
		// leading "*" binary-mode marker we should strip.
		name := strings.TrimPrefix(fields[1], "*")
		if name == archiveName {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("no checksum entry for %s", archiveName)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != expected {
		return fmt.Errorf("checksum mismatch: want %s got %s", expected, got)
	}
	return nil
}

func extractBinaryFromTarGz(archivePath, binaryName, destBin string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%s binary not found in archive", binaryName)
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) != binaryName || hdr.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.OpenFile(destBin, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
}
