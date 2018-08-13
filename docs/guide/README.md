# Triton Kubernetes Quick Start Guide

## Installation

### Pre-Requisites

**Triton Kubernetes** requires [`jq`](#install-jq) and [`terraform`](#install-terraform).

[jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor. It is leveraged by `triton-kubernetes`.

[Terraform](https://www.terraform.io/) enables you to safely and predictably create, change, and improve production infrastructure. It is an open source tool that codifies APIs into declarative configuration files that can be shared amongst team members, treated as code, edited, reviewed, and versioned.

#### Install `jq`

```bash
# OS X using brew
brew install jq

# Debian/Ubuntu
apt-get install jq

# CentOS/RHEL
yum install jq
```

#### Install `Terraform`

```bash
# OS X using brew
brew install terraform

# Debian/Ubuntu/CentOS/RHEL
wget https://releases.hashicorp.com/terraform/0.11.2/terraform_0.11.2_linux_amd64.zip
unzip terraform_0.11.2_linux_amd64.zip
mv terraform /usr/local/bin/
```

#### Install `triton-kubernetes`

[Install from source](/docs/guide/building-cli.md):

```bash
go get -u github.com/joyent/triton-kubernetes
go install github.com/joyent/triton-kubernetes
triton-kubernetes --help
```

[Install from pre-built packages](/docs/guide/installing-cli.md)

## Running Triton Kubernetes

`triton-kubernetes` can run as an interactive cli, or in [silent mode](silent-install-yaml.md) (`--non-interactive`) using yaml configuration files.

The `triton-kubernetes` cli can:

- create a cluster manager
- destroy a cluster manager and all clusters it is managing
- add/remove a cluster to/from an existing cluster manager
- backup/restore a kubernetes namespace from any of your clusters to manta/S3
- query your existing cluster managers and clusters

The cli `triton-kubernetes` allows for creating and managing a kubernetes deployment only. Application deployments will still need to be done using `kubectl`.

>Note: Keep in mind that every cloud has a resource quota. If that quota has been reached, Triton-Kubernetes will not be able to provision new machines and will throw errors.

## Supported Clouds

- [AWS](aws)
- [AZURE](azure)
- [BareMetal](bare-metal)
- [GCP](gcp)
- [GKE](gke)
- [Triton](triton)
- [vSphere](vSphere)

## Backend State

Triton Kubernetes persists state by leveraging one of the supported backends. This state is required to add/remove/modify infrastructure managed by Triton Kubernetes.

### Manta

Will persist state in the `/triton-kubernetes/` folder for the provided user in Manta Cloud Storage.

### Local

Will persist state in the `~/.triton-kubernetes/` folder on the machine Triton Kubernetes was run on.

## Helm

Helm is already installed on the Kubernetes cluster but you will be required to create Service account with cluster-admin role.

You can add a service account to Tiller using the `--service-account <NAME>` flag while you're configuring Helm. As a prerequisite, you'll have to create a role binding which specifies a [role](https://kubernetes.io/docs/admin/authorization/rbac/#role-and-clusterrole) and a [service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) name that have been set up in advance.

```bash
# create service account 'tiller'
$ kubectl create serviceaccount --namespace kube-system tiller
# create cluster role binding
$ kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
# deploy changes
$ kubectl patch deploy --namespace kube-system tiller-deploy -p '{ "spec" : { "template" : { "spec" : { "serviceAccount" : "tiller" }}}}'
# upgrade
$ helm init --service-account tiller --upgrade
```

_Note: The cluster-admin role is created by default in a Kubernetes cluster, so you don't have to define it explicitly._

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
