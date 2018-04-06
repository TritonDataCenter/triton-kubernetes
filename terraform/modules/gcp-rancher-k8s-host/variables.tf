variable "hostname" {
  description = ""
}

variable "rancher_api_url" {
  description = ""
}

variable "rancher_access_key" {
  description = ""
}

variable "rancher_secret_key" {
  description = ""
}

variable "rancher_environment_id" {
  description = "The ID of the Rancher environment this host should register itself to"
}

variable "rancher_host_labels" {
  type        = "map"
  description = "A map of key/value pairs that get passed to the rancher agent on the host."
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
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
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
