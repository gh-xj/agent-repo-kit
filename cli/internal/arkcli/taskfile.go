package arkcli

// TaskfileCmd groups the `ark taskfile` subcommands.
type TaskfileCmd struct {
	Lint TaskfileLintCmd `cmd:"" help:"lint Taskfile.yml against structural rules"`
}
