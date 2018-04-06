provider "google" {
  credentials = "${file("${var.gcp_path_to_credentials}")}"
  project     = "${var.gcp_project_id}"
  region      = "${var.gcp_compute_region}"
}

provider "rancher" {
  api_url    = "${var.rancher_api_url}"
  access_key = "${var.rancher_access_key}"
  secret_key = "${var.rancher_secret_key}"
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

    disk_mount_path = "${var.gcp_disk_mount_path}"
  }
}

resource "google_compute_instance" "host" {
  name         = "${var.hostname}"
  machine_type = "${var.gcp_machine_type}"
  zone         = "${var.gcp_instance_zone}"
  project      = "${var.gcp_project_id}"

  boot_disk {
    initialize_params {
      image = "${var.gcp_image}"
    }
  }

  attached_disk {
    source = "${var.gcp_disk_type == "" ? "" : google_compute_disk.host_volume.self_link}"
  }

  network_interface {
    network = "${var.gcp_compute_network_name}"

    access_config {
      // Ephemeral IP
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = "${data.template_file.install_rancher_agent.rendered}"
}

resource "google_compute_disk" "host_volume" {
  count = "${var.gcp_disk_type == "" ? 0 : 1}"

  type = "${var.gcp_disk_type}"
  name = "${var.hostname}-volume"
  zone = "${var.gcp_instance_zone}"
  size = "${var.gcp_disk_size}"
}
