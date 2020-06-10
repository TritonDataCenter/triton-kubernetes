variable "name" {
  description = "Human readable name used as prefix to generated names."
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

variable "gcp_path_to_credentials" {
  description = "Location of GCP JSON credentials file."
}

variable "gcp_project_id" {
  description = "GCP project ID that will be running the instances and managing the network"
}

variable "gcp_compute_region" {
  description = "GCP region to host your network"
}

variable "gcp_zone" {
  description = "Zone to deploy GKE cluster in"
}

variable "gcp_additional_zones" {
  type        = list(string)
  description = "Zones to deploy GKE cluster nodes in"
}

variable "gcp_machine_type" {
  default     = "n1-standard-1"
  description = "GCP machine type to launch the instance with"
}

variable "k8s_version" {
  default = "1.8.8-gke.0"
}

variable "node_count" {
}

variable "password" {
}

