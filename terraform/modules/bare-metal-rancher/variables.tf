variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_admin_password" {
  description = "The Rancher admin password"
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/19.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "rancher_server_image" {
  default     = "rancher/rancher:v2.4.4"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.4.4"
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

variable "ssh_user" {
  default     = "ubuntu"
  description = ""
}

variable "host" {
  description = ""
}

variable "bastion_host" {
  default     = ""
  description = ""
}

variable "key_path" {
  default     = "~/.ssh/id_rsa"
  description = ""
}
