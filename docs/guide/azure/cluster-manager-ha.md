## Cluster Manager

Cluster Managers can manage multiple clusters across regions/data-centers and/or clouds. They can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.

Creating a highly available cluster manager requires a TLS certificate before you run triton-kubernetes.

You can create a TLS certificate by leveraging [certbot](https://certbot.eff.org/docs/install.html).
```
sudo certbot certonly --manual -d test.example.com --preferred-challenge dns
```
Or you may provide your own certificate and key. The only requirement is for both to be PEM encoded.

To create a cluster manager, run the following:
```
$ triton-kubernetes create manager
✔ Backend Provider: Local
✔ Cloud Provider: Azure
✔ Cluster Manager Name: dev-manager
  Highly Available? Yes
✔ Fully Qualified Domain Name: test.example.com
✔ TLS Private Key Path: ~/Documents/privkey.pem
✔ TLS Certificate Path: ~/Documents/fullchain.pem
✔ Private Registry: None
✔ Rancher Server Image: Default
✔ Rancher Agent Image: Default
✔ Set UI Admin Password: *****
✔ Azure Subscription ID: 0535d7cf-a52e-491b-b7bc-37f674787ab8
✔ Azure Client ID: 22520959-c5bb-499a-b3d0-f97e8849385e
✔ Azure Client Secret: a19ed50f-f7c1-4ef4-9862-97bc880d2536
✔ Azure Tenant ID: 324e4a5e-53a9-4be4-a3a5-fcd3e79f2c5b
✔ Azure Environment: public
✔ Azure Location: West US
✔ Azure Size: Standard_B1ms
✔ Azure SSH User: ubuntu
✔ Azure Public Key Path: ~/.ssh/id_rsa.pub
✔ Azure Private Key Path: ~/.ssh/id_rsa
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

`triton-kubernetes` cli can takes a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/joyent/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).