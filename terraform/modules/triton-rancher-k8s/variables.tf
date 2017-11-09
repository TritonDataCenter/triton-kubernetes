variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "api_url" {
  description = ""
}

variable "access_key" {
  description = ""
}

variable "secret_key" {
  description = ""
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

variable "docker_engine_install_url" {
  default     = "https://releases.rancher.com/install-docker/1.12.sh"
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

variable "etcd_triton_machine_package" {
  default     = "k4-highcpu-kvm-1.75G"
  description = "The Triton machine package to use for Kubernetes etcd node(s). Defaults to k4-highcpu-kvm-1.75G."
}

variable "orchestration_triton_machine_package" {
  default     = "k4-highcpu-kvm-1.75G"
  description = "The Triton machine package to use for Kubernetes orchestration node(s). Defaults to k4-highcpu-kvm-1.75G."
}

variable "compute_triton_machine_package" {
  default     = "k4-highcpu-kvm-1.75G"
  description = "The Triton machine package to use for Kubernetes compute node(s). Defaults to k4-highcpu-kvm-1.75G."
}
