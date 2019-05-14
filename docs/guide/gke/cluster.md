## Cluster

A cluster is a group of physical (or virtual) computers that share resources to accomplish tasks as if they were a single system. Creating clusters require a cluster manager to be running already.

To check if cluster manager exists, run the following:

```
$ triton-kubernetes get manager
```

> <sub>NOTE: GCP Service Account must have the following permissions:</sub>
> - <sub>compute.regions.list</sub>
> - <sub>iam.serviceAccountActor</sub>
> - <sub>container.clusterRoleBindings.create</sub>

To create a cluster on GKE, run the following:

```
$ triton-kubernetes create cluster
✔ Backend Provider: Local
✔ Cluster Manager: dev-manager
✔ Cloud Provider: GKE
✔ Cluster Name: dev-cluster
✔ Path to Google Cloud Platform Credentials File: /Users/fayazg/fayazg-5b46508599f1.json
✔ GCP Compute Region: us-central1
✔ GCP Zone: us-central1-a
✔ GCP Additional Zones: us-central1-b
  Add another? No
✔ GCP Machine Type: n1-standard-1
✔ Kubernetes Version: v1.9.7
✔ Number of nodes to create: 3
✔ Kubernetes Master Password: ***************************
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


`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/mesoform/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).