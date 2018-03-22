provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

provider "rancher" {
  api_url    = "${var.rancher_api_url}"
  access_key = "${var.rancher_access_key}"
  secret_key = "${var.rancher_secret_key}"
}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
}

data "triton_image" "image" {
  name    = "${var.triton_image_name}"
  version = "${var.triton_image_version}"
}

resource "rancher_registration_token" "token" {
  name           = "${var.hostname}"
  description    = "Registration token for ${var.hostname}"
  environment_id = "${var.rancher_environment_id}"

  host_labels = "${var.rancher_host_labels}"
}

data "template_file" "install_rancher_agent" {
  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.hostname}"
    rancher_agent_command     = "${rancher_registration_token.token.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"

    mount_path = "${var.triton_volume_mount_path}"
    nfs_path = "${element(coalescelist(triton_volume.host_volume.*.filesystem_path, list("")), 0)}"
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

resource "triton_volume" "host_volume" {
  count = "${var.triton_volume_mount_path != "" ? 1 : 0}"
  name = "${var.hostname}-Volume"
  networks = ["${data.triton_network.networks.*.id}"]
}