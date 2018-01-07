terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}

module "create_rancher-example" {
  source = "./modules/triton-rancher"

  triton_network_names = ["Joyent-SDC-Public", "Joyent-SDC-Private"]

  triton_image_name    = "ubuntu-certified-16.04"
  triton_image_version = "20170619.1"
  triton_ssh_user      = "ubuntu"

  triton_account                 = "fayazg"
  triton_key_path                = "/Users/fayazg/.ssh/demo_id_rsa"
  triton_key_id                  = "2c:53:bc:63:97:9e:79:3f:91:35:5e:f4:c8:23:88:37"
  triton_url                     = "https://us-east-1.api.joyent.com"
  name                           = "global-cluster"
  master_triton_machine_package  = "k4-highcpu-kvm-1.75G"
  mysqldb_triton_machine_package = "k4-highcpu-kvm-1.75G"
  ha                             = true
  gcm_node_count                 = "2"

  # rancher_server_image      = "docker-registry.joyent.com:5000/rancher/server:v1.6.10"
  # rancher_agent_image       = "docker-registry.joyent.com:5000/rancher/agent:v1.2.6"

  # rancher_registry          = "docker-registry.joyent.com:5000"
  # rancher_registry_username = "username"
  # rancher_registry_password = "password"
}

output "masters" {
  value = "${module.create_rancher.masters}"
}
