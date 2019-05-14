## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds. They can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.

To create a cluster manager, run the following:
```
$ triton-kubernetes create manager
✔ Backend Provider: Local
✔ Cloud Provider: BareMetal
✔ Cluster Manager Name: test
✔ Private Registry: None
✔ Rancher Server Image: Default
✔ Rancher Agent Image: Default
✔ Set UI Admin Password: ****
✔ Host/IP for cluster manager: 10.25.65.44
✔ SSH User: ubuntu
✔ Bastion Host: None
✔ Key Path: ~/.ssh/id_rsa
  Proceed? Yes
```

To destroy cluster manager, run the following:

```
$ triton-kubernetes destroy manager
✔ Backend Provider: Local
✔ Cluster Manager: dev-manager
  Destroy "dev-manager"? Yes
```
> Note: Destorying cluster manager will destroy all your clusters and nodes attached to the cluster manager.

To get cluster manager, run the following:

```
$ triton-kubernetes get manager
```

`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/mesoform/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
