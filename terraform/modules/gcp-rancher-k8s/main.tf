data "external" "rancher_cluster" {
  program = ["bash", "${path.module}/files/rancher_cluster.sh"]

  query = {
    rancher_api_url       = "${var.rancher_api_url}"
    rancher_access_key    = "${var.rancher_access_key}"
    rancher_secret_key    = "${var.rancher_secret_key}"
    name                  = "${var.name}"
    k8s_version           = "${var.k8s_version}"
    k8s_network_provider  = "${var.k8s_network_provider}"
    k8s_registry          = "${var.k8s_registry}"
    k8s_registry_username = "${var.k8s_registry_username}"
    k8s_registry_password = "${var.k8s_registry_password}"
  }
}

provider "google" {
  credentials = "${file("${var.gcp_path_to_credentials}")}"
  project     = "${var.gcp_project_id}"
  region      = "${var.gcp_compute_region}"
}

resource "google_compute_network" "default" {
  name                    = "${var.name}"
  auto_create_subnetworks = "true"
}

# Firewall requirements taken from:
# https://rancher.com/docs/rancher/v2.0/en/quick-start-guide/
resource "google_compute_firewall" "rke_ports" {
  name        = "${var.name}-rke-ports"
  network     = "${google_compute_network.default.name}"
  source_tags = ["${var.name}-nodes"]

  allow {
    protocol = "tcp"

    ports = [
      "22",          # SSH
      "80",          # Canal
      "443",         # Canal
      "6443",        # Kubernetes API server
      "2379-2380",   # etcd server client API
      "10250",       # kubelet API
      "10251",       # scheduler
      "10252",       # controller
      "10256",       # kubeproxy
      "30000-32767", # NodePort Services
    ]
  }
}
