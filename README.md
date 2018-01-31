<a>
  <img src="https://www.joyent.com/assets/img/external/triton-kubernetes.svg" width="100%" height="144">
</a>

Triton Kubernetes is a multi-cloud Kubernetes solution. It has a global cluster manager which can run on any cloud - Public, Private or Bare Metal and manages Kubernetes environments. The current release uses Triton (Joyent public cloud). You can update the APIs to switch to another cloud or log an enhancement request. 

The cluster manager will manage environments running on any region of any supported cloud. Out of box, Amazon, Azure, Google and Triton (public and private) are supported. If not using a cloud, environments on bare metal servers are also planned to be supported.For an example set up, look at the [How-To](#how-to) section.

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

Triton Kubernetes allows you to create/destroy global cluster managers, Kubernetes environments and individual cluster nodes. You can also get information on a cluster manager or Kubernetes environment. Triton Kubernetes provides these features through the `create`, `destroy` and `get` commands.

### Create

```bash
triton-kubernetes create [manager or cluster or node]
```

Creates a new cluster manager, kubernetes cluster or individual kubernetes cluster node.

When creating a new kubernetes cluster, you must specify the cloud provider for that cluster (Triton, AWS, Azure GCP).

### Destroy

```bash
triton-kubernetes destroy [manager or cluster or node]
```

Destroys an existing cluster manager, kubernetes cluster or individual kubernetes cluster node.

### Get

```bash
triton-kubernetes get [manager or cluster]
```

Displays cluster manager or kubernetes cluster details.

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
