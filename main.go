package main

import (
	"os"

	"context"
	"skillet/skillet-kind/cmd"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()
	cmd := &cli.Command{
		Name:  "skillet",
		Usage: "CLI tool to deploy kind clusters",
		Commands: []*cli.Command{
			cmd.CreateCommand,
			cmd.DeleteCommand,
			cmd.ValidateCommand,
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
