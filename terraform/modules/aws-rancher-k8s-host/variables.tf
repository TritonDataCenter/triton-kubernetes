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
  default     = "rancher/rancher-agent:v2.1.7"
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
  default     = "https://raw.githubusercontent.com/mesoform/triton-kubernetes/master/scripts/docker/17.03.sh"
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

variable "aws_ami_id" {
  description = "Base AMI to launch the instances with"
}

variable "aws_instance_type" {
  default     = "t2.micro"
  description = "The AWS instance type to use for Kubernetes compute node(s). Defaults to t2.micro."
}

variable "aws_subnet_id" {
  description = "The AWS subnet id to deploy the instance to."
}

variable "aws_security_group_id" {
  description = "The AWS subnet id to deploy the instance to."
}

variable "aws_key_name" {
  description = "The AWS key name to use to deploy the instance."
}

variable "ebs_volume_device_name" {
  default     = ""
  description = "The EBS Device name"
}

variable "ebs_volume_mount_path" {
  default     = "/mnt/tk8s"
  description = "The EBS volume mount path"
}

variable "ebs_volume_type" {
  default     = "standard"
  description = "The EBS volume type. This can be gp2 for General Purpose SSD, io1 for Provisioned IOPS SSD, st1 for Throughput Optimized HDD, sc1 for Cold HDD, or standard for Magnetic volumes."
}

variable "ebs_volume_size" {
  default     = ""
  description = "The size of the volume, in GiBs."
}
