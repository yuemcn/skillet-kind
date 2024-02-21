package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	kindPrefix            = "kind-"
	kindConfigKind        = "Cluster"
	kindApiVersion        = "kind.x-k8s.io/v1alpha4"
	controlPlaneRole      = "control-plane"
	workerRole            = "worker"
	clusterConfigFilePath = "cluster-config.yaml"
)

type Cluster struct {
	Name       string
	ConfigFile string
}

type ClusterConfig struct {
	Name         string        `yaml:"name"`
	Nodes        NodesConfig   `yaml:"nodes"`
	Applications []Application `yaml:"applications"`
}

type NodesConfig struct {
	ControlPlane int `yaml:"control_plane"`
	Worker       int `yaml:"worker"`
}

type Application struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Replicas  int    `yaml:"replicas"`
	Image     string `yaml:"image"`
}

type KindConfig struct {
	Kind       string     `yaml:"kind"`
	ApiVersion string     `yaml:"apiVersion"`
	Nodes      []KindNode `yaml:"nodes"`
}

type KindNode struct {
	Role string `yaml:"role"`
}

func NewCluster(name string, config string) Cluster {
	return Cluster{
		Name:       name,
		ConfigFile: config,
	}
}

// checks if a cluster exists or not
func (c *Cluster) clusterExists() (bool, error) {
	kubeContext := kindPrefix + c.Name

	output, err := exec.Command("kubectl", "config", "get-contexts").Output()
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		var name string
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "*" {
			name = fields[1]
		} else {
			name = fields[0]
		}

		if name == kubeContext {
			return true, nil
		}
	}

	return false, nil
}

func (c *Cluster) generateKindConfig() (string, error) {
	// generate the struct
	clusterConfig, err := c.parseConfig()
	if err != nil {
		err = fmt.Errorf("error generating kind config: %w", err)
		fmt.Println(err)
		return "", err
	}

	// add the nodes
	kindNodes := []KindNode{}
	for i := 0; i < clusterConfig.Nodes.ControlPlane; i++ {
		kindNodes = append(kindNodes, KindNode{
			Role: controlPlaneRole,
		})
	}
	for i := 0; i < clusterConfig.Nodes.Worker; i++ {
		kindNodes = append(kindNodes, KindNode{
			Role: workerRole,
		})
	}

	// put the info into a KindConfig
	kindConfig := &KindConfig{
		Kind:       kindConfigKind,
		ApiVersion: kindApiVersion,
		Nodes:      kindNodes,
	}

	// Marshal the KindConfig
	config, err := yaml.Marshal(kindConfig)
	if err != nil {
		err = fmt.Errorf("error marshalling kind config: %w", err)
		fmt.Println(err)
		return "", err
	}

	return string(config), nil
}

// parses a cluster's config file into a ClusterConfig struct
func (c *Cluster) parseConfig() (*ClusterConfig, error) {
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		err = fmt.Errorf("an error occurred while reading cluster config: %w", err)
		fmt.Println(err)
		return nil, err
	}

	var config ClusterConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		err = fmt.Errorf("an error occurred while parsing cluster config: %w", err)
		fmt.Println(err)
		return nil, err
	}

	return &config, nil
}
