# k8sontriton
This tutorial explains how to automate running a Kubernetes cluster on Joyent Cloud using Rancher.  
We are using triton+Terraform+Ansible to automate Kubernetes setup. Terraform is used to provision the KVMs while Ansible roles have been created to install pre-reqs and docker-engine, Rancher server with a kubernetes environment, and connect nodes to it.

Before running the CLI, first thing that should be set up is the [Triton CLI and triton profile](#triton-cli). This profile and triton environment variables (will be set by running`eval $(triton env)`) are used to connect and provision KVMs. During the provisioning, terraform will store information about the KVMs created, so that an ansible hosts and configuration file can be created which will be used by ansible to finish the cluster setup.

## Architecture
k8sontriton will create an environment similar to the the diagram below:
![Architecture Diagram](img/20170323b-Triton-Kubernetes.jpg "Architecture Diagram")

The default setup includes a kvm for Rancher server container to run on, and multiple node kvms connected to the Kubernetes environment. Kubernetes environment will be accessible through the kubectl CLI (kubectl config is provided by rancher) and Kubernetes dashboard.
## Pre-Reqs
The following pre-reqs are to be set up on the machine performing the Kubernetes set up.
1. [Install and set up triton CLI and a profile](https://docs.joyent.com/public-cloud/api-access/cloudapi)  
   Install [nodejs](https://nodejs.org/en/download/) and run `npm install -g triton`.

   Triton CLI needs to be configured with a profile because we will be using it and its configuration information to set up our Kubernetes cluster.

   To setup triton CLI, you need to [create an account](https://sso.joyent.com/signup) with Joyent, [add your billing information](https://my.joyent.com/main/#!/account/payment) and [ssh key](https://my.joyent.com/main/#!/account) to your account.  
   For more information on how to create an account, billing and ssh key information, look at the [Getting started](https://docs.joyent.com/public-cloud/getting-started) page.

   Note: The data center that will be used must have KVM images available for provisioning.
1. [Install terraform](https://www.terraform.io/intro/getting-started/install.html)  
   Terraform is an infrastructure building, changing and versioning tool. It will be used to provision KVMs for the Kubernetes cluster.

   Terraform can be installed by getting the appropriate [package](https://releases.hashicorp.com/terraform/0.8.5/terraform_0.8.5_darwin_amd64.zip) for your system which includes a single binary program terraform. Place this binary in a directory that is on the `PATH`.

   Note: Supported version of terraform is needed ([4.15](https://releases.hashicorp.com/terraform/0.8.5/terraform_0.8.5_darwin_amd64.zip))
1. Install Ansible  
   [Ansible](http://docs.ansible.com/ansible/index.html) is and IT automation tool we are using to set up the Kubernetes cluster on JoyentCloud KVMs.

   There are [multiple ways to install ansible](http://docs.ansible.com/ansible/intro_installation.html) depending on your operating system. Simplest way to do this is by using `pip` command (python package manager).  
   `sudo pip install ansible`
1. Python v2.x  
   OSX comes with python 2.7, but if you are on windows or linux, make sure you have [python](https://www.python.org/downloads/) installed on your system.
## Create a cluster
k8sontriton uses [triton](#triton-cli), [terraform](#terraform) and [ansible](#ansible) to set up and interact with Kubernetes cluster.
To start setting up a cluster, first we need to confirm that the [pre-reqs](#pre-reqs) are met. Then run setup.sh and answer the questions prompted. Default values will be shown in parentheses and if no input is provided, defaults will be used.
```
$ ./setup.sh
Name your Kubernetes environment: (k8s dev)
```
Provide a name for the Kubernetes environment and press Enter.
```
Describe this Kubernetes environment: (k8s dev)
```
Provide a description for the environment and press Enter. Environment description can have spaces.
```
Hostname of the master: (kubemaster)
```
Provide a hostname for the KVM that will run the Rancher Server container. This KVM will be used to interact with Rancher and Kubernetes environments. Hostname must start with a letter and must be alphanumeric.
```
Enter a string to use for appending to hostnames of all the nodes: (kubenode)
```
Provide a prefix which will be used for all the KVMs that will be connected as nodes. Must be alphanumeric.
```
How many nodes should this Kubernetes cluster have: (2)
```
Provide the number of nodes that should be created/connected to Kubernetes cluster.
```
From the networks below:
1.	Joyent-SDC-Private  909c0c0d-1455-404f-85bd-04f48b7b0059
2.	Joyent-SDC-Public  31428241-4878-47d6-9fba-9a8436b596a4
3.	My-Fabric-Network  0882d255-ac1e-41b2-b1d5-e08200ebb380
4.	kubernetes  0b206464-d655-4723-a848-86d0f28764c8
What networks should the master be a part of, provide comma separated values: (31428241-4878-47d6-9fba-9a8436b596a4)
```
Triton CLI is used here to pull all the active networks for the current SDC defined in triton profile. Provide a comma separated list of networks that the master KVM should be a part of (e.g. “2,4”).
```
From the networks below:
1.	Joyent-SDC-Private  909c0c0d-1455-404f-85bd-04f48b7b0059
2.	Joyent-SDC-Public  31428241-4878-47d6-9fba-9a8436b596a4
3.	My-Fabric-Network  0882d255-ac1e-41b2-b1d5-e08200ebb380
4.	kubernetes  0b206464-d655-4723-a848-86d0f28764c8
What networks should the nodes be a part of, provide comma separated values: (31428241-4878-47d6-9fba-9a8436b596a4)
```
Provide a comma separated list of networks that the KVMs used as Kubernetes nodes should be a part of (e.g. “2,4”). The nodes must be able to communicate to the master KVM or the setup will fail.
```
From the packages below:
1.	k4-bigdisk-kvm-15.75G  7741b8f6-2733-11e6-bdb9-bf11c4147d38
2.	k4-bigdisk-kvm-31.75G  14c01a1a-d0f8-11e5-ad69-1fd27456ad73
3.	k4-bigdisk-kvm-63.75G  14c0992c-d0f8-11e5-bd78-e71dad0f8626
4.	k4-fastdisk-kvm-31.75G  14bd9600-d0f8-11e5-a69c-97be6e961834
5.	k4-fastdisk-kvm-63.75G  14be13c8-d0f8-11e5-b55b-47eb44d4e064
6.	k4-general-kvm-15.75G  14ac8f5e-d0f8-11e5-a0e5-9b622a20595f
7.	k4-general-kvm-3.75G  14aba044-d0f8-11e5-8c88-eb339a5da5d0
8.	k4-general-kvm-31.75G  14ad1a32-d0f8-11e5-a465-8f264489308b
9.	k4-general-kvm-7.75G  14ac17a4-d0f8-11e5-a400-e39503e18b19
10.	k4-highcpu-kvm-1.75G  14b5edc4-d0f8-11e5-b4d2-b3e6e8c05f9d
11.	k4-highcpu-kvm-15.75G  14b783d2-d0f8-11e5-8d93-6ba10192d750
12.	k4-highcpu-kvm-250M  14b4ff36-d0f8-11e5-a8b1-e343c129d7f0
13.	k4-highcpu-kvm-3.75G  14b67ef6-d0f8-11e5-ba19-479de37c6f75
14.	k4-highcpu-kvm-7.75G  14b6fade-d0f8-11e5-85c5-4ff7918ab5c1
15.	k4-highcpu-kvm-750M  14b5760a-d0f8-11e5-9cb1-23c9c232c00e
16.	k4-highram-kvm-15.75G  14ba876c-d0f8-11e5-8a1b-ab02fdd17b07
17.	k4-highram-kvm-31.75G  14bafb20-d0f8-11e5-a5cf-e386b841ed87
18.	k4-highram-kvm-63.75G  14bb84f0-d0f8-11e5-8014-2fb7b19ccb24
What KVM package should the master and nodes run on: (14b6fade-d0f8-11e5-85c5-4ff7918ab5c1)
```
Triton CLI is used here to pull all the available kvm packages for the current SDC defined in triton profile. Provide a the number for kvm package to be used for all the nodes.

After the package detail has been provided, the CLI will verify all the entries before creating the cluster.
```
Verify that the following configuration is correct:

Name of kubernetes environment: k8s dev
Kubernetes environment description: k8s dev
Master hostname: kubemaster
All node hostnames will start with: kubenode
Kubernetes environment will have 1 nodes
Master server will be part of these networks: 31428241-4878-47d6-9fba-9a8436b596a4
Kubernetes nodes will be a part of these networks: 31428241-4878-47d6-9fba-9a8436b596a4
This package will be used for all the hosts: 14b6fade-d0f8-11e5-85c5-4ff7918ab5c1

Make sure the above information is correct before answering:
    to view list of networks call "triton networks -l"
    to view list of packages call "triton packages -l"
Make sure that the nodes and master are part of networks that can communicate with each other.
Is the above config correct (yes | no)? yes
```
Answer the verification question and the setup will start.

This will stored the entries, a [terraform](#terraform) configuration for the environment will be generated and terraform tasks will be started to provision the kvms. After terraform tasks are finished, [ansible configuration](#ansible-config-generation) files are generated and [ansible](#ansible) roles are started to install docker-engine, started rancher, create kubernetes environment and connect all the nodes to the kubernetes environment.
```
Congradulations, your Kubernetes cluster setup has been complete.
----> Rancher dashboard is at http://<ip of master>:8080

It will take a few minutes for all the Kubernetes process to start up before you can access Kubernetes Dashboard
----> To check what processes/containers are coming up, go to http://<ip of master>:8080/env/<env id>/infra/containers
    once all these containers are up, you should be able to access Kubernetes by its dashboard or using CLI
Waiting on Kubernetes dashboard to come up.

...................................................................
----> Kubernetes dashboard is at http://<ip of master>:8080/r/projects/<env id>/kubernetes-dashboard:9090/
----> Kubernetes CLI config is at http://<ip of master>:8080/env/<env id>/kubernetes/kubectl

    CONGRATULATIONS, YOU HAVE CONFIGURED YOUR KUBERNETES ENVIRONMENT!
```
At the end after all kvms have been provisioned and kubernetes cluster has been set up and running, a [message](#end-message) will appear with details on how to connect and where to access the kubernetes cluster.

## Components
Below are the tools that are used by k8sontriton and also detailed description of some of the tasks it performs.
### Triton CLI
Triton CLI tool uses CLoudAPI to manage infrastructure in Triton datacenters. We will be using Triton CLI to pull network and package information from the current Triton datancenter configured in the profile.

<sub>For more information on Triton, click [here](https://docs.joyent.com/public-cloud).</sub>
#### Install Triton CLI
Cloud API tools require Node.js, which can be found [here](http://nodejs.org/) if you don't have it installed.
Once Node.js is intalled, you can use `npm` to install the `triton` CLI tool:
```bash
$ sudo npm install -g triton
. . .
/usr/local/bin/triton -> /usr/local/lib/node_modules/triton/bin/triton
triton@4.11.0 /usr/local/lib/node_modules/triton
├── bigspinner@3.1.0
├── assert-plus@0.2.0
├── extsprintf@1.0.2
├── wordwrap@1.0.0
├── strsplit@1.0.0
├── node-uuid@1.4.3
├── read@1.0.7 (mute-stream@0.0.6)
├── semver@5.1.0
├── vasync@1.6.3
├── once@1.3.2 (wrappy@1.0.2)
├── backoff@2.4.1 (precond@0.2.3)
├── verror@1.6.0 (extsprintf@1.2.0)
├── which@1.2.4 (isexe@1.1.2, is-absolute@0.1.7)
├── cmdln@3.5.4 (extsprintf@1.3.0, dashdash@1.13.1)
├── lomstream@1.1.0 (assert-plus@0.1.5, extsprintf@1.3.0, vstream@0.1.0)
├── mkdirp@0.5.1 (minimist@0.0.8)
├── sshpk@1.7.4 (ecc-jsbn@0.1.1, jsbn@0.1.0, asn1@0.2.3, jodid25519@1.0.2, dashdash@1.13.1, tweetnacl@0.14.3)
├── rimraf@2.4.4 (glob@5.0.15)
├── tabula@1.7.0 (assert-plus@0.1.5, dashdash@1.13.1, lstream@0.0.4)
├── smartdc-auth@2.3.0 (assert-plus@0.1.2, once@1.3.0, clone@0.1.5, dashdash@1.10.1, sshpk@1.7.1, sshpk-agent@1.2.0, vasync@1.4.3, http-signature@1.1.1)
├── restify-errors@3.0.0 (assert-plus@0.1.5, lodash@3.10.1)
├── bunyan@1.5.1 (safe-json-stringify@1.0.3, mv@2.1.1, dtrace-provider@0.6.0)
└── restify-clients@1.1.0 (assert-plus@0.1.5, tunnel-agent@0.4.3, keep-alive-agent@0.0.1, lru-cache@2.7.3, mime@1.3.4, lodash@3.10.1, restify-errors@4.2.3, dtrace-provider@0.6.0)
```
<sub>For more information on installing `triton` CLI, click [here](https://docs.joyent.com/public-cloud/api-access/cloudapi#installation).</sub>
#### Create Profile
The `triton` CLI uses "profiles" to store access information. Profiles include data center URL, login name and SSH key fingerprint.
To create a profile:
```
$ triton profile create

A profile name. A short string to identify a CloudAPI endpoint to the `triton` CLI.
name: us-sw-1

The CloudAPI endpoint URL.
url: https://us-sw-1.api.joyent.com

Your account login name.
account: jill

The fingerprint of the SSH key you have registered for your account. You may enter a local path to a public or private key to have the fingerprint calculated for you.
keyId: ~/.ssh/<ssh key name>.id_rsa
Fingerprint: 2e:c9:f9:89:ec:78:04:5d:ff:fd:74:88:f3:a5:18:a5

Saved profile "us-sw-1"
```
<sub>For more information on how to set up profiles, click [here](https://docs.joyent.com/public-cloud/api-access/cloudapi#configuration).</sub>
### Terraform
Terraform is a tool for building, changing and versioning infrastructure. We are going to use terraform to provision KVMs, set up root access, and install python. Also as terraform provisions KVMs, it creates two files. One will be called masters.ip and another hosts.ip which will include the ip addresses of the masters KVMs and host KVMs that are provisioned.
#### Install Terraform
Terraform is distributed as a [binary package](https://www.terraform.io/downloads.html) and also can be compiled from source.
To install terraform, download the [appropriate package](https://www.terraform.io/downloads.html) for your system and unzip the package into a directory where terraform will be installed.
The final step is to make sure the directory you installed terraform to is on the `PATH`.

<sub>For more details on terraform, click [here](https://www.terraform.io/intro/index.html).</sub>
### Ansible Config Generation
Content of masters.ip and hosts.ip are merged into a hosts file and triton ssh key used by triton profile is set up to be used by ansible roles. The last thing that this setup does is update the ranchermaster variable with kubernetes environment name/description and master IP.
### Ansible
Ansible is an automation tool that can configure systems, deploy software, and orchestrate more advanced IT tasks like continuous deployments and rolling updates. We are using it to configure KVMs, install pre-reqs like docker-engine and Rancher, create Kubernetes environment and connect all the nodes to set up the cluster.
Ansible by default manages machines over SSH and requires Python 2.6 or 2.7 to be installed on all the hosts.
#### Install Ansible
There are [many ways to install ansible](http://docs.ansible.com/ansible/intro_installation.html), but the simplest would be to use Python package manager (`pip`).
If you have Python installed then you should be able to run the command below to install ansible:
```
$ sudo pip install ansible
```
<sub>For more details on ansible, click [here](http://docs.ansible.com/ansible/index.html).
### End Message
After terraform provisions the KVMs and ansible sets it up as a master or a node, CLI will check the kubernetes environment to make sure it is fully up and provide URLs to access it.
Multiple URLs will be provided as the system comes up:
```
Congradulations, your Kubernetes cluster setup has been complete.
----> Rancher dashboard is at http://<ip of master>:8080

It will take a few minutes for all the Kubernetes process to start up before you can access Kubernetes Dashboard
----> To check what processes/containers are coming up, go to http://<ip of master>:8080/env/<env id>/infra/containers
    once all these containers are up, you should be able to access Kubernetes by its dashboard or using CLI
Waiting on Kubernetes dashboard to come up.

...................................................................
----> Kubernetes dashboard is at http://<ip of master>:8080/r/projects/<env id>/kubernetes-dashboard:9090/
----> Kubernetes CLI config is at http://<ip of master>:8080/env/<env id>/kubernetes/kubectl

    CONGRATULATIONS, YOU HAVE CONFIGURED YOUR KUBERNETES ENVIRONMENT!
```
#### Rancher Dashboard
This URL will be of the Rancher server. Allows to create/update/delete environments, check on status of containers/services running on the cluster, and provides APIs and CLI details for the environments hosted on the cluster.
![Rancher Dashboard](img/rancher-dashboard.png "Rancher Dashboard")
#### Infrastructure Containers
This URL provides a list of containers and their status for the created Kubernetes environment.
![Infrastructure Containers](img/infrastructure-containers.png "Infrastructure Containers")
#### Kubernetes Dashboard
This is the URL for kubernetes dashboard which can be used to deploy modify and remove kubernetes services/deployments.
![Kubernetes Dashboard](img/Kubernetes-dashboard.png "Kubernetes Dashboard")
#### Kubernetes CLI
Kubernetes environments on Rancher provide a kubectl CLI config file which can be placed in ~/.kube/config file to connect the local kubectl CLI to the environment.
![Kubernetes CLI](img/kubernetes-cli.png "Kubernetes CLI")

<sub>For more details on Kubernetes Environments on Rancher, click [here](https://docs.rancher.com/rancher/v1.5/en/kubernetes/).</sub>

## Manual Setup
There are many ways to set up a Kubernetes Cluster. We are using Rancher as a kubernetes management platform. You can read more about Rancher and all it offers [here](http://rancher.com/rancher/).  
Rancher runs as a docker container and runs/manages a production ready kubernetes cluster by running kubernetes services as containers. This section provides the steps needed to manually set up a Kubernetes cluster with three nodes on Rancher.

We will provision 4 kvms. One kvm will be used as Rancher server, and three will be used as worker nodes by kubernetes cluster. For a simple architecture diagram, click [here](#architecture).

### Provision KVMs
Provision one kvm for KubeServer
```bash
triton instance create --wait --name=kubeserver -N Joyent-SDC-Public \
    ubuntu-certified-16.04 k4-highcpu-kvm-1.75G
```
This will be the host where Rancher server container runs.

Provision three kvms for KubeNodes
```bash
triton instance create --wait --name=kubenode1 -N Joyent-SDC-Public \
    ubuntu-certified-16.04 k4-highcpu-kvm-1.75G
triton instance create --wait --name=kubenode2 -N Joyent-SDC-Public \
    ubuntu-certified-16.04 k4-highcpu-kvm-1.75G
triton instance create --wait --name=kubenode3 -N Joyent-SDC-Public \
    ubuntu-certified-16.04 k4-highcpu-kvm-1.75G
```
These are provisioned to be kubernetes worker nodes.

### Allow root access to all KVMs:
JoyentCloud’s default KVM setup allows for login only as ubuntu user with `sudo` access. We need to setup root access with our ssh key. Copy the `authorized_keys` file from ubuntu user to root for all hosts.
```bash
kubeserver=$(triton ip kubeserver)
kubenode1=$(triton ip kubenode1)
kubenode2=$(triton ip kubenode2)
kubenode3=$(triton ip kubenode3)

for h in $kubeserver $kubenode1 $kubenode2 $kubenode3; do
    ssh ubuntu@$h sudo cp /home/ubuntu/.ssh/authorized_keys /root/.ssh/
done
```
Make sure all your KVMs have been created and are running:
```bash
triton ls
SHORTID   NAME        IMG                              STATE    FLAGS  AGE
abde0e87  kubeserver  ubuntu-certified-16.04@20170221  running  K      5m
e3fe229a  kubenode1   ubuntu-certified-16.04@20170221  running  K      2m
baa582d0  kubenode2   ubuntu-certified-16.04@20170221  running  K      1m
2077abe8  kubenode3   ubuntu-certified-16.04@20170221  running  K      1m
```
### Install pre-reqs and docker-engine package on all KVMs:
Rancher and all kubernetes services run as docker containers managed by Rancher so configure and installed docker-engine version 1.12.6 on all KVMs.
```bash
for h in $kubeserver $kubenode1 $kubenode2 $kubenode3; do
ssh root@$h \
  'apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y linux-image-extra-$(uname -r) && \
  apt-get install -y linux-image-extra-virtual zfs && \
  curl -fsSL https://apt.dockerproject.org/gpg |apt-key add - && \
  add-apt-repository "deb https://apt.dockerproject.org/repo/ ubuntu-$(lsb_release -cs) main" && \
  apt-get update && \
  apt-get -y install docker-engine=1.12.6-0~ubuntu-xenial'
done
```
### Start Rancher and setup Kubernetes environment and nodes
Start the rancher/server container on kubeserver KVM:
```bash
ssh root@$kubeserver docker run -d --restart=unless-stopped \
    -p 8080:8080 rancher/server
```

After the rancher/server docker container comes up, you should be able to access the Rancher UI and create a Kubernetes environment.   
Go to the Rancher UI http://$(triton ip kubeserver):8080/ and select “Manage Environments” from the Environments tab:
![Manage Environments](img/20170324a-manage-environments.png "Manage Environments")

Add a new environment:
![Add Environments](img/20170324a-add-environment.png "Add Environment")

Select Kubernetes from the list, provide a Name/Description and click create at the bottom of the page:
![Create Environments](img/20170324a-create-environment.png "Create Environment")

Now you should have a Kubernetes Environment which you can select from the “Environments” tab and add nodes to it by clicking “Add a host” button:
![Add Host](img/20170324a-add-host.png "Add Host")

From here you will add all three nodes (kubenode1 kubenode2 and kubenode3) by performing the same steps:
1. Select Custom from the available machine drivers list
1. Enter the ip address of kubenode
1. Copy the docker command and run it on the kubenode

![Add Host](img/20170324b-add-host.png "Add Host")

After the nodes have been added, kubernetes services will be started on each of the hosts and within minutes you will have your Kubernetes environment up and ready.

To deploy your app on your kubernetes environment, Rancher provides two simple options:
* kubectl config which can be copied from “KUBERNETES -> CLI” tab
* Kubernetes UI from “KUBERNETES -> Dashboard” tab

![Kubernetes Environment](img/20170324a-kubernetes-dashboard.png "Kubernetes Environment")
