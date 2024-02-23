# `skillet-kind` Examples

This directory contains an example of a basic `cluster.yaml` for configuring a `skillet` cluster.

## Cluster Configuration File

In order to create a `skillet` cluster, a configuration file in the form of a YAML file is requried. This directory contains an example of such a file. Note that the configuration file contains the following:

- `name`: The name to give to the cluster
- `nodes`: The number of control plane and worker nodes the cluster will have
- `applications`: A list of the applications to be deployed to the cluster (currently all applications are created as deployments)

Please note that applications have the following fields:
- `name`: The name of the application
- `namespace`: The namespace in which the application will reside
- `replicas`: The number of deployment replicas the application will be deployed on
- `image`: The application's image
