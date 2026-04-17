package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	binaryName = "ark"
	appVersion = "dev"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          binaryName,
		Short:        "agent-repo-kit unified CLI",
		SilenceUsage: true,
	}
	root.PersistentFlags().Bool("json", false, "emit machine-readable JSON output")
	root.AddCommand(newVersionCmd())
	return root
}

func Execute(args []string) int {
	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
