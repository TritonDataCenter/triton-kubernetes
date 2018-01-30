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

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
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

variable "triton_network_names" {
  type        = "list"
  description = "List of Triton network names that the node(s) should be attached to."

  default = [
    "Joyent-SDC-Public",
  ]
}

variable "triton_image_name" {
  default     = "ubuntu-certified-16.04"
  description = "The name of the Triton image to use."
}

variable "triton_image_version" {
  default     = "20170619.1"
  description = "The version/tag of the Triton image to use."
}

variable "triton_ssh_user" {
  default     = "ubuntu"
  description = "The ssh user to use."
}

variable "triton_machine_package" {
  default     = "k4-highcpu-kvm-1.75G"
  description = "The Triton machine package to use for this host. Defaults to k4-highcpu-kvm-1.75G."
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
