## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system.
Create cluster command allows to create dedicated nodes for the etcd, worker and control. Creating clusters require a cluster manager to be running already.

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
✔ Cloud Provider: GCP
✔ Cluster Name: gcp-cluster
✔ Kubernetes Version: v1.17.6
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
✔ Path to Google Cloud Platform Credentials File: ~/gcp.json
✔ GCP Compute Region: us-east1
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 3
✔ Hostname prefix: gcp-nodew
✔ GCP Instance Zone: us-east1-c
✔ GCP Machine Type: n1-standard-1
✔ GCP Image: ubuntu-1604-xenial-v20180424
3 nodes added: gcp-nodew-1, gcp-nodew-2, gcp-nodew-3
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 3
✔ Hostname prefix: gcp-nodee
✔ GCP Instance Zone: us-east1-b
✔ GCP Machine Type: n1-standard-1
✔ GCP Image: ubuntu-1604-xenial-v20180424
3 nodes added: gcp-nodee-1, gcp-nodee-2, gcp-nodee-3
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 3
✔ Hostname prefix: gcp-nodec
✔ GCP Instance Zone: us-east1-d
✔ GCP Machine Type: n1-standard-1
✔ GCP Image: ubuntu-1604-xenial-v20180424
3 nodes added: gcp-nodec-1, gcp-nodec-2, gcp-nodec-3
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
