## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds.
They can be ran in high availability (HA) mode or non-HA mode.


In order to avoid service interruption, it is recommended to run the cluster managers in HA mode.

To run the cli in interactive mode, run the following:
```
$ triton-kubernetes create manager
```

`triton-kubernetes` cli can takes a configuration file (yaml) with `--config` option to run in silent mode. Below is an example yaml file to create a cluster manager in HA mode with three nodes:

```yaml
backend_provider: local
name: ha-manager
ha: true
gcm_node_count: 3
private_registry: ""
private_registry_username: ""
private_registry_password: ""
rancher_server_image: ""
rancher_agent_image: ""
triton_account: fayazg
triton_key_path: ~/.ssh/id_rsa
triton_key_id: 2c:53:bc:63:97:9e:79:3f:91:35:5e:f4:c8:23:88:37
triton_url: https://us-east-1.api.joyent.com
triton_network_names:
  - Joyent-SDC-Public
triton_image_name: ubuntu-certified-16.04
triton_image_version: 20180109
triton_ssh_user: ubuntu
master_triton_machine_package: k4-highcpu-kvm-1.75G

triton_mysql_image_name: ubuntu-certified-16.04
triton_mysql_image_version: 20170619.1
mysqldb_triton_machine_package: k4-highcpu-kvm-1.75G
```

To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).