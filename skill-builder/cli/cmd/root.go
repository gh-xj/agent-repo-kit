package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/skill-builder/cli/internal/appctx"
)

type command struct {
	Description string
	Configure   func(*cobra.Command)
	Run         func(*appctx.AppContext, *cobra.Command, []string) error
}

var commandRegistry = map[string]command{}

func init() {
	registerBuiltins()
	registerCommand("init", InitCommand())
	registerCommand("audit", AuditCommand())
}

func registerCommand(name string, cmd command) {
	commandRegistry[name] = cmd
}

func registerBuiltins() {
	registerCommand("version", command{
		Description: "print build metadata",
		Run: func(app *appctx.AppContext, _ *cobra.Command, _ []string) error {
			data := map[string]string{
				"schema_version": "v1",
				"name":           "skill-builder",
				"version":        "dev",
				"commit":         "none",
				"date":           "unknown",
			}
			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			_, err := fmt.Fprintf(os.Stdout, "%s %s (%s %s)\n", data["name"], data["version"], data["commit"], data["date"])
			return err
		},
	})
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "skill-builder",
		Short:        "skill-builder CLI",
		SilenceUsage: true,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "enable debug logs")
	root.PersistentFlags().String("config", "", "config file path")
	root.PersistentFlags().Bool("json", false, "emit machine-readable JSON output")
	root.PersistentFlags().Bool("no-color", false, "disable colorized output")

	names := make([]string, 0, len(commandRegistry))
	for name := range commandRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cmd := commandRegistry[name]
		child := &cobra.Command{
			Use:   name,
			Short: cmd.Description,
			RunE: func(command *cobra.Command, args []string) error {
				app := appctx.NewAppContext(command.Context())
				app.Meta = appctx.AppMeta{
					Name:    "skill-builder",
					Version: "dev",
					Commit:  "none",
					Date:    "unknown",
				}

				jsonFlag, _ := command.Flags().GetBool("json")
				configPath, _ := command.Flags().GetString("config")
				noColor, _ := command.Flags().GetBool("no-color")
				app.Values["json"] = jsonFlag
				app.Values["config"] = configPath
				app.Values["no-color"] = noColor

				if cmd.Run == nil {
					return nil
				}
				return cmd.Run(app, command, args)
			},
		}
		if cmd.Configure != nil {
			cmd.Configure(child)
		}
		root.AddCommand(child)
	}
	return root
}

func Execute(args []string) int {
	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return resolveCode(err)
	}
	return appctx.ExitSuccess
}

func resolveCode(err error) int {
	if err == nil {
		return appctx.ExitSuccess
	}
	if code := usageErrorCode(err); code != 0 {
		return code
	}
	return appctx.ResolveExitCode(err)
}

func usageErrorCode(err error) int {
	if err == nil {
		return 0
	}
	text := err.Error()
	usageIndicators := []string{
		"unknown command",
		"unknown flag",
		"accepts",
		"requires",
		"usage:",
	}
	for _, marker := range usageIndicators {
		if strings.Contains(strings.ToLower(text), marker) {
			return appctx.ExitUsage
		}
	}
	return 0
}
