triton-rancher-k8s terraform
============================

Overview
-   A terraform module that provisions a Rancher master on Triton that is capable of provisioning hosts on Triton.

Getting Started
-   Checkout this project
-   Modify create-rancher.tf or create your own
-   `terraform init`
-   `terraform apply -target module.create_rancher` # Create Rancher Cluster
-   `terraform apply -target module.triton_example` # Create Rancher Environment
