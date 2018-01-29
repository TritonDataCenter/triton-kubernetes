triton-kubernetes terraform
============================

Overview
-   A set of terraform modules that provisions Rancher Masters and Hosts.

Getting Started
-   Checkout this project
-   Modify create-rancher.tf or create your own
-   `terraform init`
-   `terraform apply -target module.create_rancher` # Create Rancher Cluster
-   `terraform apply -target module.triton_example` # Create Rancher Environment
