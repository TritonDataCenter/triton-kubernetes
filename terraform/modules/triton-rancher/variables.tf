variable "name" {
  description = "Human readable name used as prefix to generated names."
}

variable "rancher_admin_username" {
  description = "The Rancher admin username"
}

variable "rancher_admin_password" {
  description = "The Rancher admin password"
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

variable "triton_mysql_image_name" {
  default     = "ubuntu-certified-16.04"
  description = "The name of the Triton image to use."
}

variable "triton_mysql_image_version" {
  default     = "20170619.1"
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
  default     = "https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh"
  description = "The URL to the shell script to install the docker engine."
}

variable "ha" {
  default     = false
  description = "Should Rancher be deployed in HA, if true a mysqldb node and 2 Rancher master nodes will be created."
}

variable "gcm_node_count" {
  default     = "1"
  description = "Number of Global Cluster Managers to cluster."
}

variable "gcm_private_network_name" {
  default     = "Joyent-SDC-Private"
  description = "Should Rancher be deployed in HA, this network will contain the mysqldb, Rancher master, nginx proxy, and bastion nodes. In non-HA mode, this will be ignored."
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

variable "rancher_server_image" {
  default     = "rancher/server:v1.6.14"
  description = "The Rancher Server image to use, can be a url to a private registry leverage docker_login_* variables to authenticate to registry."
}

variable "rancher_agent_image" {
  default     = ""
  description = "If set will pass the CATTLE_BOOTSTRAP_REQUIRED_IMAGE environment variable to the Rancher Server start command."
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

variable "rancher_tls_private_key_path" {
  default = ""
  description = "The path to the TLS private key"
}

variable "rancher_tls_cert_path" {
  default = ""
  description = "The path to the TLS certificate"
}

variable "rancher_domain_name" {
  default = ""
  description = "When a TLS/SSL certificate has been provided, this should be set to the domain name that is compatible with the certificate"
}
