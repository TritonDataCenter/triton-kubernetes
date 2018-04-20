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
  default     = "rancher/agent:v2.0.0-beta2"
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
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "ssh_user" {
  default = "ubuntu"
  description = ""
}

variable "host" {
  description = ""
}

variable "bastion_host" {
  default = ""
  description = ""
}

variable "key_path" {
  default = "~/.ssh/id_rsa"
  description = ""
}