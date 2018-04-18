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

variable "k8s_version" {
  default = "v1.9.5-rancher1-1"
}

variable "k8s_network_provider" {
  default = "flannel"
}

variable "vsphere_user" {
  description = "The username of the vCenter Server user."
}

variable "vsphere_password" {
  description = "The password of the vCenter Server user."
}

variable "vsphere_server" {
  description = "The IP address or FQDN of the vCenter Server."
}

variable "vsphere_datacenter_name" {
  description = "Name of the datacenter to use."
}

variable "vsphere_datastore_name" {
  description = "Name of the datastore to use."
}

variable "vsphere_resource_pool_name" {
  description = "Name of the resource pool to use."
}

variable "vsphere_network_name" {
  description = "Name of the network to use."
}
