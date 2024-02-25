package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"unicode"

	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Cluster) Validate(ctx context.Context) error {
	config, err := c.parseConfig()
	if err != nil {
		err = fmt.Errorf("error parsing cluster config, command: Validate, cluster: %s, error: %w", c.Name, err)
		slog.Error(err.Error())
		return err
	}

	if config.Name == "" {
		err = status.Errorf(codes.FailedPrecondition, "a cluster name must be specified")
		slog.Error(err.Error())
		return err
	}

	if err := followsNamingConvention(config.Name); err != nil {
		err = fmt.Errorf("cluster name does not follow naming convention: %w", err)
		slog.Error(err.Error())
		return err
	}

	// validate control plane nodes is at least 1
	if config.Nodes.ControlPlane < 1 {
		err = status.Errorf(codes.FailedPrecondition, "the number of control plane nodes must be greater than 0")
		slog.Error(err.Error())
		return err
	}

	// validate worker nodes is not negative
	if config.Nodes.Worker < 0 {
		err = status.Errorf(codes.FailedPrecondition, "the number of worker nodes cannot be negative")
		slog.Error(err.Error())
		return err
	}

	// validate name, namespace, replicas, and image are specified
	for _, app := range config.Applications {
		if app.Name == "" {
			err = status.Errorf(codes.FailedPrecondition, "application name must not be empty")
			slog.Error(err.Error())
			return err
		}

		if err := followsNamingConvention(app.Name); err != nil {
			err = fmt.Errorf("application name does not follow naming convention: %w", err)
			slog.Error(err.Error())
			return err
		}

		if app.Namespace == "" {
			err = status.Errorf(codes.FailedPrecondition, "application namespace must not be empty")
			slog.Error(err.Error())
			return err
		}

		if err := followsNamingConvention(app.Namespace); err != nil {
			err = fmt.Errorf("application namespace does not follow naming convention: %w", err)
			slog.Error(err.Error())
			return err
		}

		if app.Replicas < 1 {
			err = status.Errorf(codes.FailedPrecondition, "application replicas must be at least 1")
			slog.Error(err.Error())
			return err
		}

		if app.Image == "" {
			err = status.Errorf(codes.FailedPrecondition, "application image must not be empty")
			slog.Error(err.Error())
			return err
		}

		if err := imageExists(ctx, app.Image); err != nil {
			err = fmt.Errorf("error retreiving image %s: %w", app.Image, err)
			slog.Error(err.Error())
			return err
		}
	}

	slog.Info("Cluster configuration is valid")
	return nil
}

func followsNamingConvention(s string) error {
	// only contains letters, numbers, and hyphens
	if !regexp.MustCompile(`^[a-z0-9-]*$`).MatchString(s) {
		return status.Errorf(codes.FailedPrecondition, "only lowercase letters, numbers, and hyphens are allowed")
	}

	// starts with lowercase letter
	if firstChar := rune(s[0]); !unicode.IsLower(firstChar) {
		return status.Errorf(codes.FailedPrecondition, "first char must be a lowercase letter")
	}

	// does not end with a hyphen
	if lastChar := rune(s[len(s)-1]); lastChar == '-' {
		return status.Errorf(codes.FailedPrecondition, "last char must be lowercase letter or number")
	}

	return nil
}

func imageExists(ctx context.Context, image string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("error creating docker client: %w", err)
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return status.Errorf(codes.NotFound, "image does not exist")
	}

	return nil
}
