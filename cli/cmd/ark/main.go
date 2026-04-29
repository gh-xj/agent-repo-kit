package main

import (
	"os"

	"github.com/gh-xj/agent-repo-kit/cli/internal/arkcli"
)

func main() {
	os.Exit(arkcli.Execute(os.Args[1:]))
}
