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
✔ Cloud Provider: AWS
✔ Cluster Name: dev-cluster
✔ Kubernetes Version: v1.16.3
✔ Kubernetes Network Provider: calico
✔ Private Registry: None
✔ k8s Registry: None
✔ AWS Access Key: [changeme]
✔ AWS Secret Key: [changeme]
✔ AWS Region: us-east-1
✔ Name for new aws public key: triton-kubernetes_public_key
✔ AWS Public Key Path: ~/.ssh/id_rsa.pub
✔ AWS VPC CIDR: 10.0.0.0/16
✔ AWS Subnet CIDR: 10.0.2.0/24
  Create new node? Yes
✔ Host Type: worker
✔ Number of nodes to create: 3
✔ Hostname prefix: dev-worker
✔ AWS AMI: ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405
✔ AWS Instance Type: t2.micro
  Volume Created? Yes
✔ EBS Volume Device Name: /dev/sdf
✔ EBS Volume Mount Path: /mnt/triton-kubernetes
  EBS Volume Type? General Purpose SSD
✔ EBS Volume Size in GiB: 100
3 nodes added: dev-worker-1, dev-worker-2, dev-worker-3
  Create new node? Yes
✔ Host Type: etcd
  Number of nodes to create? 3
✔ Hostname prefix: dev-etcd
✔ AWS AMI: ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405
✔ AWS Instance Type: t2.micro
  Volume Created? No
3 nodes added: dev-etcd-1, dev-etcd-2, dev-etcd-3
  Create new node? Yes
✔ Host Type: control
  Number of nodes to create? 3
✔ Hostname prefix: dev-control
✔ AWS AMI: ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405
✔ AWS Instance Type: t2.micro
  Volume Created? No
3 nodes added: dev-control-1, dev-control-2, dev-control-3
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