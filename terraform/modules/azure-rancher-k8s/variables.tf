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

variable "k8s_plane_isolation" {
  default     = "none"
  description = "Plane isolation of the Kubernetes cluster. required or none"
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

variable "azure_ssh_user" {
  default = "root"
}

variable "azure_public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "azure_resource_group_name" {
  default = "k8s"
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

variable "etcd_azure_size" {
  default = "Standard_A0"
}

variable "orchestration_azure_size" {
  default = "Standard_A0"
}

variable "compute_azure_size" {
  default = "Standard_A0"
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

variable "k8s_registry" {
  default     = ""
  description = "The docker registry to use for k8s images"
}

variable "k8s_registry_username" {
  default     = ""
  description = "The username to login as."
}

variable "k8s_registry_password" {
  default     = ""
  description = "The password to use."
}
