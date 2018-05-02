variable "hostname" {
  description = ""
}

variable "rancher_api_url" {
  description = ""
}

variable "rancher_cluster_registration_token" {}

variable "rancher_cluster_ca_checksum" {}

variable "rancher_host_labels" {
  type        = "map"
  description = "A map of key/value pairs that get passed to the rancher agent on the host."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.0.0"
  description = "The Rancher Agent image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
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
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
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

variable "azure_disk_mount_path" {
  default = ""
}

variable "azure_disk_size" {
  default = ""
}
