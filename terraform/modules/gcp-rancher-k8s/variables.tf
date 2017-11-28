variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "api_url" {
  description = ""
}

variable "access_key" {
  description = ""
}

variable "secret_key" {
  description = ""
}

variable "k8s_plane_isolation" {
  default     = "none"
  description = "Plane isolation of the Kubernetes cluster. required or none"
}

variable "etcd_node_count" {
  default     = "3"
  description = "The number of etcd node(s) to initialize in the Kubernetes cluster."
}

variable "orchestration_node_count" {
  default     = "3"
  description = "The number of orchestration node(s) to initialize in the Kubernetes cluster."
}

variable "compute_node_count" {
  default     = "3"
  description = "The number of compute node(s) to initialize in the Kubernetes cluster."
}

variable "docker_engine_install_url" {
  default     = "https://releases.rancher.com/install-docker/1.12.sh"
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

variable  etcd_gcp_instance_type {
  default = "n1-standard-1"
  description = "GCP machine type to launch the etcd instance with"
}

variable  orchestration_gcp_instance_type {
  default = "n1-standard-1"
  description = "GCP machine type to launch the orchestration instance with"
} 

variable  compute_gcp_instance_type {
  default = "n1-standard-1"
  description = "GCP machine type to launch the compute instance with"
} 

variable "gcp_instance_zone" {
  description = "Zone to deploy GCP machine in"
}

variable "gcp_image" {
  description = "GCP image to be used for instance"
  default = "ubuntu-1604-xenial-v20171121a"
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

variable "k8s_registry" {
  default     = ""
  description = "The docker registry to use for k8s images"
}

variable "k8s_registry_username" {
  default     = ""
  description = "The username to login as."
}

variable "k8s_registry_password" {
  default     = ""
  description = "The password to use."
}
