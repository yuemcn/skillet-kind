package cluster

import (
	"context"
	"fmt"
	"os"
	"skillet/skillet-kind/charts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kind_cluster "sigs.k8s.io/kind/pkg/cluster"
)

func (c *Cluster) Create(ctx context.Context) error {

	fmt.Println("Checking if cluster exists")
	clusterExists, err := c.clusterExists()
	if err != nil {
		err = fmt.Errorf("an error occurred while checking if cluster exists: %w", err)
		fmt.Println(err)
		return err
	} else if clusterExists {
		err = status.Errorf(codes.AlreadyExists, "a cluster with name %v already exists", c.Name)
		fmt.Println(err)
		return err
	}

	fmt.Println("Starting creation of cluster. Please note this may take some time")

	if c.ConfigFile == "" {
		err = c.createWithName(ctx)
		if err != nil {
			err = fmt.Errorf("error creation cluster with name: %w", err)
			fmt.Println(err)
			return err
		}
	} else {
		err = c.createWithConfig(ctx)
		if err != nil {
			err = fmt.Errorf("error creating cluster with config: %w", err)
			fmt.Println(err)
			return err
		}

		fmt.Println("Deploying user applications to cluster")
		clientset, err := createKubernetesClient()
		if err != nil {
			err = fmt.Errorf("error setting up clientset: %w", err)
			fmt.Println(err)
			return err
		}

		err = c.DeployApplications(ctx, clientset)
		if err != nil {
			err = fmt.Errorf("error deploying applications: %w", err)
			fmt.Println(err)
			return err
		}
	}

	fmt.Println("Applying default resources to cluster")
	kubeContext := kindPrefix + c.Name
	err = charts.ApplyDefaultResources(ctx, kubeContext)
	if err != nil {
		err = fmt.Errorf("error while applying default resources: %w", err)
		fmt.Println(err)
		return err
	}

	fmt.Println("Cluster has been successfully created")
	return nil
}

// create a cluster given a cluster name
func (c *Cluster) createWithName(ctx context.Context) error {
	provider := kind_cluster.NewProvider()
	err := provider.Create(c.Name)
	if err != nil {
		err = fmt.Errorf("failed to created cluster: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}

// create a cluster given a config file
func (c *Cluster) createWithConfig(ctx context.Context) error {
	// create the cluster-config.yaml file
	_, err := os.Stat(clusterConfigFilePath)
	var file *os.File
	if os.IsNotExist(err) {
		file, err = os.Create(clusterConfigFilePath)
		if err != nil {
			err = fmt.Errorf("error creating kind config file: %w", err)
			fmt.Println(err)
			return err
		}
	}

	endFile := func() {
		file.Close()
		os.Remove(clusterConfigFilePath)
	}

	defer endFile()

	// create kind config
	kindConfig, err := c.generateKindConfig()
	if err != nil {
		err = fmt.Errorf("an error occurred while parsing kind config: %w", err)
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(clusterConfigFilePath, []byte(kindConfig), 0666)
	if err != nil {
		err = fmt.Errorf("error writing kind configuration to file: %w", err)
		fmt.Println(err)
		return err
	}

	options := kind_cluster.CreateWithConfigFile(clusterConfigFilePath)
	provider := kind_cluster.NewProvider()
	err = provider.Create(c.Name, options)
	if err != nil {
		err = fmt.Errorf("failed to create cluster from config: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}
