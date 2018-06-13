## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system. 
The `create cluster` command allows you to create dedicated nodes for the etcd, worker and control. Creating clusters require a cluster manager to be running already.

To check if cluster manager exists, run the following:

```
$ triton-kubernetes get manager
```

Below is a demo of how to create baremetal cluster:
[![asciicast](https://asciinema.org/a/HAIwMBxNDk2yylCLXk488zAal.png)](https://asciinema.org/a/HAIwMBxNDk2yylCLXk488zAal)

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


`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read more about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
