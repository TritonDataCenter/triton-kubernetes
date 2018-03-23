#!/bin/bash

# Download kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
sudo chmod +x ./kubectl

# Download ark 0.7.1
curl -LO https://github.com/heptio/ark/releases/download/v0.7.1/ark-v0.7.1-linux-arm64.tar.gz
tar xvf ark-v0.7.1-linux-arm64.tar.gz

# Download .yaml files from github
curl -LO https://github.com/heptio/ark/archive/v0.7.1.tar.gz
tar xvf v0.7.1.tar.gz

# Saving kubeconfig file to file system
echo "${kubeconfig_filedata}" > kubeconfig.yaml

# Install Ark and Minio on cluster
./kubectl apply -f ark-0.7.1/examples/common/00-prereqs.yaml --kubeconfig=kubeconfig.yaml
./kubectl apply -f ark-0.7.1/examples/minio/ --kubeconfig=kubeconfig.yaml
./kubectl apply -f ark-0.7.1/examples/common/10-deployment.yaml --kubeconfig=kubeconfig.yaml
