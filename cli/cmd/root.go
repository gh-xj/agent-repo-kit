package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

const (
	binaryName = "ark"
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

type command struct {
	Description string
	Configure   func(*cobra.Command)
	Run         func(*appctx.AppContext, *cobra.Command, []string) error
}

var (
	commandRegistry         = map[string]command{}
	skillCommandRegistry    = map[string]command{}
	taskfileCommandRegistry = map[string]command{}
)

func init() {
	registerBuiltins()
	registerSkillCommand("init", SkillInitCommand())
	registerSkillCommand("audit", SkillAuditCommand())
}

func registerCommand(name string, cmd command) {
	commandRegistry[name] = cmd
}

func registerSkillCommand(name string, cmd command) {
	skillCommandRegistry[name] = cmd
}

func registerTaskfileCommand(name string, cmd command) {
	taskfileCommandRegistry[name] = cmd
}

func registerBuiltins() {
	registerCommand("version", VersionCommand())
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          binaryName,
		Short:        "agent-repo-kit unified CLI",
		SilenceUsage: true,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "enable debug logs")
	root.PersistentFlags().String("config", "", "config file path")
	root.PersistentFlags().Bool("json", false, "emit machine-readable JSON output")
	root.PersistentFlags().Bool("no-color", false, "disable colorized output")

	for _, name := range sortedKeys(commandRegistry) {
		root.AddCommand(newLeafCmd(name, commandRegistry[name]))
	}

	root.AddCommand(newSkillCmd())
	root.AddCommand(newTaskfileCmd())
	root.AddCommand(newAdaptersCmd())
	return root
}

func newSkillCmd() *cobra.Command {
	skill := &cobra.Command{
		Use:          "skill",
		Short:        "manage agent skills (init, audit)",
		SilenceUsage: true,
	}
	for _, name := range sortedKeys(skillCommandRegistry) {
		skill.AddCommand(newLeafCmd(name, skillCommandRegistry[name]))
	}
	return skill
}

func newTaskfileCmd() *cobra.Command {
	taskfile := &cobra.Command{
		Use:          "taskfile",
		Short:        "author and lint Taskfile.yml files",
		SilenceUsage: true,
	}
	for _, name := range sortedKeys(taskfileCommandRegistry) {
		taskfile.AddCommand(newLeafCmd(name, taskfileCommandRegistry[name]))
	}
	return taskfile
}

func newLeafCmd(name string, cmd command) *cobra.Command {
	child := &cobra.Command{
		Use:   name,
		Short: cmd.Description,
		RunE: func(command *cobra.Command, args []string) error {
			app := appctx.NewAppContext(command.Context())
			app.Meta = appctx.AppMeta{
				Name:    binaryName,
				Version: appVersion,
				Commit:  appCommit,
				Date:    appDate,
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
	return child
}

func sortedKeys(m map[string]command) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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

// usageErrorCode returns appctx.ExitUsage for cobra-generated usage errors
// (unknown flag/command, bad arg count, invalid flag value) and 0 for
// application errors. It matches cobra/pflag error-message *prefixes* so
// that application errors whose text happens to contain tokens like
// "requires" or "accepts" (e.g. a validator saying "requires justification")
// are not misclassified as exit 2.
func usageErrorCode(err error) int {
	if err == nil {
		return 0
	}
	text := strings.ToLower(err.Error())
	usagePrefixes := []string{
		"unknown command",
		"unknown flag",
		"unknown shorthand flag",
		"invalid argument",
		"flag needs an argument",
		"accepts ",       // cobra ExactArgs / MaximumNArgs / RangeArgs
		"requires at ",   // cobra MinimumNArgs
		"requires exact", // defensive variant
	}
	for _, prefix := range usagePrefixes {
		if strings.HasPrefix(text, prefix) {
			return appctx.ExitUsage
		}
	}
	return 0
}
