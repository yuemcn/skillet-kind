package cluster

import (
	"context"
	"fmt"
	"os/exec"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Cluster) Delete(ctx context.Context) error {
	fmt.Println("Checking if cluster exists")
	clusterExists, err := c.clusterExists()
	if err != nil {
		err = fmt.Errorf("an error occurred while checking if cluster exists: %w", err)
		fmt.Println(err)
		return err
	} else if !clusterExists {
		err = status.Errorf(codes.NotFound, "could not find cluster with name %v", c.Name)
		fmt.Println(err)
		return err
	}

	// delete cluster by name
	fmt.Println("Deleting cluster", c.Name)
	command := exec.Command("kind", "delete", "cluster", "--name", c.Name)
	err = command.Run()
	if err != nil {
		err = fmt.Errorf("failed to delete cluster: %w", err)
		fmt.Println(err)
		return err
	}

	fmt.Println("Cluster has been successfully deleted")
	return nil
}
