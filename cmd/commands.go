package cmd

import (
	"context"
	"skillet/skillet-kind/cluster"

	"github.com/urfave/cli/v3"
)

var CreateCommand = &cli.Command{
	Name:  "create",
	Usage: "create a new cluster",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "The name of the cluster to be created",
		},
		&cli.StringFlag{
			Name:  "file",
			Usage: "Path to the file containing the cluster configuration",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cluster := cluster.NewCluster(cmd.String("name"), cmd.String("file"))
		err := cluster.Create(ctx)
		return err
	},
}

var DeleteCommand = &cli.Command{
	Name:  "delete",
	Usage: "delete a cluster",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "The name of the cluster to be deleted",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cluster := cluster.NewCluster(cmd.String("name"), cmd.String("file"))
		err := cluster.Delete(ctx)
		return err
	},
}
