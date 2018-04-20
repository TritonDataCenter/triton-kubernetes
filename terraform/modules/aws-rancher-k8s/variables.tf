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
  default = "v1.9.5-rancher1-1"
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

variable "aws_access_key" {
  description = "AWS access key"
}

variable "aws_secret_key" {
  description = "AWS secret access key"
}

variable "aws_region" {
  description = "AWS region to host your network"
}

variable "aws_vpc_cidr" {
  description = "CIDR for VPC"
  default     = "10.0.0.0/16"
}

variable "aws_subnet_cidr" {
  description = "CIDR for subnet"
  default     = "10.0.2.0/24"
}

variable "aws_ami_id" {
  description = "Base AMI to launch the instances with"
  default     = ""
}

variable "aws_public_key_path" {
  description = "Path to a public key. If set, a key_pair will be made in AWS named aws_key_name"
  default     = ""
}

variable "aws_key_name" {
  description = "Name of the public key to be used for provisioning"
}
