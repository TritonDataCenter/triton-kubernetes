variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_admin_password" {
  description = "The Rancher admin password"
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/mesoform/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "rancher_server_image" {
  default     = "rancher/rancher:v2.1.7"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.1.7"
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

variable "azure_subscription_id" {
  default = ""
}

variable "azure_client_id" {
  default = ""
}

variable "azure_client_secret" {
  default = ""
}

variable "azure_tenant_id" {
  default = ""
}

variable "azure_environment" {
  default = "public"
}

variable "azure_location" {
  default = "West US 2"
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

variable "azure_virtual_network_name" {
  default = "rancher-network"
}

variable "azure_virtual_network_address_space" {
  default = "10.0.0.0/16"
}

variable "azure_subnet_name" {
  default = "rancher-subnet"
}

variable "azure_subnet_address_prefix" {
  default = "10.0.2.0/24"
}

variable "azurerm_network_security_group_name" {
  default = "rancher-firewall"
}

variable "azure_resource_group_name" {}

variable "azure_size" {
  default = "Standard_A0"
}

variable "azure_ssh_user" {
  default = "ubuntu"
}

variable "azure_public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "azure_private_key_path" {
  default = "~/.ssh/id_rsa"
}
