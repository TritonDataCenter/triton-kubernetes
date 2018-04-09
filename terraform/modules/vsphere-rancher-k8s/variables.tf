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

variable "vsphere_user" {

}

variable "vsphere_password" {

}

variable "vsphere_server" {

}

variable "vsphere_datacenter_name" {

}

variable "vsphere_datastore_name" {

}

variable "vsphere_resource_pool_name" {

}

variable "vsphere_network_name" {
  
}
