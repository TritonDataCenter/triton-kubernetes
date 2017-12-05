k8s-triton-supervisor packer
=========================

Build
1. Create a symbolic link for the variable file you want to use. `ln -s triton-us-west-1.yaml triton.yaml`
1. For each of the yaml files:
	1. Convert the yaml into packer json `./packer-config rancher-host.yaml > rancher-host.json`
	1. Build `packer build rancher-host.json`
