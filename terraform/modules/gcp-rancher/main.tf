provider "google" {
  credentials = file(var.gcp_path_to_credentials)
  project     = var.gcp_project_id
  region      = var.gcp_compute_region
}

resource "google_compute_network" "default" {
  name                    = var.name
  auto_create_subnetworks = "true"
}

# Firewall requirements taken from:
# https://rancher.com/docs/rancher/v2.0/en/quick-start-guide/
resource "google_compute_firewall" "rancher_master_ports" {
  name          = "${var.name}-rancher-master-ports"
  network       = google_compute_network.default.name
  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "tcp"

    ports = [
      "22",  # SSH
      "80",  # Rancher UI
      "443", # Rancher UI
    ]
  }
}

resource "google_compute_instance" "rancher_master" {
  name         = var.name
  machine_type = var.gcp_machine_type
  zone         = var.gcp_instance_zone
  project      = var.gcp_project_id

  boot_disk {
    initialize_params {
      image = var.gcp_image
    }
  }

  network_interface {
    network = google_compute_network.default.name

    access_config {
      // Ephemeral IP
    }
  }

  metadata = {
    sshKeys = "${var.gcp_ssh_user}:${file(var.gcp_public_key_path)}"
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = data.template_file.install_docker.rendered
}

locals {
  rancher_master_id = google_compute_instance.rancher_master.instance_id
  rancher_master_ip = google_compute_instance.rancher_master.network_interface[0].access_config[0].assigned_nat_ip
  ssh_user          = var.gcp_ssh_user
  key_path          = var.gcp_private_key_path
}

data "template_file" "install_docker" {
  template = file("${path.module}/files/install_docker_rancher.sh.tpl")

  vars = {
    docker_engine_install_url = var.docker_engine_install_url
    rancher_server_image      = var.rancher_server_image
    rancher_registry          = var.rancher_registry
    rancher_registry_username = var.rancher_registry_username
    rancher_registry_password = var.rancher_registry_password
  }
}

data "template_file" "install_rancher_master" {
  template = file("${path.module}/files/install_rancher_master.sh.tpl")

  vars = {
    rancher_server_image      = var.rancher_server_image
    rancher_registry          = var.rancher_registry
    rancher_registry_username = var.rancher_registry_username
    rancher_registry_password = var.rancher_registry_password
  }
}

resource "null_resource" "install_rancher_master" {
  # Changes to any instance of the cluster requires re-provisioning
  triggers = {
    rancher_master_id = local.rancher_master_id
  }

  connection {
    type        = "ssh"
    user        = local.ssh_user
    host        = local.rancher_master_ip
    private_key = file(local.key_path)
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.install_rancher_master.rendered}
      
EOF

  }
}

data "template_file" "setup_rancher_k8s" {
  template = file("${path.module}/files/setup_rancher.sh.tpl")

  vars = {
    name                   = var.name
    rancher_host           = "https://127.0.0.1"
    host_registration_url  = "https://${local.rancher_master_ip}"
    rancher_admin_password = var.rancher_admin_password
  }
}

resource "null_resource" "setup_rancher_k8s" {
  depends_on = [null_resource.install_rancher_master]

  # Changes to any instance of the cluster requires re-provisioning
  triggers = {
    rancher_master_id = local.rancher_master_id
  }

  connection {
    type        = "ssh"
    user        = local.ssh_user
    host        = local.rancher_master_ip
    private_key = file(local.key_path)
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.setup_rancher_k8s.rendered}
      
EOF

  }
}

// The setup_rancher_k8s script will have stored a file with an api key
// We need to retrieve the contents of that file and output it.
// This is a hack to get around the Terraform Rancher provider not having resources for api keys.
data "external" "rancher_server" {
  program = ["bash", "${path.module}/files/rancher_server.sh"]

  query = {
    id        = null_resource.setup_rancher_k8s.id // used to create an implicit dependency
    ssh_host  = local.rancher_master_ip
    ssh_user  = local.ssh_user
    key_path  = local.key_path
    file_path = "~/rancher_api_key"
  }
}

