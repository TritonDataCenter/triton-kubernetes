## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds. They can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.

To create cluster manager, run the following:
```
$ triton-kubernetes create manager
✔ Backend Provider: Local
✔ Cloud Provider: GCP
✔ Cluster Manager Name: dev-manager
✔ Private Registry: None
✔ Rancher Server Image: Default
✔ Rancher Agent Image: Default
✔ Set UI Admin Password: *****
✔ Path to Google Cloud Platform Credentials File: ~/gcp.json
✔ GCP Compute Region: us-east1
✔ GCP Instance Zone: us-east1-c
✔ GCP Machine Type: n1-standard-1
✔ GCP Image: ubuntu-1604-xenial-v20180424
✔ GCP Public Key Path: ~/.ssh/id_rsa.pub
✔ GCP Private Key Path: ~/.ssh/id_rsa
✔ GCP SSH User: root
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

`triton-kubernetes` cli can takes a configuration file (yaml) with `--config` option to run in silent mode.To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).