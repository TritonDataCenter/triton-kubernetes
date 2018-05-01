## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds. They can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.

To create cluster manager, run the following:
```
$ triton-kubernetes create manager
✔ Backend Provider: Local
✔ Cloud Provider: AWS
✔ Cluster Manager Name: dev-manager
✔ Private Registry: None
✔ Rancher Server Image: Default
✔ Rancher Agent Image: Default
✔ Set UI Admin Password: ****
✔ AWS Access Key: [changeme]
✔ AWS Secret Key: [changeme]
✔ AWS Region: us-east-1
✔ Name for new aws public key: triton-kubernetes_public_key
✔ AWS Public Key Path: ~/.ssh/id_rsa.pub
✔ AWS Private Key Path: ~/.ssh/id_rsa
✔ AWS SSH User: ubuntu
✔ AWS VPC CIDR: 10.0.0.0/16
✔ AWS Subnet CIDR: 10.0.2.0/24
✔ AWS AMI: ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405
✔ AWS Instance Type: t2.micro
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

`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode.To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).