variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "triton_account" {
  description = "The Triton account name, usually the username of your root user."
}

variable "triton_key_path" {
  description = "The path to a private key that is authorized to communicate with the Triton API."
}

variable "triton_key_id" {
  description = "The md5 fingerprint of the key at triton_key_path. Obtained by running `ssh-keygen -E md5 -lf ~/path/to.key`"
}

variable "triton_url" {
  description = "The CloudAPI endpoint URL. e.g. https://us-west-1.api.joyent.com"
}

variable "triton_network_names" {
  type        = "list"
  description = "List of Triton network names that the node(s) should be attached to."
}

variable "triton_image_name" {
  description = "The name of the Triton image to use."
}

variable "triton_image_version" {
  description = "The version/tag of the Triton image to use."
}

variable "triton_ssh_user" {
  default     = "root"
  description = "The ssh user to use."
}

variable "master_triton_machine_package" {
  description = "The Triton machine package to use for Rancher master node(s). e.g. k4-highcpu-kvm-1.75G"
}

variable "docker_engine_install_url" {
  default     = "https://releases.rancher.com/install-docker/1.12.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "docker_machine_driver_triton_url" {
  default     = "https://github.com/nimajalali/docker-machine-driver-triton/releases/download/v0.0.5/docker-machine-driver-triton"
  description = "A URL to a binary that implements the docker-machine-driver interface for Triton."
}

variable "docker_machine_driver_triton_checksum" {
  default     = "b0ed3600c2d136c788eeac47d595ef9c"
  description = "The md5 checksum for the docker-machine-driver-triton binary."
}

variable "rancher_ui_driver_triton" {
  default     = "https://s3-us-west-1.amazonaws.com/static.nimajalali.com/ui-driver-triton/component.js"
  description = "A URL to the Rancher UI driver for Triton."
}

variable "ha" {
  default     = false
  description = "Should Rancher be deployed in HA, if true a mysqldb node and 2 Rancher master nodes will be created."
}

variable "mysqldb_triton_machine_package" {
  default     = ""
  description = "The Triton machine package to use for the Rancher mysqldb node. Defaults to master_triton_machine_package."
}

variable "mysqldb_port" {
  default     = "3306"
  description = "The port to host mysqldb on."
}

variable "mysqldb_username" {
  default     = "cattle"
  description = "The username that will be setup for Rancher to connect to mysqldb."
}

variable "mysqldb_password" {
  default     = "cattle"
  description = "The password that will be setup for Rancher to connect to mysqldb."
}

variable "mysqldb_database_name" {
  default     = "cattle"
  description = "The database name that will be setup for Rancher to connect to mysqldb."
}
