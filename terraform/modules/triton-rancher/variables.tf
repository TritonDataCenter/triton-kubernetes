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
  default     = "rancher/rancher:v2.4.11"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.4.11"
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

variable "triton_account" {
  description = "The Triton account name, usually the username of your root user."
}

variable "triton_key_path" {
  description = "The path to a private key that is authorized to communicate with the Triton API."
}

variable "triton_key_id" {
  description = "The md5 fingerprint of the key at triton_key_path. Obtained by running `ssh-keygen -E md5 -lf ~/path/to.key`"
}

variable "triton_url" {
  description = "The CloudAPI endpoint URL. e.g. https://us-west-1.api.joyent.com"
}

variable "triton_network_names" {
  type        = list
  description = "List of Triton network names that the node(s) should be attached to."
}

variable "triton_image_name" {
  description = "The name of the Triton image to use."
}

variable "triton_image_version" {
  description = "The version/tag of the Triton image to use."
}

variable "triton_ssh_user" {
  default     = "root"
  description = "The ssh user to use."
}

variable "master_triton_machine_package" {
  description = "The Triton machine package to use for Rancher master node(s). e.g. k4-highcpu-kvm-1.75G"
}
