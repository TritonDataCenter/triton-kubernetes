provider "triton" {
  version = "~> 0.7.0"

  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
}

data "triton_image" "image" {
  name    = "${var.triton_image_name}"
  version = "${var.triton_image_version}"
}

locals {
  rancher_node_role = "${element(keys(var.rancher_host_labels), 0)}"
}

data "template_file" "install_rancher_agent" {
  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.hostname}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_api_url                    = "${var.rancher_api_url}"
    rancher_cluster_registration_token = "${var.rancher_cluster_registration_token}"
    rancher_cluster_ca_checksum        = "${var.rancher_cluster_ca_checksum}"
    rancher_node_role                  = "${local.rancher_node_role == "control" ? "controlplane" : local.rancher_node_role}"
    rancher_agent_image                = "${var.rancher_agent_image}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "triton_machine" "host" {
  package = "${var.triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.hostname}"

  user_script = "${data.template_file.install_rancher_agent.rendered}"

  networks = ["${data.triton_network.networks.*.id}"]

  cns = {
    services = ["${element(keys(var.rancher_host_labels), 0)}.${var.hostname}"]
  }

  affinity = ["role!=~${element(keys(var.rancher_host_labels), 0)}"]

  tags = {
    role = "${element(keys(var.rancher_host_labels), 0)}"
  }
}
