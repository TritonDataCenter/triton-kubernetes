variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_admin_password" {
  description = "The Rancher admin password"
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/mesoform/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "rancher_server_image" {
  default     = "rancher/rancher:v2.1.7"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.1.7"
  description = "The Rancher Agent image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_registry" {
  default     = ""
  description = "The docker registry to use for rancher server and agent images"
}

variable "rancher_registry_username" {
  default     = ""
  description = "The username to login as."
}

variable "rancher_registry_password" {
  default     = ""
  description = "The password to use."
}

variable "gcp_path_to_credentials" {
  description = "Location of GCP JSON credentials file."
}

variable "gcp_compute_region" {
  description = "GCP region to host your network"
}

variable "gcp_project_id" {
  description = "GCP project ID that will be running the instances and managing the network"
}

variable gcp_machine_type {
  default     = "n1-standard-1"
  description = "GCP machine type to launch the instance with"
}

variable "gcp_instance_zone" {
  description = "Zone to deploy GCP machine in"
}

variable "gcp_image" {
  description = "GCP image to be used for instance"
  default     = "ubuntu-1604-xenial-v20171121a"
}

variable "gcp_ssh_user" {
  default     = "ubuntu"
  description = "The ssh user to use."
}

variable "gcp_public_key_path" {
  description = "Path to a public key."
  default     = "~/.ssh/id_rsa.pub"
}

variable "gcp_private_key_path" {
  description = "Path to a private key."
  default     = "~/.ssh/id_rsa"
}
