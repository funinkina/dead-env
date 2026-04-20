package main

import (
	"context"
	"fmt"
	"os"

	"funinkina/deadenv/cmd"
)

func main() {
	root := cmd.NewRootCommand()

	if err := root.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
