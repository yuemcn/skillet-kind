package cluster

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
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
	if c.Name == "" {
		clusterConfig, err := c.parseConfig()
		if err != nil {
			err = fmt.Errorf("error occurred while parsing config: %w", err)
			fmt.Println(err)
			return false, err
		}

		c.Name = clusterConfig.Name
	}

	kubeContext := kindPrefix + c.Name

	contexts, err := listContexts()
	if err != nil {
		err = fmt.Errorf("error listing contexts: %w", err)
		fmt.Println(err)
		return false, err
	}

	for _, context := range contexts {
		if context.Cluster == kubeContext {
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

func createKubernetesClient() (*kubernetes.Clientset, error) {
	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		err = fmt.Errorf("error getting kubeconfig path: %w", err)
		fmt.Println(err)
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		err = fmt.Errorf("error building configs for kubeconfig: %w", err)
		fmt.Println(err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = fmt.Errorf("error creating clientset for config: %w", err)
		fmt.Println(err)
		return nil, err
	}

	return clientset, nil
}

func getKubeconfigPath() (string, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return *kubeconfig, nil
}

func listContexts() (map[string]*api.Context, error) {
	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides)

	config, err := kubeconfig.RawConfig()
	if err != nil {
		err = fmt.Errorf("error getting raw kubeconfig: %w", err)
		fmt.Println(err)
		return nil, err
	}

	return config.Contexts, nil
}
