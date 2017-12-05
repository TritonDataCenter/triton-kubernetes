provider "google" {
  credentials = "${file("${var.gcp_path_to_credentials}")}"
  project     = "${var.gcp_project_id}"
  region      = "${var.gcp_compute_region}"
}

resource "google_compute_firewall" "default" {
  name          = "${var.compute_firewall}"
  network       = "default"
  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "udp"
    ports    = ["500", "4500"]
  }
}

provider "rancher" {
  api_url = "${var.api_url}"
}

data "external" "rancher_environment_template" {
  program = ["bash", "${path.module}/files/rancher_environment_template.sh"]

  query = {
    rancher_api_url     = "${var.api_url}"
    name                = "${var.name}-kubernetes"
    k8s_plane_isolation = "${var.k8s_plane_isolation}"
    k8s_registry        = "${var.k8s_registry}"
  }
}

resource "rancher_environment" "k8s" {
  name                = "${var.name}"
  project_template_id = "${data.external.rancher_environment_template.result.id}"
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

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "google_compute_instance" "etcd" {
  count = "${var.etcd_node_count}"

  name         = "${var.name}-etcd-${count.index + 1}"
  machine_type = "${var.etcd_gcp_instance_type}"
  zone         = "${var.gcp_instance_zone}"

  boot_disk {
    initialize_params {
      image = "${var.gcp_image}"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = "${element(data.template_file.install_rancher_agent_etcd.*.rendered, count.index)}"
}

resource "rancher_registration_token" "orchestration" {
  name           = "orchestration_host_tokens"
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

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "google_compute_instance" "orchestration" {
  count = "${var.orchestration_node_count}"

  name         = "${var.name}-orchestration-${count.index + 1}"
  machine_type = "${var.orchestration_gcp_instance_type}"
  zone         = "${var.gcp_instance_zone}"

  boot_disk {
    initialize_params {
      image = "${var.gcp_image}"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = "${element(data.template_file.install_rancher_agent_orchestration.*.rendered, count.index)}"
}

resource "rancher_registration_token" "compute" {
  name           = "compute_host_tokens"
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

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "google_compute_instance" "compute" {
  count = "${var.compute_node_count}"

  name         = "${var.name}-compute-${count.index + 1}"
  machine_type = "${var.compute_gcp_instance_type}"
  zone         = "${var.gcp_instance_zone}"

  boot_disk {
    initialize_params {
      image = "${var.gcp_image}"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = "${element(data.template_file.install_rancher_agent_compute.*.rendered, count.index)}"
}
