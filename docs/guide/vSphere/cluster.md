## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system.
The `create cluster` command allows you to create dedicated nodes for the etcd, worker and control. Creating clusters require a cluster manager to be running already.

To check if cluster manager exists, run the following:

```
$ triton-kubernetes get manager
```

To create cluster, run the following:

```
$ triton-kubernetes create cluster
✔ Backend Provider: Local
create cluster called
✔ Cluster Manager: dev-manager
✔ Cloud Provider: vSphere
✔ Cluster Name: dev-cluster
✔ Kubernetes Version: v1.8.10
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
✔ vSphere User: [changeme]
✔ vSphere Password: ****
✔ vSphere Server: [changeme]
✔ vSphere Datacenter Name: [changeme]
✔ vSphere Datastore Name: [changeme]
✔ vSphere Resource Pool Name: [changeme]
✔ vSphere Network Name: [changeme]
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 1
✔ Hostname prefix: worker-host
✔ VM Template Name: vm-worker-temp
✔ SSH User: [changeme]
✔ Private Key Path: ~/.ssh/id_rsa
1 node added: worker-host-1
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 1
✔ Hostname prefix: etcd-host
✔ VM Template Name: vm-etcd-temp
✔ SSH User: [changeme]
✔ Private Key Path: ~/.ssh/id_rsa
1 node added: etcd-host-1
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 1
✔ Hostname prefix: control-host
✔ VM Template Name: vm-control-temp
✔ SSH User: [changeme]
✔ Private Key Path: ~/.ssh/id_rsa
1 node added: control-host-1
  Create new node? No
  Proceed? Yes
```
To destroy cluster , run the following:

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


`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read more about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
