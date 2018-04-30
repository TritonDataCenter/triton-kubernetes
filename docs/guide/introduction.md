# Introduction

This document describes the process of setting up Triton Multi-Cloud Kubernetes.

![Triton-Kubernetes](https://github.com/joyent/triton-kubernetes/raw/master/docs/imgs/Triton-Kubernetes.png)


## Overview

`triton-kubernetes` interactive cli is one of two ways you can interact with Triton Multi-Cloud Kubernetes. Using `triton-kubernetes` cli, you can:

- create a cluster manager
- destroy a cluster manager and all clusters it is managing
- add/remove a cluster to/from an existing cluster manager
- backup/restore a kubernetes namespace from any of your clusters to manta/S3
- query your existing cluster managers and clusters

The cli `triton-kubernetes` allows for creating and managing a kubernetes deployment only. Application deployments will still need to be done using `kubectl`.

> <sub>Note: Keep in mind that every cloud has a resource quota. If that quota has been reached, Triton-Kubernetes will not be able to provision new machines and throw errors.</sub>

## Working with the CLI

To get help on a command, use the --help flag. For example:

```
$ triton-kubernetes --help
This is a multi-cloud Kubernetes solution. Triton Kubernetes has a global
cluster manager which will run on Triton and manages Kubernetes environments. This
cluster manager will manage environments running on any region of any supported cloud.
For an example set up, look at the How-To section.

Usage:
  triton-kubernetes [command]

Available Commands:
  create      Create cluster managers, kubernetes clusters or individual kubernetes cluster nodes.
  destroy     Destroy cluster managers, kubernetes clusters or individual kubernetes cluster nodes.
  get         Display resource information
  help        Help about any command
  version     Print the version number of triton-kubernetes

Flags:
      --config string     config file (default is $HOME/.triton-kubernetes.yaml)
  -h, --help              help for triton-kubernetes
      --non-interactive   Prevent interactive prompts
  -t, --toggle            Help message for toggle

Use "triton-kubernetes [command] --help" for more information about a command.
```
