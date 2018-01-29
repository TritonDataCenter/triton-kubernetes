<a>
  <img src="https://www.joyent.com/assets/img/external/triton-kubernetes.svg" width="100%" height="144">
</a>

Triton Kubernetes is a multi-cloud Kubernetes solution. It has a global cluster manager which will run on Triton and manages Kubernetes environments. This cluster manager will manage environments running on any region of any supported cloud. For an example set up, look at the [How-To](#how-to) section.

![Triton-Kubernetes](docs/imgs/Triton-Kubernetes.png)

> NOTE: This package has been tested on Linux/OSX.

## Quick Start Guide

### Pre-Reqs
In order to run **Triton Kubernetes**, you must create a [Triton](https://my.joyent.com/) account and install [`jq`](#install-jq) and [`terraform`](#install-terraform).

[Triton](https://www.joyent.com/why) is our container-native and open source cloud, which we will use to provide the infrastructure required for your Kubernetes cluster.

[jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor. It's leveraged by `triton-kubernetes`.

[Terraform](https://www.terraform.io/) enables you to safely and predictably create, change, and improve production infrastructure. It is an open source tool that codifies APIs into declarative configuration files that can be shared amongst team members, treated as code, edited, reviewed, and versioned.

#### Install `jq`

Install `jq` for the system you are on:

```bash
# OS X using brew
brew install jq

# Debian/Ubuntu
apt-get install jq

# CentOS/RHEL
yum install jq
```

#### Install Terraform

Install `terraform` for the system you are on:
```bash
# OS X using brew
brew install terraform

# Debian/Ubuntu/CentOS/RHEL
wget https://releases.hashicorp.com/terraform/0.11.2/terraform_0.11.2_linux_amd64.zip
unzip terraform_0.11.2_linux_amd64.zip
mv terraform /usr/local/bin/
```

#### Install `triton-kubernetes`
Download Binary:
TODO

From Source:
```bash
go get -u github.com/joyent/triton-kubernetes
go install github.com/joyent/triton-kubernetes
triton-kubernetes --help
```

## How-To

To see the multi-cloud capabilities of Triton Kubernetes, we are going to create a global cluster manager with four identical Kubernetes environments. These environments are going to be all configured in HA mode and on different cloud providers ([Triton](#setup-questions-kubernetes-cluster-on-triton), [AWS](#setup-questions-kubernetes-cluster-on-aws), [Azure](#setup-questions-kubernetes-cluster-on-azure), [GCP](#setup-questions-kubernetes-cluster-on-gcp)).

### Starting a Global Cluster Manager
Install [**Triton Kubernetes**](#install-triton-kubernetes) and run it:

```bash
triton-kubernetes create manager
```

Follow the on screen instructions answering questions about the cluster. You can use the default by pressing “Enter”/”Return” key.

#### Setup Questions

| Question                                                                       | Default                                               | Input                |
|--------------------------------------------------------------------------------|-------------------------------------------------------|----------------------|
| Name your Global Cluster Manager                                               | global-cluster                                        | <kbd>enter</kbd>     |
| Do you want to set up the Global Cluster Manager in HA mode                    |                                                       | yes                  |
| Which Triton networks should be used for this environment                      | Joyent-SDC-Public                                     | <kbd>enter</kbd>     |
| Which Triton package should be used for Global Cluster Manager server(s)       | k4-highcpu-kvm-1.75G                                  | k4-highcpu-kvm-3.75G |
| Which Triton package should be used for Global Cluster Manager database server | k4-highcpu-kvm-1.75G                                  | k4-highcpu-kvm-3.75G |
| docker-engine install script                                                   | <https://releases.rancher.com/install-docker/17.03.sh> | <kbd>enter</kbd>     |

After verification of the entries, setup will start Cluster Manager in HA mode on Joyent Cloud. This will be a two node HA configuration with a shared database node.

### Adding Kubernetes Environments to Global Cluster Manager
Invoke the following command and follow the on screen instructions answering questions about the Kubernetes environment.

```bash
triton-kubernetes create cluster
```

#### Setup Questions: Kubernetes cluster on Triton

| Question | Default | Input |
|---------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|----------------------|
| Which cloud do you want to run your environment on | 1 | <kbd>enter</kbd> |
| Name your environment | triton-test | <kbd>enter</kbd> |
| Do you want this environment to run in HA mode |  | yes |
| Number of compute nodes for triton-test environment | 3 | <kbd>enter</kbd> |
| Which Triton networks should be used for this environment | Joyent-SDC-Public | <kbd>enter</kbd> |
| Which Triton package should be used for triton-test environment etcd nodes | k4-highcpu-kvm-1.75G | k4-highcpu-kvm-3.75G |
| Which Triton package should be used for triton-test environment orchestration nodes running apiserver/scheduler/controllermanager/... | k4-highcpu-kvm-1.75G | k4-highcpu-kvm-3.75G |
| Which Triton package should be used for triton-test environment compute nodes | k4-highcpu-kvm-1.75G | k4-highcpu-kvm-3.75G |
| docker-engine install script | <https://releases.rancher.com/install-docker/17.03.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode on Joyent Cloud. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).

#### Setup Questions: Kubernetes cluster on AWS

| Question | Default | Input |
|-------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|------------------|
| Which cloud do you want to run your environment on | 1 | 2 |
| AWS Access Key |  | \[AWS access key id] |
| AWS Secret Key |  | \[AWS secret key] |
| Name your environment | aws-test | <kbd>enter</kbd> |
| Do you want this environment to run in HA mode |  | yes |
| Number of compute nodes for aws-test environment | 3 | <kbd>enter</kbd> |
| Where should the aws-test environment be located | us-west-2 | <kbd>enter</kbd> |
| Which image should be used for the nodes | ami-0def3275 | <kbd>enter</kbd> |
| What size hosts should be used for aws-test environment etcd nodes | t2.micro | t2.small |
| What size hosts should be used for aws-test environment orchestration nodes running apiserver/scheduler/controllermanager/... | t2.micro | t2.small |
| What size hosts should be used for aws-test environment compute nodes | t2.micro | t2.small |
| Which ssh public key should these hosts be set up with |  | \[your public ssh key] |
| docker-engine install script | <https://releases.rancher.com/install-docker/17.03.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode running on AWS. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).

#### Setup Questions: Kubernetes cluster on Azure

| Question | Default | Input |
|---------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|--------------------------|
| Which cloud do you want to run your environment on | 1 | 3 |
| Azure subscription id |  | \[Azure subscription id] |
| Azure client id |  | \[Azure client id] |
| Azure client secret |  | \[Azure client secret] |
| Azure tenant id |  | \[Azure tenant id] |
| Name your environment | azure-test | <kbd>enter</kbd> |
| Do you want this environment to run in HA mode |  | yes |
| Number of compute nodes for azure-test environment | 3 | <kbd>enter</kbd> |
| Where should the azure-test environment be located | westus2 | <kbd>enter</kbd> |
| What size hosts should be used for azure-test environment etcd nodes | Standard_A1 | Standard_A2 |
| What size hosts should be used for azure-test environment orchestration nodes running apiserver/scheduler/controllermanager/... | Standard_A1 | Standard_A2 |
| What size hosts should be used for azure-test environment compute nodes | Standard_A1 | Standard_A2 |
| Which ssh public key should these hosts be set up with |  | \[public ssh key] |
| docker-engine install script | <https://releases.rancher.com/install-docker/17.03.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode running on Azure. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).

#### Setup Questions: Kubernetes cluster on GCP

| Question | Default | Input |
|-------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|------------------------------|
| Which cloud do you want to run your environment on | 1 | 4 |
| Path to GCP credentials file |  | \[GCP json credentials file] |
| GCP Project ID |  | \[GCP project id] |
| Name your environment | gcp-test | <kbd>enter</kbd> |
| Do you want this environment to run in HA mode |  | yes |
| Number of compute nodes for gcp-test environment | 3 | <kbd>enter</kbd> |
| Compute Region | us-west1 | <kbd>enter</kbd> |
| Instance Zone | us-west1-a | <kbd>enter</kbd> |
| What size hosts should be used for gcp-test environment etcd nodes | n1-standard-1 | n1-standard-2 |
| What size hosts should be used for gcp-test environment orchestration nodes running apiserver/scheduler/controllermanager/... | n1-standard-1 | n1-standard-2 |
| What size hosts should be used for gcp-test environment compute nodes | n1-standard-1 | n1-standard-2 |
| docker-engine install script | <https://releases.rancher.com/install-docker/17.03.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode running on GCP. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).

## Backend State

Triton Kubernetes persists state by leveraging one of the supported backends. This state is required to add/remove/modify infrastructure managed by Triton Kubernetes.

### Manta
Will persist state in the `/triton-kubernetes/` folder for the provided user in Manta Cloud Storage.

### Local
Will persist state in the `~/.triton-kubernetes/` folder on the machine Triton Kubernets was run on.

## Developing Locally

### Testing terraform module changes
The `SOURCE_URL` flag will override the default terraform module source. Default is `github.com/joyent/triton-kubernetes`.

The `SOURCE_REF` flag will override the default branch/tag/commit reference for the terraform module source. Default is `master`

Testing local changes
```bash
SOURCE_URL=/full/path/to/working/dir/triton-kubernetes ./triton-kubernetes
```

Testing remote changes
```bash
SOURCE_URL=github.com/fayazg/triton-kubernets SOURCE_REF=new-branch ./triton-kubernetes
```
