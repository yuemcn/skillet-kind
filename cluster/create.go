package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"skillet/skillet-kind/charts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kind_cluster "sigs.k8s.io/kind/pkg/cluster"
)

func (c *Cluster) Create(ctx context.Context) error {

	slog.Info("Checking if cluster exists")
	clusterExists, err := c.clusterExists()
	if err != nil {
		err = fmt.Errorf("an error occurred while checking if cluster exists: %w", err)
		slog.Error(err.Error())
		return err
	} else if clusterExists {
		err = status.Errorf(codes.AlreadyExists, "a cluster with name %v already exists", c.Name)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Starting creation of cluster. Please note this may take some time")

	if c.ConfigFile == "" {
		err = c.createWithName()
		if err != nil {
			err = fmt.Errorf("error creation cluster with name: %w", err)
			slog.Error(err.Error())
			return err
		}
	} else {
		// validate cluster configuration
		if err = c.Validate(ctx); err != nil {
			err = fmt.Errorf("an error occurred while validating cluster configuration: %w", err)
			slog.Error(err.Error())
			return err
		}

		err = c.createWithConfig()
		if err != nil {
			err = fmt.Errorf("error creating cluster with config: %w", err)
			slog.Error(err.Error())
			return err
		}

		slog.Info("Deploying user applications to cluster")
		clientset, err := createKubernetesClient()
		if err != nil {
			err = fmt.Errorf("error setting up clientset: %w", err)
			slog.Error(err.Error())
			return err
		}

		err = c.DeployApplications(ctx, clientset)
		if err != nil {
			err = fmt.Errorf("error deploying applications: %w", err)
			slog.Error(err.Error())
			return err
		}
	}

	slog.Info("Applying default resources to cluster")
	kubeContext := kindPrefix + c.Name
	err = charts.ApplyDefaultResources(ctx, kubeContext)
	if err != nil {
		err = fmt.Errorf("error while applying default resources: %w", err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Cluster has been successfully created")
	return nil
}

// create a cluster given a cluster name
func (c *Cluster) createWithName() error {
	provider := kind_cluster.NewProvider()
	err := provider.Create(c.Name)
	if err != nil {
		err = fmt.Errorf("failed to created cluster: %w", err)
		slog.Error(err.Error())
		return err
	}

	return nil
}

// create a cluster given a config file
func (c *Cluster) createWithConfig() error {
	kindConfig, err := c.generateKindConfig()
	if err != nil {
		err = fmt.Errorf("an error occurred while parsing kind config: %w", err)
		slog.Error(err.Error())
		return err
	}

	options := kind_cluster.CreateWithRawConfig([]byte(kindConfig))
	provider := kind_cluster.NewProvider()
	err = provider.Create(c.Name, options)
	if err != nil {
		err = fmt.Errorf("failed to create cluster from config: %w", err)
		slog.Error(err.Error())
		return err
	}

	return nil
}
