## Quick Start Guide
This guide will serve as an example on how to run the installer and what it expects.
We are going to document how you can start a Kubernete cluster manager and start three environments (HA/non-HA on triton and 1 non-HA on azure) for it to manage.

> NOTE: This package has been tested on OSX.

### Pre-Reqs
In order to start running Kubernetes Triton Supervisor, you must create a **Triton** account and install **jq** (required), **Triton CLI** (optional), and **Terraform** (downloaded automatically if not found).

Having Triton CLI will allow you to easily check the status of your instances, and delete them (the supervisor at its current state can't remove instances). Also if you have a working triton profile, with executing `eval "$(triton env)"` you can have the supervisor pull/use that triton profile as your triton account and you won't have to provide triton details to the supervisor.

#### Install Triton

On OSX, `jq` can be installed simply by running `brew install jq` or downloading the [binary](https://github.com/stedolan/jq/releases/download/jq-1.5/jq-osx-amd64) into a directory in your path (`/usr/local/bin/`).

#### Terraform (optional)

Latest version of Terraform will be downloaded automatically under the `<k8s-triton-supervisor>/bin/` directory, if you have an older one or the one under your `$PATH` is older.

#### Install the Kubernetes CLI (optional)

This can be used to run/deploy apps and/or check the status of your Kubernetes environment.

There are different ways to [install `kubectl`](https://kubernetes.io/docs/tasks/kubectl/install/), but the simplest way is via `curl`:

```sh
# OS X
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl

# Linux
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl

# Windows
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/windows/amd64/kubectl.exe
```

### Components
Short description of each component that will be started as an example:

#### K8s Triton Supervisor
This is name of the package which builds/runs a multi-environment and multi-cloud Kubernetes solution.

#### Cluster Manager
Cluster Manager is what helps manage all the Kubernetes environments (clusters) easily from one place.

It can be ran in HA (active/active configuration with a shared mysqldb) or non-HA configuration. For anything more than a test, HA would be the recommended configuration and is what this example go over.

#### Kubernetes Environment
A fully containerized Kubernetes cluster which can be ran as HA (dedicated 3 node ETCD + 3 node k8s services + x compute node) configuration or non-HA (ETCD/k8s services scales/runs on compute nodes) configuration.

Example for both configurations is provided here.


### Get and test the package
Download the k8s-triton-supervisor package and run `k8s-triton-supervisor.sh` from inside package home.

Follow the on screen instructions answering questions about the cluster. You can use the default by pressing “Enter”/”Return” key.

```
$ git clone https://github.com/joyent/k8s-triton-supervisor.git -b multicloud-cli
Cloning into 'k8s-triton-supervisor'...
remote: Counting objects: 393, done.
remote: Compressing objects: 100% (54/54), done.
remote: Total 393 (delta 31), reused 63 (delta 20), pack-reused 319
Receiving objects: 100% (393/393), 5.39 MiB | 4.00 MiB/s, done.
Resolving deltas: 100% (165/165), done.
$ cd k8s-triton-supervisor 
$ ./k8s-triton-supervisor.sh

    Usage:
        ./k8s-triton-supervisor.sh (-c|-e) [conf file]
    Options:
        -c  create a cluster manager on Triton
        -e  create and add a Kubernetes environment to an existing cluster manager
```

There are two options here, `-c` to create cluster manager and `-e` to create/add an environment to the cluster manager.

> WARM: There is also a cleanup flag `--cleanAll true` which will delete the entire setup and reset this package. Hosts still have to be removed manually.

#### Brief Overview of Parameters
Setup includes questions about naming, package, networking, cloud details, and also has verification steps (yes/no).
Everything is case sensitive and it is always best to copy and paste instead of typing out options.
> HELP: There is the option to skip the interactive CLI and do a silent install by providing a conf file. Cluster Manager and Kubernetes environments have special config options so copy the appropriate conf template and modify as needed.

#### Private Repositories
There is the option to use private repositories. The images that need to be pushed to the private repositories are listed below.

```
{
   "repositories":[
      "busybox",
      "gcr.io/google_containers/heapster-amd64",
      "gcr.io/google_containers/heapster-grafana-amd64",
      "gcr.io/google_containers/heapster-influxdb-amd64",
      "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64",
      "gcr.io/google_containers/k8s-dns-kube-dns-amd64",
      "gcr.io/google_containers/k8s-dns-sidecar-amd64",
      "gcr.io/google_containers/kubernetes-dashboard-amd64",
      "gcr.io/google_containers/pause-amd64",
      "gcr.io/kubernetes-helm/tiller",
      "rancher/agent",
      "rancher/dns",
      "rancher/etc-host-updater",
      "rancher/etcd",
      "rancher/healthcheck",
      "rancher/k8s",
      "rancher/kubectld",
      "rancher/kubernetes-agent",
      "rancher/kubernetes-auth",
      "rancher/lb-service-rancher",
      "rancher/metadata",
      "rancher/net",
      "rancher/network-manager",
      "rancher/server"
   ]
}
```

In the conf file for Cluster Manager and the environments, update the registry details.
`rancher_registry` must have all the _docker hub_ (rancher) images and the `k8s_registry` should include all the _gcr.io_ images. If there is one registry for all the images, then both `rancher_registry` and `k8s_registry` must be set to the same.
> WARN: If using private registry, both `rancher_registry` and `k8s_registry` configurations must be set. They can point to the same registry.


#### WARNINGS
-   Don't use the same environment more than once
-   Everything is case sensitive
-   Don't delete or modify manually any files from `<package>/terraform/` directory
-   If you want to delete a host, use Rancher's UI to deactivate and delete it, and delete it from the cloud provider using the cli or their portal
-   Always run `k8s-triton-supervisor.sh` from inside the package directory (`cd` into the package so that the same cluster manager can be used)

#### Run silently with conf files
Steps to start a cluster with an environment, all on triton, clone the repo/branch:
```sh
$ git clone https://github.com/joyent/k8s-triton-supervisor.git -b multicloud-cli
$ cd k8s-triton-supervisor
```
Modify **_template_clustermanager.conf_** and **_template_triton_environment.conf_** files with your credentials and the start cluster manager first, followed by environment:
```sh
$ ./k8s-triton-supervisor.sh -c template_clustermanager.conf
...
$ ./k8s-triton-supervisor.sh -e template_triton_environment.conf
...
```
The first few lines that the script will output is confirmation of what is in the config files and at the end, when it exits, it will output the Cluster Manager and environment URLs:
```sh
State path: terraform.tfstate

Outputs:

masters = [
    72.2.119.90,
]

Environment triton-noha has been started.
This is a non-HA setup so Kubernetes services could run on any of the compute nodes.
Cluster Manager URL:
    http://72.2.119.90:8080/settings/env
Kubernetes Hosts URL:
    http://72.2.119.90:8080/env/1a74/infra/hosts?mode=dot
Kubernetes Health:
    http://72.2.119.90:8080/env/1a74/apps/stacks?which=cattle

NOTE: Nodes might take a few minutes to connect and come up.

To start another environment, run:
    ./k8s-triton-supervisor.sh -e
```

#### Set up an HA Cluster Manager (always on Triton) with triton profile that is set `eval "$(triton env)"`
> NOTE: Not having triton profile will just add 3-4 more questions you will have to answer. For an example of what that will look like, look at the [last environment](#set-up-an-non-ha-cluster-manager-without-triton-profile-set) added.

If there is an acceptable default (defaults are in parentheses) provided, you can accept it by leaving the input blank.
> WARNING: Keep in mind that there is no input verification. If you enter un-acceptable values, the setup will still continue.

```
$ pwd
/Users/fayazg/tmp
$ ./k8s-triton-supervisor.sh -c

Downloading latest Terraform zip'd binary

Extracting Terraform executable
Using /Users/fayazg/tmp/k8s-triton-supervisor/bin/terraform ...
```
> If latest terraform binary is not found, it will download it

```
Name your Global Cluster Manager: (global-cluster) my-globalcm
```
> This will be prepended to the host names; it is case sensitive and must be alphanumeric and can have `-`

```
Do you want to set up the Global Cluster Manager in HA mode? (yes | no) yes
```
> Case sensitive and will only accept `yes` or `no`

```
From below options:
Joyent-SDC-Public
Joyent-SDC-Private
Both
Which Triton networks should be used for this environment: (Joyent-SDC-Public) Both
```
> Case sensitive. Options that provide a list, as here given three options, it is best to just copy and paste your answer so that there are no spelling or capitalization issues.

```
From below packages:
k4-highcpu-kvm-250M
k4-highcpu-kvm-750M
k4-highcpu-kvm-1.75G
k4-highcpu-kvm-3.75G
k4-highcpu-kvm-7.75G
k4-highcpu-kvm-15.75G
k4-general-kvm-3.75G
k4-general-kvm-7.75G
k4-general-kvm-15.75G
k4-general-kvm-31.75G
k4-highram-kvm-15.75G
k4-highram-kvm-31.75G
k4-highram-kvm-63.75G
k4-bigdisk-kvm-15.75G
k4-fastdisk-kvm-31.75G
k4-bigdisk-kvm-31.75G
k4-fastdisk-kvm-63.75G
k4-bigdisk-kvm-63.75G
Which Triton package should be used for Global Cluster Manager server(s): (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-7.75G
```
> Case sensitive. Options that provide a list, as here given three options, it is best to just copy and paste your answer so that there are no spelling or capitalization issues.

```
Which Triton package should be used for Global Cluster Manager database server: (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-3.75G
```
> Case sensitive. Options that provide a list, as here given three options, it is best to just copy and paste your answer so that there are no spelling or capitalization issues.

```
############################################################

Cluster Manager my-globalcm will be created on Triton.
my-globalcm will be running in HA configuration and provision three Triton machines ...
    my-globalcm-master-1 k4-highcpu-kvm-7.75G
    my-globalcm-master-2 k4-highcpu-kvm-7.75G
    my-globalcm-mysqldb k4-highcpu-kvm-3.75G

Do you want to start the setup? (yes | no) yes
```
> Case sensitive and will only accept `yes` or `no`
> After this point, the `k8s-triton-supervisor.sh` updates terraform module for the cluster manager and call `terraform`.  

```
Downloading modules...
...
Initializing the backend...
...
Initializing provider plugins...
...
data.template_file.install_rancher_mysqldb: Refreshing state...
...
State path: terraform.tfstate

Outputs:

masters = [
    72.2.119.90,
    72.2.119.84
]

Cluster Manager my-globalcm has been started.
This is an HA Active/Active setup so you can use either of the IP addresses.
    http://72.2.119.90:8080/settings/env

Next step is adding Kubernetes environments to be managed here.
To start your first environment, run:
    ./k8s-triton-supervisor.sh -e
```
> After the setup completes, it will give some info on what it did and how to access your HA Cluster Manager.
> 
> WARNING: `terraform` will download some plugins from the Internet (on a first run this will almost always happen)

#### Set up an HA Kubernetes environment on Triton with triton profile set `eval "$(triton env)"`
If there is an acceptable default (defaults are in parentheses) provided, you can accept it by leaving the input blank.
> WARNING: Keep in mind that there is no input verification. If you enter un-acceptable values or choose a duplicate environment name, the setup will still continue and crash.

```
$ ./k8s-triton-supervisor.sh -e
Using /Users/fayazg/tmp/k8s-triton-supervisor/bin/terraform ...

From clouds below:
1. Triton
2. Azure
Which cloud do you want to run your environment on: (1)
Name your environment: (triton-test)
Do you want this environment to run in HA mode? (yes | no) yes
Number of compute nodes for triton-test environment: (3)
From below options:
Joyent-SDC-Public
Joyent-SDC-Private
Both
Which Triton networks should be used for this environment: (Joyent-SDC-Public)
From below packages:
k4-highcpu-kvm-250M
k4-highcpu-kvm-750M
k4-highcpu-kvm-1.75G
k4-highcpu-kvm-3.75G
k4-highcpu-kvm-7.75G
k4-highcpu-kvm-15.75G
k4-general-kvm-3.75G
k4-general-kvm-7.75G
k4-general-kvm-15.75G
k4-general-kvm-31.75G
k4-highram-kvm-15.75G
k4-highram-kvm-31.75G
k4-highram-kvm-63.75G
k4-bigdisk-kvm-15.75G
k4-fastdisk-kvm-31.75G
k4-bigdisk-kvm-31.75G
k4-fastdisk-kvm-63.75G
k4-bigdisk-kvm-63.75G
Which Triton package should be used for triton-test environment etcd nodes: (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-3.75G
Which Triton package should be used for triton-test environment orchestration nodes running apiserver/scheduler/controllermanager/...: (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-3.75G
Which Triton package should be used for triton-test environment compute nodes: (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-7.75G
############################################################

Environment triton-test will be created on Triton.
triton-test will be running in HA configuration ...
6 dedicated hosts will be created ...
    triton-test-etcd-[123] k4-highcpu-kvm-3.75G
    triton-test-orchestration-[123] k4-highcpu-kvm-3.75G
3 compute nodes will be created for this environment ...
    triton-test-compute-# k4-highcpu-kvm-7.75G

Do you want to start the setup? (yes | no) yes
```
> Similar to running with `-c` option, everything is case sensitive and for option related to package and network option, it is always a better idea to copy and paste.

```
...
State path: terraform.tfstate

Outputs:

masters = [
    72.2.119.90,
    72.2.119.84
]

Environment triton-test has been started.
This is an HA setup of Kubernetes cluster so there are 3 dedicated etcd and 3 orchestration nodes.
Cluster Manager URL:
    http://72.2.119.90:8080/settings/env
Kubernetes Hosts URL:
    http://72.2.119.90:8080/env/1a7/infra/hosts?mode=dot
Kubernetes Health:
    http://72.2.119.90:8080/env/1a7/apps/stacks?which=cattle

NOTE: Nodes might take a few minutes to connect and come up.

To start another environment, run:
    ./k8s-triton-supervisor.sh -e
```
> This time `terraform` should not need to download anything from the Internet.

#### Set up an non-HA Kubernetes environment on Triton, without triton profile set
```
$ ./k8s-triton-supervisor.sh -e
Using /Users/fayazg/tmp/k8s-triton-supervisor/bin/terraform ...

From clouds below:
1. Triton
2. Azure
Which cloud do you want to run your environment on: (1)
Your Triton account login name: fayazg
The Triton CloudAPI endpoint URL: (https://us-east-1.api.joyent.com)
Your Triton account key id: 2c:53:bc:63:97:9e:79:3f:91:35:5e:f4:c8:23:88:37
Name your environment: (triton-test) triton-noha
Do you want this environment to run in HA mode? (yes | no) no
Number of compute nodes for triton-noha environment: (3) 7
From below options:
Joyent-SDC-Public
Joyent-SDC-Private
Both
Which Triton networks should be used for this environment: (Joyent-SDC-Public) Joyent-SDC-Public
From below packages:
k4-highcpu-kvm-250M
k4-highcpu-kvm-750M
k4-highcpu-kvm-1.75G
k4-highcpu-kvm-3.75G
k4-highcpu-kvm-7.75G
k4-highcpu-kvm-15.75G
k4-general-kvm-3.75G
k4-general-kvm-7.75G
k4-general-kvm-15.75G
k4-general-kvm-31.75G
k4-highram-kvm-15.75G
k4-highram-kvm-31.75G
k4-highram-kvm-63.75G
k4-bigdisk-kvm-15.75G
k4-fastdisk-kvm-31.75G
k4-bigdisk-kvm-31.75G
k4-fastdisk-kvm-63.75G
k4-bigdisk-kvm-63.75G
Which Triton package should be used for triton-noha environment compute nodes: (k4-highcpu-kvm-1.75G) k4-highcpu-kvm-3.75G
############################################################

Environment triton-noha will be created on Triton.
triton-noha will be running in non-HA configuration ...
7 compute nodes will be created for this environment ...
    triton-noha-compute-# k4-highcpu-kvm-3.75G

Do you want to start the setup? (yes | no) yes
...
State path: terraform.tfstate

Outputs:

masters = [
    72.2.119.90,
    72.2.119.84
]

Environment triton-noha has been started.
This is a non-HA setup so Kubernetes services could run on any of the compute nodes.
Cluster Manager URL:
    http://72.2.119.90:8080/settings/env
Kubernetes Hosts URL:
    http://72.2.119.90:8080/env/1a74/infra/hosts?mode=dot
Kubernetes Health:
    http://72.2.119.90:8080/env/1a74/apps/stacks?which=cattle

NOTE: Nodes might take a few minutes to connect and come up.

To start another environment, run:
    ./k8s-triton-supervisor.sh -e
```
> Similar to HA config but since triton profile isn't set, had to provide triton account login, CloudAPI endpoint URL, and account key id. `k8s-triton-supervisor.sh` will look through the `$HOME/.ssh/` directory for the matching ssh key and use that.
>
> If the matching ssh key isn't under `$HOME/.ssh/`, `k8s-triton-supervisor.sh` will prompt for it's location.

#### Set up an non-HA Kubernetes environment on Azure
This is going to be similar to setting it up on Triton with a minor difference.
-   Select option 2 when selecting a cloud
-   The next 4 parameters will be your subscription id, client id, client secret and tenant id
-   For Azure you have to provide a location where the hosts will be created. This is case sensitive so just copy and paste (`az account list-locations` from the cli gives a list of all available locations).
-   Case sensitive. The size of hosts for "West US 2" can be looked up using `az vm list-sizes -l "West US 2"` or `az vm list-sizes -l westus2`, just copy and paste the size name you want.
-   Last difference is, public ssh key that will be loaded into the VMs that get provisioned so they can be ssh-ed into without a password

```
$ ./k8s-triton-supervisor.sh -e
Using /Users/fayazg/tmp/k8s-triton-supervisor/bin/terraform ...

From clouds below:
1. Triton
2. Azure
Which cloud do you want to run your environment on: (1) 2
Azure subscription id: c26b3274-21ba-4d33-bc75-d3a1dfb523af
Azure client id: 71fb8ab2-b5ce-494b-9a3d-9bf5bc5a744c
Azure client secret: 2ae408ac-cf48-4616-9cd2-6c00dbafec0b
Azure tenant id: 3f6358ac-cf48-4656-9cd2-1d003ba4e50d
Name your environment: (azure-test)
Do you want this environment to run in HA mode? (yes | no) no
Number of compute nodes for azure-test environment: (3) 5
Where should the azure-test environment be located: (West US 2)
What size hosts should be used for azure-test environment compute nodes: (Standard_A1) Standard_A2
Which ssh public key should these hosts be set up with: (/Users/fayazg/.ssh/id_rsa.pub)
############################################################

Environment azure-test will be created on Azure.
azure-test will be running in non-HA configuration ...
5 compute nodes will be created for this environment ...
    azure-test-compute-# Standard_A2

Do you want to start the setup? (yes | no) yes
...
```
> All case sensitive.
>
> Once everything is provisioned, the output will be displayed similar to the [non-HA environment on Triton](#set-up-an-non-ha-kubernetes-environment-on-triton-without-triton-profile-set).
