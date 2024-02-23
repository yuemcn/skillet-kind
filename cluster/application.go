package cluster

import (
	"context"
	"fmt"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Application struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Replicas  int    `yaml:"replicas"`
	Image     string `yaml:"image"`
}

func (c *Cluster) DeployApplications(ctx context.Context, clientset *kubernetes.Clientset) error {
	clusterConfig, err := c.parseConfig()
	if err != nil {
		err = fmt.Errorf("error parsing cluster config: %w", err)
		slog.Error(err.Error())
		return err
	}

	// read applications from cluster config
	for _, app := range clusterConfig.Applications {
		slog.Info("Deploying application", "application", app.Name)
		app.ApplyDeployment(ctx, clientset)
		slog.Info("Successfully deployed application", "application", app.Name)
	}

	slog.Info("Successfully deployed all applications to cluster")
	return nil
}

// deploy an application
func (a *Application) ApplyDeployment(ctx context.Context, clientset *kubernetes.Clientset) error {
	err := a.CreateNamespace(ctx, clientset)
	if err != nil {
		err = fmt.Errorf("error creating namespace for application %s: %w", a.Name, err)
		slog.Error(err.Error())
		return err
	}

	err = a.CreateDeployment(ctx, clientset)
	if err != nil {
		err = fmt.Errorf("error creating deployment for application %s: %w", a.Name, err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Successfully applied deployment", "application", a.Name)
	return nil
}

func (a *Application) CreateDeployment(ctx context.Context, clientset *kubernetes.Clientset) error {
	slog.Info("Creating deployment", "application", a.Name)

	deploymentsClient := clientset.AppsV1().Deployments(a.Namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: a.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(a.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": a.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": a.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  a.Name,
							Image: a.Image,
						},
					},
				},
			},
		},
	}

	_, err := deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		err = fmt.Errorf("error creating deployment %s: %w", a.Name, err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Successfully created deployment", "application", a.Name)
	return nil
}

func (a *Application) CreateNamespace(ctx context.Context, clientset *kubernetes.Clientset) error {
	// check if namespace exists
	_, err := clientset.CoreV1().Namespaces().Get(ctx, a.Namespace, metav1.GetOptions{})
	if err == nil {
		slog.Info("namespace already exists. Skipping creation")
		return nil
	}

	// create namespace
	slog.Info("creating namespace", "namespace", a.Namespace)
	namespaceSpec := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: a.Namespace,
		},
	}

	ns, err := clientset.CoreV1().Namespaces().Create(ctx, namespaceSpec, metav1.CreateOptions{})
	if err != nil {
		err = fmt.Errorf("error creating namespace client for application %s: %w", a.Name, err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("Successfully created namespace", "namespace", ns.Name)
	return nil
}

func int32Ptr(i int) *int32 {
	j := int32(i)
	return &j
}
