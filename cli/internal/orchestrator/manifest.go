package orchestrator

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ChunkState mirrors the orchestratorChunkState JSON shape emitted into
// the handoff manifest.
type ChunkState struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	HardFailDimensions []string `json:"hard_fail_dimensions"`
	SoftFailDimensions []string `json:"soft_fail_dimensions"`
}

// EvidenceRecord captures a single evidence artifact referenced from the
// evidence manifest written by the orchestrator.
type EvidenceRecord struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	CWD         string `json:"cwd"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at"`
	ExitCode    int    `json:"exit_code"`
	ToolName    string `json:"tool_name"`
	ToolVersion string `json:"tool_version"`
	StdoutPath  string `json:"stdout_path"`
	StderrPath  string `json:"stderr_path"`
}

// EvidenceManifest bundles evidence records referenced by the handoff.
type EvidenceManifest struct {
	Records []EvidenceRecord `json:"records"`
}

// HandoffManifest is the JSON document the orchestrator writes for the
// evaluator to consume.
type HandoffManifest struct {
	ManifestID             string       `json:"manifest_id"`
	ContractPath           string       `json:"contract_path"`
	BriefPath              string       `json:"brief_path"`
	GeneratedArtifactPaths []string     `json:"generated_artifact_paths"`
	RawEvidenceBundlePath  string       `json:"raw_evidence_bundle_path"`
	LaunchReceiptPath      string       `json:"launch_receipt_path"`
	RequestedScope         string       `json:"requested_scope"`
	RequestedChunkID       string       `json:"requested_chunk_id,omitempty"`
	CheckerJSONPath        string       `json:"checker_json_path"`
	ReportPath             string       `json:"report_path"`
	ResultPath             string       `json:"result_path"`
	ChunkState             []ChunkState `json:"chunk_state"`
}

// LaunchReceipt captures the evaluator-launch metadata written before
// the evaluator runs.
type LaunchReceipt struct {
	ParentInvocationID    string `json:"parent_invocation_id"`
	EvaluatorInvocationID string `json:"evaluator_invocation_id"`
	LaunchMode            string `json:"launch_mode"`
	FreshContext          bool   `json:"fresh_context"`
	ForkContext           *bool  `json:"fork_context,omitempty"`
	HandoffManifestID     string `json:"handoff_manifest_id"`
	LaunchedAt            string `json:"launched_at"`
}

// EvaluationResult is the JSON document the evaluator writes as its final
// verdict.
type EvaluationResult struct {
	Scope              string   `json:"scope"`
	Status             string   `json:"status"`
	ChunkID            string   `json:"chunk_id,omitempty"`
	HardFailDimensions []string `json:"hard_fail_dimensions"`
	SoftFailDimensions []string `json:"soft_fail_dimensions"`
	ReportPath         string   `json:"report_path"`
	CheckerJSONPath    string   `json:"checker_json_path"`
	LaunchReceiptPath  string   `json:"launch_receipt_path"`
	GeneratedAt        string   `json:"generated_at"`
}

// resolveArtifactPath converts a repo-relative path into an absolute path
// rooted at root. Absolute paths are returned untouched.
func resolveArtifactPath(root, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(root, filepath.FromSlash(relPath))
}

// loadJSONArtifact reads a JSON artifact (absolute or repo-relative) and
// unmarshals it into dest.
func loadJSONArtifact(root, path string, dest any) error {
	full := resolveArtifactPath(root, path)
	data, err := os.ReadFile(full)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// writeJSONArtifact serializes value to an indented JSON file at the
// repo-relative path, creating parent directories as needed.
func writeJSONArtifact(root, relPath string, value any) error {
	full := resolveArtifactPath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

// writeTextArtifact writes a plain-text artifact at the repo-relative
// path, creating parent directories as needed.
func writeTextArtifact(root, relPath, content string) error {
	full := resolveArtifactPath(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}
