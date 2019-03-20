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

variable k8s_version {
  default = "v1.13.4-rancher1-1"
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
  default = "k8s-network"
}

variable "azure_virtual_network_address_space" {
  default = "10.0.0.0/16"
}

variable "azure_subnet_name" {
  default = "k8s-subnet"
}

variable "azure_subnet_address_prefix" {
  default = "10.0.2.0/24"
}

variable "azurerm_network_security_group_name" {
  default = "k8s-firewall"
}
