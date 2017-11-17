data "terraform_remote_state" "rancher" {
  backend = "local"

  config {
    path = "terraform.tfstate"
  }
}

module "triton_example" {
  source = "./modules/triton-rancher-k8s"

  api_url    = "http://${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
  access_key = ""
  secret_key = ""

  name = "triton-example"

  etcd_node_count          = "3"
  orchestration_node_count = "3"
  compute_node_count       = "<set compute_node_count>"

  k8s_plane_isolation = "required"

  triton_account       = "<set triton_account>"
  triton_key_path      = "<set triton_key_path>"
  triton_key_id        = "<set triton_key_id>"
  triton_url           = "<set triton_url>"
  triton_image_name    = "ubuntu-certified-16.04"
  triton_image_version = "20170619.1"
  triton_ssh_user      = "ubuntu"

  triton_network_names = [
    "Joyent-SDC-Public",
    "Joyent-SDC-Private",
  ]

  etcd_triton_machine_package          = "<set etcd_triton_machine_package>"
  orchestration_triton_machine_package = "<set orchestration_triton_machine_package>"
  compute_triton_machine_package       = "<set compute_triton_machine_package>"

  # rancher_registry          = "docker-registry.joyent.com:5000"
  # rancher_registry_username = "username"
  # rancher_registry_password = "password"

  # k8s_registry          = "docker-registry.joyent.com:5000"
  # k8s_registry_username = "username"
  # k8s_registry_password = "password"
}

module "azure_example" {
  source = "./modules/azure-rancher-k8s"

  api_url    = "http://${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
  access_key = ""
  secret_key = ""

  name = "azure-example"

  etcd_node_count          = "3"
  orchestration_node_count = "3"
  compute_node_count       = "3"

  k8s_plane_isolation = "none"

  azure_subscription_id = ""
  azure_client_id       = ""
  azure_client_secret   = ""
  azure_tenant_id       = ""

  azure_location = "West US 2"

  azure_ssh_user        = "ubuntu"
  azure_public_key_path = "~/.ssh/id_rsa.pub"

  etcd_azure_size          = "Standard_A1"
  orchestration_azure_size = "Standard_A1"
  compute_azure_size       = "Standard_A1"

  # rancher_registry          = "docker-registry.joyent.com:5000"
  # rancher_registry_username = "username"
  # rancher_registry_password = "password"

  # k8s_registry          = "docker-registry.joyent.com:5000"
  # k8s_registry_username = "username"
  # k8s_registry_password = "password"
}

module "gcp_example" {
  source = "./modules/rancher-k8s"

  api_url    = "http://${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
  access_key = ""
  secret_key = ""

  name = "gcp-example"

  etcd_node_count          = "3"
  orchestration_node_count = "3"
  compute_node_count       = "3"

  gcp = "true"
}
