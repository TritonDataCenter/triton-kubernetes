variable "name" {
  description = "Human readable name used as prefix to generated names."
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

variable "azure_subscription_id" {}

variable "azure_client_id" {}

variable "azure_client_secret" {}

variable "azure_tenant_id" {}

variable "azure_environment" {
  default = "public"
}

variable "azure_location" {}

variable "azure_size" {}

variable "azure_ssh_user" {
  default = "root"
}

variable "azure_public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "k8s_version" {
  default = "1.9.6"
}

variable "node_count" {}
