variable "hostname" {
  description = "hostname of the host to be added"
}

variable "networks" {
  type        = "list"
  description = "list of networks"
}

variable "root_authorized_keys" {
  default     = "~/.ssh/id_rsa"
  description = "public ssh key for root login"
}

variable "image" {
  default     = "0867ef86-e69d-4aaa-ba3b-8d2aef0c204e"
  description = "long image ID, default is ubuntu-certified-16.04"
}

variable "package" {
  default     = "k4-highcpu-kvm-1.75G"
  description = "triton package, default is k4-highcpu-kvm-1.75G"
}
