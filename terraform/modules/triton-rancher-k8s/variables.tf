variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_api_url" {
  description = "Rancher API url"
}

variable "rancher_access_key" {
  description = "Rancher API access key"
}

variable "rancher_secret_key" {
  description = "Rancher API access key."
}

variable k8s_version {
  default = "v1.19.4-rancher1-2"
}

variable k8s_network_provider {
  default = "flannel"
}

variable "rancher_registry" {
  default     = ""
  description = "The docker registry to use for Rancher images"
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
  description = "The docker registry to use for Kubernetes images"
}

variable "k8s_registry_username" {
  default     = ""
  description = "The username to login as."
}

variable "k8s_registry_password" {
  default     = ""
  description = "The password to use."
}

variable "triton_account" {
  default     = ""
  description = "The Triton account name, usually the username of your root user."
}

variable "triton_key_path" {
  default     = ""
  description = "The path to a private key that is authorized to communicate with the Triton API."
}

variable "triton_key_id" {
  default     = ""
  description = "The md5 fingerprint of the key at triton_key_path. Obtained by running `ssh-keygen -E md5 -lf ~/path/to.key`"
}

variable "triton_url" {
  default     = ""
  description = "The CloudAPI endpoint URL. e.g. https://us-west-1.api.joyent.com"
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/19.03.sh"
  description = "The URL to the shell script to install the docker engine."
}
