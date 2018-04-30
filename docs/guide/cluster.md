## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system. 
The 'create cluster' command allows you to create dedicated nodes for the etcd, worker and control. Creating clusters require a cluster manager to be running already.

To check if cluster manager is exist, run the following:

```
$ triton-kubernetes get manager
```

To create cluster, run the following:

```
$ triton-kubernetes create cluster
✔ Backend Provider: Local
create cluster called
✔ Cluster Manager: dev-manager
✔ Cloud Provider: Triton
✔ Cluster Name: dev-cluster
✔ Kubernetes Version: v1.9.5
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
✔ Triton Account Name: [changeme]
✔ Triton Key Path: ~/.ssh/id_rsa
✔ Triton URL: https://us-east-1.api.joyent.com
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 3
✔ Hostname prefix: dev-etcd
✔ Triton Network Attached: Joyent-SDC-Public
  Attach another? No
✔ Triton Image: ubuntu-certified-16.04@20180222
✔ Triton SSH User: ubuntu
✔ Triton Machine Package: k4-highcpu-kvm-1.75G
3 nodes added: dev-etcd-1, dev-etcd-2, dev-etcd-3
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 3
✔ Hostname prefix: dev-worker
✔ Triton Network Attached: Joyent-SDC-Public
  Attach another? No
✔ Triton Image: ubuntu-certified-16.04@20180222
✔ Triton SSH User: ubuntu
✔ Triton Machine Package: k4-highcpu-kvm-1.75G
3 nodes added: dev-worker-1, dev-worker-2, dev-worker-3
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 3
✔ Hostname prefix: dev-control
✔ Triton Network Attached: Joyent-SDC-Public
  Attach another? No
✔ Triton Image: ubuntu-certified-16.04@20180222
✔ Triton SSH User: ubuntu
✔ Triton Machine Package: k4-highcpu-kvm-1.75G
3 nodes added: dev-control-1, dev-control-2, dev-control-3
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