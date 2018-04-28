provider "google" {
  credentials = "${file("${var.gcp_path_to_credentials}")}"
  project     = "${var.gcp_project_id}"
  region      = "${var.gcp_compute_region}"
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

  # There's now way to specify for 0 attached_disk blocks
  # We need to wait for Terraform for_each support https://github.com/hashicorp/terraform/issues/7034
  # This way we'll only add an attached_disk block when there are disks to attach
  # attached_disk {
  #   source = "${element(concat(google_compute_disk.host_volume.*.self_link, list("")), 0)}"
  # }

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
