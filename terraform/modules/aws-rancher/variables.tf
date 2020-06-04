variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_admin_password" {
  description = "The Rancher admin password"
}

variable "docker_engine_install_url" {
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/19.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "rancher_server_image" {
  default     = "rancher/rancher:v2.3.3"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = "rancher/rancher-agent:v2.3.3"
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
}

variable "aws_instance_type" {
  default     = "t2.micro"
  description = "The AWS instance type to use for Kubernetes compute node(s). Defaults to t2.micro."
}

variable "aws_key_name" {
  description = "The AWS key name to use to deploy the instance."
}

variable "aws_public_key_path" {
  description = "Path to a public key. If set, a key_pair will be made in AWS named aws_key_name"
  default     = "~/.ssh/id_rsa.pub"
}

variable "aws_private_key_path" {
  description = "Path to a private key."
  default     = "~/.ssh/id_rsa"
}

variable "aws_ssh_user" {
  default     = "ubuntu"
  description = "The ssh user to use."
}
