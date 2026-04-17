package appctx

import "context"

const (
	ExitSuccess = 0
	ExitError   = 1
	ExitUsage   = 2
)

type AppMeta struct {
	Name    string
	Version string
	Commit  string
	Date    string
}

type AppContext struct {
	Context context.Context
	Meta    AppMeta
	Values  map[string]any
}

func NewAppContext(ctx context.Context) *AppContext {
	if ctx == nil {
		ctx = context.Background()
	}
	return &AppContext{Context: ctx, Values: map[string]any{}}
}

func ResolveExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	return ExitError
}
