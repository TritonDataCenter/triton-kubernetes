provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

provider "rancher" {
  api_url    = "${var.api_url}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
}

data "triton_image" "image" {
  name    = "${var.triton_image_name}"
  version = "${var.triton_image_version}"
}

resource "rancher_environment" "k8s" {
  name          = "${var.name}"
  orchestration = "kubernetes"
}

resource "rancher_registration_token" "etcd" {
  name           = "etcd_host_tokens"
  description    = "Registration token for ${var.name} etcd hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    etcd = "true"
  }
}

data "template_file" "install_rancher_agent_etcd" {
  count = "${var.etcd_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-etcd-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.etcd.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "triton_machine" "etcd_host" {
  count = "${var.etcd_node_count}"

  package = "${var.etcd_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-etcd-${count.index + 1}"

  user_script = "${element(data.template_file.install_rancher_agent_etcd.*.rendered, count.index)}"

  networks = ["${data.triton_network.networks.*.id}"]
}

resource "rancher_registration_token" "orchestration" {
  name           = "orchestration_hosts_token"
  description    = "Registration token for ${var.name} orchestration hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    orchestration = "true"
  }
}

data "template_file" "install_rancher_agent_orchestration" {
  count = "${var.orchestration_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-orchestration-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.orchestration.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "triton_machine" "orchestration_host" {
  count = "${var.orchestration_node_count}"

  package = "${var.orchestration_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-orchestration-${count.index + 1}"

  user_script = "${element(data.template_file.install_rancher_agent_orchestration.*.rendered, count.index)}"

  networks = ["${data.triton_network.networks.*.id}"]
}

resource "rancher_registration_token" "compute" {
  name           = "compute_hosts_token"
  description    = "Registration token for ${var.name} compute hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    compute = "true"
  }
}

data "template_file" "install_rancher_agent_compute" {
  count = "${var.compute_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-compute-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.compute.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "triton_machine" "compute_host" {
  count = "${var.compute_node_count}"

  package = "${var.compute_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-compute-${count.index + 1}"

  user_script = "${element(data.template_file.install_rancher_agent_compute.*.rendered, count.index)}"

  networks = ["${data.triton_network.networks.*.id}"]
}
