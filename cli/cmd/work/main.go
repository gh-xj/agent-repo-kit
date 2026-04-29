package main

import (
	"os"

	"github.com/gh-xj/agent-repo-kit/cli/internal/workcli"
)

func main() {
	os.Exit(workcli.Execute(os.Args[1:]))
}
