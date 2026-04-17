package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
	"github.com/gh-xj/agent-repo-kit/cli/internal/orchestrator"
)

func init() {
	registerCommand("orchestrate", OrchestrateCommand())
}

// OrchestrateCommand wires the convention orchestrator (brief → checker
// → handoff → evaluator launch → evaluation result) behind `ark
// orchestrate`. As of Stage 5 the evaluator is in-process; the former
// --evaluator-path / --evaluator-script flags are no longer needed.
func OrchestrateCommand() command {
	return command{
		Description: "run the convention orchestrator and launch the evaluator",
		Configure: func(command *cobra.Command) {
			command.Flags().String("repo-root", ".", "path to repository root")
			command.Flags().String("config", "", "path to contract checker config JSON (defaults to tracked .convention-engineering.json)")
			command.Flags().String("topic", "convention-run", "stable topic label for orchestration artifacts")
			command.Flags().String("scope", orchestrator.ScopeFinal, "evaluation scope: final or chunk")
			command.Flags().String("chunk-id", "", "chunk id for chunk-scoped orchestration")
			command.Flags().String("generated-artifacts", "", "comma-separated repo-relative artifact paths under review")
			command.Flags().String("parent-invocation-id", "manual", "parent invocation id for orchestration launch receipts")
		},
		Run: func(app *appctx.AppContext, command *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unexpected positional args: %s", strings.Join(args, " "))
			}

			repoRoot, _ := command.Flags().GetString("repo-root")
			configPath, _ := command.Flags().GetString("config")
			topic, _ := command.Flags().GetString("topic")
			scope, _ := command.Flags().GetString("scope")
			chunkID, _ := command.Flags().GetString("chunk-id")
			generatedArtifacts, _ := command.Flags().GetString("generated-artifacts")
			parentInvocationID, _ := command.Flags().GetString("parent-invocation-id")

			// Fall back to the persistent --config flag when the
			// orchestrate-local --config was not supplied.
			if strings.TrimSpace(configPath) == "" {
				if persistent, ok := app.Values["config"].(string); ok {
					configPath = persistent
				}
			}

			root, err := filepath.Abs(repoRoot)
			if err != nil {
				return fmt.Errorf("resolve repo root: %w", err)
			}

			req := orchestrator.Request{
				Topic:                  topic,
				ParentInvocationID:     parentInvocationID,
				RequestedScope:         scope,
				RequestedChunkID:       chunkID,
				GeneratedArtifactPaths: orchestrator.ParseGeneratedArtifactList(generatedArtifacts),
			}

			launcher := orchestrator.NewProcessEvaluatorLauncher(root, parentInvocationID, time.Now)
			exitCode := orchestrator.Run(root, configPath, req, launcher, os.Stdout, os.Stderr, time.Now)
			if exitCode == appctx.ExitSuccess {
				return nil
			}
			return appctx.NewExitError(exitCode, "")
		},
	}
}
