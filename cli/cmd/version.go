package cmd

import (
	"encoding/json"
	"fmt"
)

// VersionCmd prints build metadata. Honors the global --json flag.
type VersionCmd struct{}

func (c *VersionCmd) Run(globals *CLI) error {
	out := globals.stdout()
	data := map[string]string{
		"schema_version": "v1",
		"name":           binaryName,
		"version":        appVersion,
		"commit":         appCommit,
		"date":           appDate,
	}
	if globals.JSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	_, err := fmt.Fprintf(out, "%s %s\n", data["name"], data["version"])
	return err
}
