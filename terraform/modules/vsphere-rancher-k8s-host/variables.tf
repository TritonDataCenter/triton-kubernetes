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