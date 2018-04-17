## Clusters

Creating clusters require a cluster manager to be running already.

A cluster can be ran in HA or non-HA mode. In HA mode, etcd and kubernetes components will run in their own dedicated nodes. In a non-HA cluster, etcd and all other services will run on the compute nodes next to your Kubernetes deployments.

Below is an example yaml file which can be passed to the `triton-kubernetes` cli to create an HA Kubernetes Cluster (with 3 etcd nodes, 3 nodes for Kubernetes services and 4 nodes for deployments) to the already running `ha-manager` cluster manager.

```yaml
backend_provider: local
cluster_manager: ha-manager
name: ha-cluster
cluster_cloud_provider: triton
k8s_plane_isolation: required
private_registry: ""
private_registry_username: ""
private_registry_password: ""
k8s_registry: ""
k8s_registry_username: ""
k8s_registry_password: ""
triton_account: fayazg
triton_key_path: ~/.ssh/id_rsa
triton_key_id: 2c:53:bc:63:97:9e:79:3f:91:35:5e:f4:c8:23:88:37
triton_url: https://us-east-1.api.joyent.com
nodes:
  - node_count: 3
    rancher_host_label: etcd
    hostname: test-etcd
    triton_network_names:
      - Joyent-SDC-Public
    triton_image_name: ubuntu-certified-16.04
    triton_image_version: 20180109
    triton_ssh_user: ubuntu
    triton_machine_package: k4-highcpu-kvm-1.75G
  - node_count: 3
    rancher_host_label: orchestration
    hostname: test-orch
    triton_network_names:
      - Joyent-SDC-Public
    triton_image_name: ubuntu-certified-16.04
    triton_image_version: 20180109
    triton_ssh_user: ubuntu
    triton_machine_package: k4-highcpu-kvm-1.75G
  - node_count: 1
    rancher_host_label: compute
    hostname: test-compute
    triton_network_names:
      - Joyent-SDC-Public
    triton_image_name: ubuntu-certified-16.04
    triton_image_version: 20180109
    triton_ssh_user: ubuntu
    triton_machine_package: k4-highcpu-kvm-1.75G
```

To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).