package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

func VersionCommand() command {
	return command{
		Description: "print build metadata",
		Run: func(app *appctx.AppContext, _ *cobra.Command, _ []string) error {
			data := map[string]string{
				"schema_version": "v1",
				"name":           binaryName,
				"version":        appVersion,
				"commit":         appCommit,
				"date":           appDate,
			}
			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			_, err := fmt.Fprintf(os.Stdout, "%s %s\n", data["name"], data["version"])
			return err
		},
	}
}
