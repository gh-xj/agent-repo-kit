package contract

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// Report is the machine-readable contract run summary emitted by Run in
// --json mode.
type Report struct {
	RepoRoot   string        `json:"repo_root"`
	ConfigPath string        `json:"config_path"`
	Failed     int           `json:"failed"`
	Results    []CheckResult `json:"results"`
}

// Run resolves the repository root, loads the contract config, executes all
// checks, and writes human- or machine-readable output to stdout. Errors
// loading the config or encoding JSON are written to stderr. It returns the
// process exit code: 0 on pass, 1 on check failures, 2 on config/IO errors.
func Run(repoRoot, configPath string, jsonMode bool, stdout, stderr io.Writer) int {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve repo root: %v\n", err)
		return 2
	}

	cfg, resolvedConfigPath, err := LoadConfig(root, configPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 2
	}

	results := RunChecks(root, cfg)
	failed := CountFailures(results)

	if jsonMode {
		report := Report{
			RepoRoot:   root,
			ConfigPath: resolvedConfigPath,
			Failed:     failed,
			Results:    results,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(stderr, "failed to encode json: %v\n", err)
			return 2
		}
		if failed > 0 {
			return 1
		}
		return 0
	}

	printResults(stdout, results)

	if failed > 0 {
		fmt.Fprintf(stdout, "\nConvention contract failed: %d check(s) failed.\n", failed)
		return 1
	}

	fmt.Fprintln(stdout, "\nConvention contract passed.")
	return 0
}

func printResults(stdout io.Writer, results []CheckResult) {
	for _, r := range results {
		if r.Passed {
			fmt.Fprintf(stdout, "[PASS] %s\n", r.Name)
			continue
		}
		fmt.Fprintf(stdout, "[FAIL] %s: %s\n", r.Name, r.Detail)
	}
}
