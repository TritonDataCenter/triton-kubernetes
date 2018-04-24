# Introduction

Triton Multi-cloud Kubernetes cli can be invoked in interactive or silent modes. In silent mode, the cli takes in a yaml configuration file which contains different parameters depending on what is being ran. Standing up a cluster manager in HA mode takes in different arguments than non-HA mode and the same goes for kubernetes cluster creation.

For sample YAML files, look under [examples/silent-install](https://github.com/joyent/triton-kubernetes/tree/master/examples/silent-install).

## Cluster Manager YAML

Before creating a Kubernetes cluster, we need to have a running cluster manager. The parameters for cluster manager are:

| Parameter        | Description  |
| ------------- |:-----|
| `backend_provider` | Where/how to store the configuration for this cluster manager and clusters it manages. Options are `manta` or `local`. |
| `name` | Name of this cluster manager |
| `ha` | Weather this cluster manager should run in HA mode or not. Options are `true` or `false`. |
| `gcm_node_count` | How many cluster manager nodes should be set up in HA mode. If `ha=false`, this parameter is ignored. |
| `private_registry` | URL of the private registry that includes rancher containers |
| `private_registry_username` | Username for the private registry |
| `private_registry_password` | Password for the private registry |
| `rancher_server_image` | URL for the rancher/server container within the private registry |
| `rancher_agent_image` | URL for the rancher/agent container within the private registry |
| `triton_account` | Triton account name |
| `triton_key_path` | SSH key path for the `triton_account` |
| `triton_key_id` | SSH key fingerprint for the `triton_key_path` file |
| `triton_url` | Triton API URL |
| `triton_network_names` | List of Triton network names that are available to `triton_account` in `triton_url` data-center.
| `triton_image_name` | Triton image to use for the cluster manager. Must be available in the selected data-center for the user. |
| `triton_image_version` | Triton image version to use for the image `triton_image_name`. |
| `triton_ssh_user` | Default SSH user available for the selected image. NOTE: Ubuntu images default SSH user is `ubuntu`. |
| `master_triton_machine_package` | Triton KVM package to use for the cluster managers. |
| `triton_mysql_image_name` | Triton image to use for the shared mysqldb of the HA cluster manager. This parameter will be ignored if `ha=false`. |
| `triton_mysql_image_version` | Triton image version to use for the image `triton_mysql_image_name`. This parameter will be ignored if `ha=false`. |
| `mysqldb_triton_machine_package` | Default SSH user available for the `triton_mysql_image_name` image. This parameter will be ignored if `ha=false`. NOTE: Ubuntu images default SSH user is `ubuntu`. |

## Cluster YAML

YAML parameters for cluster are:

| Parameter        | Description  |
| ------------- |:-----|
| `backend_provider` | Where/how to store the configuration for this cluster manager and clusters it manages. Options are `manta` or `local`. |
| `cluster_manager` | Which cluster manager should manage this new cluster that is going to be created. |
| `name` | Cluster name |
| `cluster_cloud_provider` | Which cloud should the cluster run on. Options are `triton`, `aws`, `gcp`, or `azure`. |
| `k8s_plane_isolation` | Should this cluster run in HA mode. Options are `required` and `none`. If `k8s_plane_isolation=required`, etcd and kubernetes services will run in dedicated nodes. |
| `private_registry` | URL of the private registry that includes rancher containers |
| `private_registry_username` | Username for the private registry |
| `private_registry_password` | Password for the private registry |
| `k8s_registry` | URL of the private registry that includes rancher containers and `gcr.io` containers needed. |
| `k8s_registry_username` | Username for the private registry |
| `k8s_registry_password` | Password for the private registry |
| `triton_account` | Triton account name |
| `triton_key_path` | SSH key path for the `triton_account` |
| `triton_key_id` | SSH key fingerprint for the `triton_key_path` file |
| `triton_url` | Triton API URL |
| `nodes` | Parameters needed for the different type of nodes that should be created for this cluster. |