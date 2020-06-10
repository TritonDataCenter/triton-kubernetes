## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system.
Create cluster command allows to create dedicated nodes for the etcd, worker and control. Creating clusters require a cluster manager to be running already.

To check if cluster manager exists, run the following:

```
$ triton-kubernetes get manager
```

To create cluster, run the following:

```
triton-kubernetes create cluster
✔ Backend Provider: Local
create cluster called
✔ Cluster Manager: dev-manager
✔ Cloud Provider: Azure
✔ Cluster Name: azure-cluster
✔ Kubernetes Version: v1.17.6
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
✔ Azure Subscription ID: 0535d7cf-a52e-491b-b7bc-37f674787ab8
✔ Azure Client ID: 22520959-c5bb-499a-b3d0-f97e8849385e
✔ Azure Client Secret: a19ed50f-f7c1-4ef4-9862-97bc880d2536
✔ Azure Tenant ID: 324e4a5e-53a9-4be4-a3a5-fcd3e79f2c5b
✔ Azure Environment: public
✔ Azure Location: West US
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 3
✔ Hostname prefix: azure-node-worker
✔ Azure Size: Standard_B1ms
✔ Azure SSH User: ubuntu
✔ Azure Public Key Path: ~/.ssh/id_rsa.pub
  Disk created? No
3 nodes added: azure-node-worker-1, azure-node-worker-2, azure-node-worker-3
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 3
✔ Hostname prefix: azure-node-etcd
✔ Azure Size: Standard_B1ms
✔ Azure SSH User: ubuntu
✔ Azure Public Key Path: ~/.ssh/id_rsa.pub
  Disk created? No
3 nodes added: azure-node-etcd-1, azure-node-etcd-2, azure-node-etcd-3
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 3
✔ Hostname prefix: azure-node-contol
✔ Azure Size: Standard_B1ms
✔ Azure SSH User: ubuntu
✔ Azure Public Key Path: ~/.ssh/id_rsa.pub
  Disk created? No
3 nodes added: azure-node-contol-1, azure-node-contol-2, azure-node-contol-3
  Create new node? No
  Proceed? Yes
```


To destroy cluster, run the following:

```
$ triton-kubernetes destroy cluster
✔ Backend Provider: Local
✔ Cluster Manager: dev-manager
✔ Cluster: dev-cluster
  Destroy "dev-cluster"? Yes
```

To get cluster, run the following:

```
$ triton-kubernetes get cluster
```


`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
