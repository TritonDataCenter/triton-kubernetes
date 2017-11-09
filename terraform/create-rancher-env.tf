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

  triton_account       = "<set triton_account"
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
}

module "azure_example" {
  source = "./modules/rancher-k8s"

  api_url    = "http://${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
  access_key = ""
  secret_key = ""

  name = "azure-example"

  etcd_node_count          = "3"
  orchestration_node_count = "3"
  compute_node_count       = "3"

  azure = "true"
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
