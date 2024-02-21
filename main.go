package main

import (
	"context"
	"os"
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
		},
	}

	cmd.Run(ctx, os.Args)
}
