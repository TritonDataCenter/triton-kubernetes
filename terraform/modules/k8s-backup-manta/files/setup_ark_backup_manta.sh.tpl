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

# Saving triton private key to file system
echo "${triton_private_key}" > ${triton_key_path}

# Saving minio_manta_deployment.yaml with Manta Gateway configuration
echo '${minio_manta_yaml_filedata}' > minio_manta_deployment.yaml

# Install Ark and Minio on cluster
./kubectl apply -f ark-0.7.1/examples/common/00-prereqs.yaml --kubeconfig=kubeconfig.yaml
./kubectl apply -f minio_manta_deployment.yaml --kubeconfig=kubeconfig.yaml
./kubectl apply -f ark-0.7.1/examples/common/10-deployment.yaml --kubeconfig=kubeconfig.yaml

# Cleaning up
rm -rf ark ark-* ${triton_key_path} minio_manta_deployment.yaml kubeconfig.yaml kubectl v0.7.1.tar.gz
