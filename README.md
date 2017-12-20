<a>
  <img src="https://www.joyent.com/assets/img/external/triton-kubernetes.svg" width="100%" height="144">
</a>

## Quick Start Guide

This is a multi-cloud Kubernetes solution. **Triton Kubernetes** has a global cluster manager which will run on Triton and manages Kubernetes environments. This cluster manager will manage environments running on any region of any supported cloud. For an example set up, look at the [How-To](#how-to) section.

![Triton-Kubernetes](docs/imgs/Triton-Kubernetes.png)

> NOTE: This package has been tested on Linux/OSX.

### Pre-Reqs
In order to start running **Triton Kubernetes**, you must create a [Triton](https://my.joyent.com/) account and install [`triton` CLI](#install-triton), [`wget`](#install-wget-), and [`kubectl`](#install-the-kubernetes-cli). `terraform` is also a requirement, but if it isn't found, it will be downloaded automatically.

[Triton](https://www.joyent.com/why) is our container-native and open source cloud, which we will use to provide the infrastructure required for your Kubernetes cluster. 

[Terraform](https://www.terraform.io/) enables you to safely and predictably create, change, and improve production infrastructure. It is an open source tool that codifies APIs into declarative configuration files that can be shared amongst team members, treated as code, edited, reviewed, and versioned.

#### Install Triton CLI

In order to install `triton`, first you must have a [Triton account](https://sso.joyent.com/signup). As a new user you will receive a $250 credit to enable you to give Triton and Kubernetes a test run, but it's important to [add your billing information](https://my.joyent.com/main/#!/account/payment) and [add an ssh key](https://my.joyent.com/main/#!/account) to your account. If you need instructions for how to generate and SSH key, [read our documentation](https://docs.joyent.com/public-cloud/getting-started).

1.  Install [Node.js](https://nodejs.org/en/download/) and run `npm install -g triton` to install Triton CLI.
1.  `triton` uses profiles to store access information. You'll need to set up profiles for relevant data centers.
    +   `triton profile create` will give a [step-by-step walkthrough](https://docs.joyent.com/public-cloud/api-access/cloudapi) of how to create a profile.
	+   Choose a profile to use for your Kubernetes Cluster.
1.  Get into the Triton environment with `eval $(triton env <profile name>)`.
1.  Run `triton info` to test your configuration.

#### Terraform

Terraform will be downloaded automatically under the `<triton-kubernetes>/bin/` directory.

#### Install `wget`

Install `wget` for the system you are on:

```sh
# OS X using brew
brew install wget

# Debian/Ubuntu
apt-get install wget

# CentOS/RHEL
yum install wget
```

#### Install the Kubernetes CLI

There are different ways to [install `kubectl`](https://kubernetes.io/docs/tasks/kubectl/install/), but the simplest way is via `curl`:

```sh
# OS X
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl

# Linux
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl

# Windows
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/windows/amd64/kubectl.exe
```

## How-To

To see the multi-cloud capabilities of Triton Kubernetes, we are going to create a global cluster manager with four identical Kubernetes environments. These environments are going to be all configured in HA mode and on different cloud providers ([Triton](#setup-questions-kubernetes-cluster-on-triton), [AWS](#setup-questions-kubernetes-cluster-on-aws), [Azure](#setup-questions-kubernetes-cluster-on-azure), [GCP](#setup-questions-kubernetes-cluster-on-gcp)).

### Starting a Global Cluster Manager
Make sure you have a Triton profile created and active.
Download the **Triton Kubernetes** package and run `triton-kubernetes.sh`:

```bash
$ eval "$(triton env)"
$ git clone https://github.com/joyent/triton-kubernetes.git
Cloning into 'triton-kubernetes'...
$ cd triton-kubernetes 
$ ./triton-kubernetes.sh -c
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
| docker-engine install script                                                   | <https://releases.rancher.com/install-docker/1.12.sh> | <kbd>enter</kbd>     |

After verification of the entries, setup will start Cluster Manager in HA mode on Joyent Cloud. This will be a two node HA configuration with a shared database node.

### Adding a Kubernetes Environments to Global Cluster Manager
From the same repository directory that [global-cluster](#starting-a-global-cluster-manager) Cluster Manager was created, invoke the following command and follow the on screen instructions answering questions about the Kubernetes environment. We are adding four environments one on each cloud and their inputs are below. You can use the defaults by pressing “Enter”/”Return” key:

```bash
$ ./triton-kubernetes.sh -e
```

#### Setup Questions: Kubernetes cluster on Triton

The Triton credentials are pulled from environment variables. If `eval "$(triton env)"` was not ran, you will be prompted for Triton account, CloudAPI endpoint URL, ssh key ID and private key.

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
| docker-engine install script | <https://releases.rancher.com/install-docker/1.12.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode on Joyent Cloud. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).

#### Setup Questions: Kubernetes cluster on AWS

Before getting started, you will need to find the ImageId for an Ubuntu 16.04 instance in the region you are planning on deploying.

Canonical keeps a list of AMI ids here: https://cloud-images.ubuntu.com/locator/ec2/ you are looking for Ubuntu Server 16.04 amd64 with instance type of _hvm:ebs-ssd_

If you have the aws cli tools installed (and jq), then you can run the following command (don't forget to modify the region, otherwise you won't get the correct AMI image id.)

```bash
aws --region us-west-2 ec2 describe-images --owners 099720109477 --filters \
  "Name=root-device-type,Values=ebs" "Name=virtualization-type,Values=hvm" \
  "Name=architecture, Values=x86_64" "Name=is-public,Values=true" \
  "Name=name,Values='*hvm-ssd/ubuntu-xenial-16.04-amd64-server*'" \
  --query 'Images[*].{ImageId:ImageId,Name:Name,CreationDate:CreationDate}' | \
  jq '.|sort_by(.CreationDate)|last'
```

If you need more information on the AMI, you can use the following command. (Unfortunately, it doesn't tell you which region the image lives in.)

```bash
aws ec2 describe-images --image-id ami-0def3275
```

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
| docker-engine install script | <https://releases.rancher.com/install-docker/1.12.sh> | <kbd>enter</kbd> |

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
| docker-engine install script | <https://releases.rancher.com/install-docker/1.12.sh> | <kbd>enter</kbd> |

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
| docker-engine install script | <https://releases.rancher.com/install-docker/1.12.sh> | <kbd>enter</kbd> |

After verification of the entries, setup will create a Kubernetes environment in HA mode running on GCP. This will be a three worker/three ETCD/three Kubernetes Services node configuration managed by the previously started cluster manager ([global-cluster](#starting-a-global-cluster-manager)).
