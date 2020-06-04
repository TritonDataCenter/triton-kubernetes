variable "hostname" {
  description = ""
}

variable "rancher_api_url" {
  description = ""
}

variable "rancher_cluster_registration_token" {}

variable "rancher_cluster_ca_checksum" {}

variable "rancher_host_labels" {
  type        = "map"
  description = "A map of key/value pairs that get passed to the rancher agent on the host."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.3.3"
  description = "The Rancher Agent image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_registry" {
  default     = ""
  description = "The docker registry to use for rancher images"
}

variable "rancher_registry_username" {
  default     = ""
  description = "The username to login as."
}

variable "rancher_registry_password" {
  default     = ""
  description = "The password to use."
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/19.03.sh"
  description = "The URL to the shell script to install the docker engine."
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

variable "gcp_compute_network_name" {
  description = "Network to deploy GCP machine in"
}

variable "gcp_compute_firewall_host_tag" {
  description = "Tag that should be applied to nodes so the firewall source rules can be applied"
}

variable "gcp_disk_type" {
  default     = ""
  description = "The disk type which can be either 'pd-ssd' for SSD or 'pd-standard' for Standard"
}

variable "gcp_disk_size" {
  default     = ""
  description = "The disk size"
}

variable "gcp_disk_mount_path" {
  default     = ""
  description = "The mount path"
}
