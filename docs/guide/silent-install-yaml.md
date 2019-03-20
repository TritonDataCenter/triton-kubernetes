# Introduction

Triton Multi-Cloud Kubernetes cli can be invoked in interactive or silent modes. In silent mode, the cli takes in a yaml configuration file which contains different parameters depending on what is being run. Standing up a cluster manager in HA mode takes in different arguments than non-HA mode and the same goes for kubernetes cluster creation.

For sample YAML files, look under [examples/silent-install](https://github.com/joyent/triton-kubernetes/tree/master/examples/silent-install).

## Cluster Manager YAML

Before creating a Kubernetes cluster, we need to have a running cluster manager. The parameters for cluster manager are:

| Parameter        | Description  |
| ------------- |:-----|
| `backend_provider` | Where/how to store the configuration for this cluster manager and clusters it manages. Options are `manta` or `local`. |
| `triton_account` `triton_key_path` `triton_url` `manta_url` | If using `manta` as a `backend_provider`, these parameters need to be provided. |
| `name` | Name of this cluster manager |
| `private_registry` | URL of the private registry that includes rancher containers |
| `private_registry_username` | Username for the private registry |
| `private_registry_password` | Password for the private registry |
| `rancher_server_image` | URL for the rancher/server container within the private registry |
| `rancher_agent_image` | URL for the rancher/agent container within the private registry |
| `triton_account` | Triton account name |
| `triton_key_path` | SSH key path for the `triton_account` |
| `triton_url` | Triton API URL |
| `triton_network_names` | List of Triton network names that are available to `triton_account` in `triton_url` data-center.
| `triton_image_name` | Triton image to use for the cluster manager. Must be available in the selected data-center for the user. |
| `triton_image_version` | Triton image version to use for the image `triton_image_name`. |
| `triton_ssh_user` | Default SSH user available for the selected image. NOTE: For Ubuntu images, default SSH user is `ubuntu`. |
| `master_triton_machine_package` | Triton KVM package to use for the cluster manager. |
| `rancher_admin_password` | UI password for admin user |

## Cluster YAML

YAML parameters for cluster are:

| Parameter        | Description  |
| ------------- |:-----|
| `backend_provider` | Where/how to store the configuration for this cluster manager and clusters it manages. Options are `manta` or `local`. |
| `cluster_manager` | Which cluster manager should manage this new cluster that is going to be created. |
| `cluster_cloud_provider` | Which cloud should the cluster run on. Options are `triton`, `aws`, `gcp`, or `azure`. |
| `name` | Cluster name |
| `k8s_version` | Version of Kubernetes to deploy for this cluster. Available versions are: `v1.8.11-rancher2-1`, `v1.9.7-rancher2-2`, `v1.10.3-rancher2-1`, `v1.11.8-rancher1-1`, `v1.12.6-rancher1-1`, and `v1.13.4-rancher1-1``. |
| `k8s_network_provider` | Network stack to use for this Kubernetes cluster. Available options are: `calico` and `flannel`. |
| `private_registry` | URL of the private registry that includes rancher containers |
| `private_registry_username` | Username for the private registry |
| `private_registry_password` | Password for the private registry |
| `k8s_registry` | URL of the private registry that includes rancher containers and `gcr.io` containers needed. |
| `k8s_registry_username` | Username for the private registry |
| `k8s_registry_password` | Password for the private registry |
| `nodes` | Parameters needed for the different type of nodes that should be created for this cluster. |

For examples, look in [examples/silent-install](https://github.com/joyent/triton-kubernetes/tree/master/examples/silent-install).

> Note: Spreading a cluster across multiple clouds could cause performance issues.