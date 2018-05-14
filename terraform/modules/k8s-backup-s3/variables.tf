variable "rancher_api_url" {}

variable "rancher_access_key" {}

variable "rancher_secret_key" {}

variable "rancher_cluster_id" {}

variable "aws_access_key" {
  default     = ""
  description = "AWS access key"
}

variable "aws_secret_key" {
  default     = ""
  description = "AWS secret key"
}

variable "aws_region" {
  default     = ""
  description = "AWS region where the Heptio ARK backup will be stored."
}

variable "aws_s3_bucket" {
  default     = ""
  description = "Name of the AWS bucket where the Heptio ARK backup will be stored."
}
