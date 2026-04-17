package main

import (
	"os"

	"github.com/gh-xj/agent-repo-kit/cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
