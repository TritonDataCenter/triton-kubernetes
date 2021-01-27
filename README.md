Triton Kubernetes is a multi-cloud Kubernetes solution. It has a global cluster manager (control plane) which can run and manage Kubernetes environments on any cloud - Public, Private or Bare Metal.

The cluster manager manages environments running on any region. AWS, Azure, Google and Triton (public and private) are supported. If you don't want to use a cloud, environments on bare metal servers and VMWare are supported as well.

View the [Quick Start Guide](docs/guide/) for installation instructions.

## Using The CLI

Triton Kubernetes allows you to create and destroy global cluster managers, kubernetes environments and individual cluster nodes. You can also get information on a cluster manager or kubernetes environment.

For help with a command, use the --help flag. For example:

```bash
$ triton-kubernetes --help
This is a multi-cloud Kubernetes solution. Triton Kubernetes has a global
cluster manager which can manage multiple clusters across regions/data-centers and/or clouds. 
Cluster manager can run anywhere (Triton/AWS/Azure/GCP/Baremetal) and manage Kubernetes environments running on any region of any supported cloud.
For an example set up, look at the How-To section.

Usage:
  triton-kubernetes [command]

Available Commands:
  create      Create resources
  destroy     Destroy cluster managers, kubernetes clusters or individual kubernetes cluster nodes.
  get         Display resource information
  help        Help about any command
  version     Print the version number of triton-kubernetes

Flags:
      --config string             config file (default is $HOME/.triton-kubernetes.yaml)
  -h, --help                      help for triton-kubernetes
      --non-interactive           Prevent interactive prompts
      --terraform-configuration   Create terraform configuration only
  -t, --toggle                    Help message for toggle

Use "triton-kubernetes [command] --help" for more information about a command.
```
