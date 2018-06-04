variable "rancher_api_url" {}

variable "rancher_access_key" {}

variable "rancher_secret_key" {}

variable "rancher_cluster_id" {}

variable "triton_key_path" {
  default     = ""
  description = "The path to a private key that is authorized to communicate with the triton_ssh_host."
}

variable "triton_account" {
  description = "The Triton account name, usually the username of your root user."
}

variable "triton_key_id" {
  description = "The md5 fingerprint of the key at triton_key_path. Obtained by running `ssh-keygen -E md5 -lf ~/path/to.key`"
}

variable "manta_subuser" {
  default     = ""
  description = "The Manta subuser"
}
