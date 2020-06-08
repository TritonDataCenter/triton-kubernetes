variable "hostname" {
  description = ""
}

variable "rancher_api_url" {
  description = ""
}

variable "rancher_cluster_registration_token" {
}

variable "rancher_cluster_ca_checksum" {
}

variable "rancher_host_labels" {
  type        = map(string)
  description = "A map of key/value pairs that get passed to the rancher agent on the host."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.4.4"
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

variable "vsphere_template_name" {
  description = "VM template to use."
}

variable "ssh_user" {
  default     = "ubuntu"
  description = ""
}

variable "key_path" {
  default     = "~/.ssh/id_rsa"
  description = ""
}

