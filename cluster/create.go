package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skillet/skillet-kind/charts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Cluster) Create(ctx context.Context) error {
	// get cluster name if using config
	clusterConfig, err := c.parseConfig()
	if err != nil {
		err = fmt.Errorf("error parsing cluster config: %w", err)
		fmt.Println(err)
		return err
	}

	if clusterConfig.Name != "" {
		c.Name = clusterConfig.Name
	}

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

	fmt.Println("Starting creation of cluster. Please note this may take time")
	if c.ConfigFile == "" {
		err = c.createWithName(ctx)
	} else {
		err = c.createWithConfig(ctx)
	}
	if err != nil {
		err = fmt.Errorf("error creating cluster: %w", err)
		fmt.Println(err)
		return err
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
	command := exec.Command("kind", "create", "cluster", "--name", c.Name)
	err := command.Run()
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

	command := exec.Command("kind", "create", "cluster", "--config", clusterConfigFilePath, "--name", c.Name)
	err = command.Run()
	if err != nil {
		err = fmt.Errorf("failed to create cluster from config: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}
