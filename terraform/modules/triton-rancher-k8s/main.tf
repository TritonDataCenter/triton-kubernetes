provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

provider "rancher" {
  api_url    = "${var.rancher_api_url}"
  access_key = "2594FD14620DC702102A"
  secret_key = "QcUNzwK1HWLHk6W52hFWg3hD6tQXxY7NaMc3XGLz"
}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
}

data "triton_image" "image" {
  name    = "${var.triton_image_name}"
  version = "${var.triton_image_version}"
}

data "external" "rancher_environment_template" {
  program = ["bash", "${path.module}/files/rancher_environment_template.sh"]

  query = {
    rancher_api_url     = "${var.rancher_api_url}"
    rancher_access_key  = "2594FD14620DC702102A"
    rancher_secret_key  = "QcUNzwK1HWLHk6W52hFWg3hD6tQXxY7NaMc3XGLz"
    name                = "${var.name}-kubernetes"
    k8s_plane_isolation = "${var.k8s_plane_isolation}"
    k8s_registry        = "${var.k8s_registry}"
  }
}

resource "rancher_environment" "k8s" {
  name                = "${var.name}"
  project_template_id = "${data.external.rancher_environment_template.result.id}"

  member {
    external_id      = "1a1"
    external_id_type = "rancher_id"
    role             = "owner"
  }
}

resource "rancher_registry" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name           = "${var.rancher_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.rancher_registry}"
}

resource "rancher_registry_credential" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name         = "${var.rancher_registry}"
  registry_id  = "${rancher_registry.rancher_registry.id}"
  email        = "${var.rancher_registry_username}"
  public_value = "${var.rancher_registry_username}"
  secret_value = "${var.rancher_registry_password}"
}

resource "rancher_registry" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name           = "${var.k8s_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.k8s_registry}"
}

resource "rancher_registry_credential" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name         = "${var.k8s_registry}"
  registry_id  = "${rancher_registry.k8s_registry.id}"
  email        = "${var.k8s_registry_username}"
  public_value = "${var.k8s_registry_username}"
  secret_value = "${var.k8s_registry_password}"
}

resource "rancher_registration_token" "etcd" {
  count = "${var.etcd_node_count}"

  name           = "${var.name}-etcd-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-etcd-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.etcd.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "triton_machine" "etcd_host" {
  count = "${var.etcd_node_count}"

  package = "${var.etcd_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-etcd-${count.index + 1}"

  user_script = "${element(data.template_file.install_rancher_agent_etcd.*.rendered, count.index)}"

  networks = ["${data.triton_network.networks.*.id}"]

  cns = {
    services = ["etcd.${var.name}"]
  }

  affinity = ["role!=~etcd"]

  tags = {
    role = "etcd"
  }
}

resource "rancher_registration_token" "orchestration" {
  count = "${var.orchestration_node_count}"

  name           = "${var.name}-orchestration-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-orchestration-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.orchestration.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "triton_machine" "orchestration_host" {
  count = "${var.orchestration_node_count}"

  package = "${var.orchestration_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-orchestration-${count.index + 1}"

  user_script = "${element(data.template_file.install_rancher_agent_orchestration.*.rendered, count.index)}"

  networks = ["${data.triton_network.networks.*.id}"]

  cns = {
    services = ["orchestration.${var.name}"]
  }

  affinity = ["role!=~orchestration"]

  tags = {
    role = "orchestration"
  }
}

resource "rancher_registration_token" "compute" {
  count = "${var.compute_node_count}"

  name           = "${var.name}-compute-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-compute-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.compute.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
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
