terraform {
  backend "local" {
    path       = "terraform.tfstate"
  }
}

module "create_rancher" {
  source = "./modules/triton-rancher"

  triton_network_names = [
    "Joyent-SDC-Public",
    "Joyent-SDC-Private",
  ]

  triton_image_name    = "ubuntu-certified-16.04"
  triton_image_version = "20170619.1"
  triton_ssh_user      = "ubuntu"

  # triton_image_name    = "centos-7"
  # triton_image_version = "20170327"
  # triton_ssh_user      = "root"

  triton_account  = "<set triton_account>"
  triton_key_path = "<set triton_key_path>"
  triton_key_id   = "<set triton_key_id>"
  triton_url      = "<set triton_url>"
  name = "<set name>"
  master_triton_machine_package = "<set master_triton_machine_package>"
  ha                            = "<set ha>"
}

output "masters" {
  value = "${module.create_rancher.masters}"
}
