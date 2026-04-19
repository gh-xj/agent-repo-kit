package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/orchestrator"
)

// OrchestrateCmd wires the convention orchestrator (brief → checker →
// handoff → evaluator launch → evaluation result) behind `ark orchestrate`.
// As of Stage 5 the evaluator is in-process; the former --evaluator-path /
// --evaluator-script flags are no longer needed.
//
// The contract-checker config path comes from the global --config flag
// on CLI (previously there was a duplicate local --config; that
// duplicate served no purpose and was removed in the kong migration).
type OrchestrateCmd struct {
	RepoRoot           string `name:"repo-root" help:"path to repository root" default:"."`
	Topic              string `help:"stable topic label for orchestration artifacts" default:"convention-run"`
	Scope              string `help:"evaluation scope: final or chunk" enum:"final,chunk" default:"final"`
	ChunkID            string `name:"chunk-id" help:"chunk id for chunk-scoped orchestration"`
	GeneratedArtifacts string `name:"generated-artifacts" help:"comma-separated repo-relative artifact paths under review"`
	ParentInvocationID string `name:"parent-invocation-id" help:"parent invocation id for orchestration launch receipts" default:"manual"`
}

func (c *OrchestrateCmd) Run(globals *CLI) error {
	root, err := filepath.Abs(c.RepoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	req := orchestrator.Request{
		Topic:                  c.Topic,
		ParentInvocationID:     c.ParentInvocationID,
		RequestedScope:         c.Scope,
		RequestedChunkID:       c.ChunkID,
		GeneratedArtifactPaths: orchestrator.ParseGeneratedArtifactList(c.GeneratedArtifacts),
	}

	launcher := orchestrator.NewProcessEvaluatorLauncher(root, c.ParentInvocationID, time.Now)
	exitCode := orchestrator.Run(root, globals.Config, req, launcher, globals.stdout(), globals.stderr(), time.Now)
	if exitCode == appctx.ExitSuccess {
		return nil
	}
	return appctx.NewExitError(exitCode, "")
}
