package main

import (
	"os"

	"github.com/chickenzord/linkding-cli/internal/cli"
)

func main() {
	root := cli.NewRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(cli.ExitError)
	}
}
