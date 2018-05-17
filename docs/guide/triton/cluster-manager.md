## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds. They can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.

To create a cluster manager, run the following:
```
$ triton-kubernetes create manager
✔ Backend Provider: Local
✔ Cloud Provider: Triton
✔ Cluster Manager Name: dev-manager
✔ Private Registry: None
✔ Rancher Server Image: Default
✔ Rancher Agent Image: Default
✔ Set UI Admin Password: ****
✔ Triton Account Name: [changeme]
✔ Triton Key Path: ~/.ssh/id_rsa
✔ Triton URL: https://us-east-1.api.joyent.com
✔ Triton Networks: Joyent-SDC-Public
  Attach another? No
✔ Triton Image: ubuntu-certified-16.04@20180222
✔ Triton SSH User: ubuntu
✔ Rancher Master Triton Machine Package: k4-highcpu-kvm-1.75G
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

`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
