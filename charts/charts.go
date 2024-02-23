package charts

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	defaultResourcesPath = "./charts/charts.yaml"
	helmDriverEnv        = "HELM_DRIVER"
)

type DefaultResources struct {
	HelmCharts []HelmChart `yaml:"helm_charts"`
}

type HelmChart struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Repo      string `yaml:"repo"`
	URL       string `yaml:"url"`
	TGZ       string `yaml:"tgz"`
}

// apply default resources
func ApplyDefaultResources(ctx context.Context, kubeContext string) error {
	slog.Info("parsing chart config")
	chartConfig, err := parseChartConfig()
	if err != nil {
		err = fmt.Errorf("error while parsing chart config: %w", err)
		slog.Error(err.Error())
		return err
	}

	slog.Info("applying helm charts")
	for _, chart := range chartConfig.HelmCharts {
		settings := cli.New()

		slog.Info("creating action config", "chart", chart.Name)
		actionConfig := new(action.Configuration)
		err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv(helmDriverEnv), log.Printf)
		if err != nil {
			err = fmt.Errorf("error initializing action for chart %s: %w", chart.Name, err)
			slog.Error(err.Error())
			return err
		}

		slog.Info("locating chart", "chart", chart.Name)
		client := action.NewInstall(actionConfig)
		chartPath, err := client.LocateChart(chart.URL, settings)
		if err != nil {
			err = fmt.Errorf("error locating chart %s: %w", chart.Name, err)
			slog.Error(err.Error())
			return err
		}

		// TODO: Fix this part so it's not hard-coded. Currently this is a workaroud
		chartPath = strings.TrimSuffix(chartPath, "helm-charts") + chart.TGZ

		slog.Info("loading chart", "chart", chart.Name)
		helmChart, err := loader.Load(chartPath)
		if err != nil {
			err = fmt.Errorf("error loading chart %s: %w", chart.Name, err)
			slog.Error(err.Error())
			return err
		}

		slog.Info("installing chart", "chart", chart.Name)
		install := action.NewInstall(actionConfig)
		install.ReleaseName = chart.Name
		install.CreateNamespace = true
		install.Namespace = chart.Namespace
		var vals map[string]interface{}
		_, err = install.RunWithContext(ctx, helmChart, vals)
		if err != nil {
			err = fmt.Errorf("error installing chart %s: %w", chart.Name, err)
			slog.Error(err.Error())
			return err
		}

		slog.Info("Successfully installed chart", "chart", chart.Name)
	}

	slog.Info("Successfully installed all charts")
	return nil
}

// parse chart config into struct
func parseChartConfig() (*DefaultResources, error) {
	data, err := os.ReadFile(defaultResourcesPath)
	if err != nil {
		err = fmt.Errorf("an error occurred while reading chart config: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	var config DefaultResources
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		err = fmt.Errorf("an error occurred while parsing chart config: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &config, nil
}
