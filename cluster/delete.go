package cluster

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kind_cluster "sigs.k8s.io/kind/pkg/cluster"
)

func (c *Cluster) Delete(ctx context.Context) error {
	slog.Info("Checking if cluster exists")
	clusterExists, err := c.clusterExists()
	if err != nil {
		err = fmt.Errorf("an error occurred while checking if cluster exists: %w", err)
		slog.Error(err.Error())
		return err
	} else if !clusterExists {
		err = status.Errorf(codes.NotFound, "could not find cluster with name %v", c.Name)
		slog.Error(err.Error())
		return err
	}

	// delete cluster by name
	slog.Info("Deleting cluster", "cluster", c.Name)

	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		err = fmt.Errorf("error getting kubeconfig path: %w", err)
		slog.Error(err.Error())
		return err
	}

	provider := kind_cluster.NewProvider()
	err = provider.Delete(c.Name, kubeconfig)
	if err != nil {
		err = fmt.Errorf("error deleting cluster: %w", err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Cluster has been successfully deleted")
	return nil
}
