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
✔ Cloud Provider: BareMetal
✔ Cluster Name: dev-cluster
✔ Kubernetes Version: v1.8.10
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 3
✔ Hostname prefix: worker-host
✔ SSH User: ubuntu
✔ Bastion Host: None
✔ Key Path: ~/.ssh/id_rsa
✔ Host/IP for worker-host-1: 10.152.75.1
✔ Host/IP for worker-host-2: 10.152.75.2
✔ Host/IP for worker-host-3: 10.152.75.3
3 nodes added: worker-host-1, worker-host-2, worker-host-3
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 3
✔ Hostname prefix: etcd-host
✔ SSH User: ubuntu
✔ Bastion Host: None
✔ Key Path: ~/.ssh/id_rsa
✔ Host/IP for etcd-host-1: 10.133.10.1
✔ Host/IP for etcd-host-2: 10.133.10.1?2
✔ Host/IP for etcd-host-3: 10.133.10.3
3 nodes added: etcd-host-1, etcd-host-2, etcd-host-3
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 3
✔ Hostname prefix: control-host
✔ SSH User: ubuntu
✔ Bastion Host: None
✔ Key Path: ~/.ssh/id_rsa
✔ Host/IP for control-host-1: 10.133.10.5
✔ Host/IP for control-host-2: 10.133.10.6
✔ Host/IP for control-host-3: 10.133.10.7
3 nodes added: control-host-1, control-host-2, control-host-3
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


`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode.To read more about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
