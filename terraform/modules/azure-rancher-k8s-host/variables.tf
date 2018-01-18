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
  default     = "https://releases.rancher.com/install-docker/1.12.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "azure_subscription_id" {}

variable "azure_client_id" {}

variable "azure_client_secret" {}

variable "azure_tenant_id" {}

variable "azure_environment" {
  default = "public"
}

variable "azure_location" {}

variable "azure_resource_group_name" {}

variable "azure_network_security_group_id" {}

variable "azure_subnet_id" {}

variable "azure_size" {
  default = "Standard_A0"
}

variable "azure_image_publisher" {
  default = "Canonical"
}

variable "azure_image_offer" {
  default = "UbuntuServer"
}

variable "azure_image_sku" {
  default = "16.04-LTS"
}

variable "azure_image_version" {
  default = "latest"
}

variable "azure_ssh_user" {
  default = "root"
}

variable "azure_public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}
