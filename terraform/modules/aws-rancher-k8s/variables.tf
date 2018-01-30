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
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
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

variable "etcd_aws_instance_type" {
  default     = "t2.micro"
  description = "The AWS instance type to use for Kubernetes etcd node(s). Defaults to t2.micro."
}

variable "orchestration_aws_instance_type" {
  default     = "t2.micro"
  description = "The AWS instance type to use for Kubernetes orchestration node(s). Defaults to t2.micro."
}

variable "compute_aws_instance_type" {
  default     = "t2.micro"
  description = "The AWS instance type to use for Kubernetes compute node(s). Defaults to t2.micro."
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
