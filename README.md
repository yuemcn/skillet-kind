# `skillet`

## Overview

Skillet is a command line tool that allows users to provision clusters. When a cluster is created, it comes pre-packaged with popular, essential helm charts such as Prometheus and Grafana.

In this version, support is only offered to `kind` clusters.

## Usage

To run the program, execute the following command in the main directory:

```bash
go run . [commands] [options]
```

### Create a Cluster

#### Create a cluster configuration file

In order to create or update a cluster, a user must provide a cluster configuration YAML file. A sample cluster configuration file can be found in `examples`.The components of a cluster configuration file are as follows:
- `name`: The name to give to the cluster
- `nodes`: The number of control plane and worker nodes the cluster will have
- `applications`: A list of the applications to be deployed to the cluster (currently all applications are created as deployments)

User applications have the following fields:
- `name`: The name of the application
- `namespace`: The namespace in which the application will reside
- `replicas`: The number of deployment replicas the application will be deployed on
- `image`: The application's image

Please note that the names for clusters, application names, and application namespaces must adhere to the following convention:
- name must only contain lowercase letters, numbers, and hyphens
- name must start with a lowercase letter
- name must only end with a lowercase letter or number


#### Create the cluster

To create a cluster with no custom configuration, run the following command:

```bash
go run . create --name $CLUSTER_NAME
```

To create a cluster from a specified cluster configuration file, run the following command:

```bash
go run . create --file $CONFIG_FILE_PATH
```

### Delete a Cluster

To delete a cluster, run the following command:

```bash
go run . delete --name $CLUSTER_NAME
```

## Default Resources

When a cluster is created, it comes pre-packaged with popular, essential helm charts to make it production ready. The following is a list of resources that are deployed upon cluster creation:
- Prometheus
- Grafana
