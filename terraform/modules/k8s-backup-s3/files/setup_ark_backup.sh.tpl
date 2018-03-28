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

# Creating the credentials-ark file
cat > credentials-ark <<EOF
[default]
aws_access_key_id=${aws_access_key}
aws_secret_access_key=${aws_secret_key}
EOF

# Applying the heptio ark prereqs
./kubectl apply -f ark-0.7.1/examples/common/00-prereqs.yaml --kubeconfig=kubeconfig.yaml

# Creating the generic secret cloud-credentials 
ARK_SERVER_NAMESPACE=heptio-ark-server
./kubectl create secret generic cloud-credentials \
    --namespace $ARK_SERVER_NAMESPACE \
    --from-file cloud=credentials-ark \
    --kubeconfig=kubeconfig.yaml

# Injecting AWS bucket and region into ark configuration
sed -i 's/<YOUR_BUCKET>/${aws_s3_bucket}/g' ark-0.7.1/examples/aws/00-ark-config.yaml
sed -i 's/<YOUR_REGION>/${aws_region}/g' ark-0.7.1/examples/aws/00-ark-config.yaml

# TODO examples/common/10-deployment.yaml - Make sure that spec.template.spec.containers[*].env.name is "AWS_SHARED_CREDENTIALS_FILE".

# Install Ark with S3 Configuration
./kubectl apply -f ark-0.7.1/examples/aws/00-ark-config.yaml --kubeconfig=kubeconfig.yaml
./kubectl apply -f ark-0.7.1/examples/common/10-deployment.yaml --kubeconfig=kubeconfig.yaml

# Cleaning up
rm -rf ark ark-* credentials-ark kubeconfig.yaml kubectl v0.7.1.tar.gz
