data "external" "rancher_cluster" {
  program = ["bash", "${path.module}/files/rancher_cluster_import.sh"]

  query = {
    rancher_api_url    = var.rancher_api_url
    rancher_access_key = var.rancher_access_key
    rancher_secret_key = var.rancher_secret_key
    name               = var.name
  }
}

provider "google" {
  credentials = file(var.gcp_path_to_credentials)
  project     = var.gcp_project_id
  region      = var.gcp_compute_region
}

resource "google_container_cluster" "primary" {
  name               = var.name
  zone               = var.gcp_zone
  initial_node_count = var.node_count

  min_master_version = var.k8s_version
  node_version       = var.k8s_version

  additional_zones = var.gcp_additional_zones

  master_auth {
    username = "admin"
    password = var.password
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }
}

locals {
  kube_config_path = "./${var.name}_config"
}

# Bootstrap rancher in gke environment
resource "null_resource" "import_rancher" {
  triggers = {
    cluster = google_container_cluster.primary.endpoint
  }

  provisioner "local-exec" {
    command = "gcloud auth activate-service-account --key-file ${var.gcp_path_to_credentials}"
  }

  provisioner "local-exec" {
    command = "gcloud container clusters get-credentials ${var.name} --zone ${var.gcp_zone} --project ${var.gcp_project_id}"

    environment = {
      KUBECONFIG = local.kube_config_path
    }
  }

  provisioner "local-exec" {
    command = "curl --insecure -sfL ${var.rancher_api_url}/v3/import/${data.external.rancher_cluster.result.registration_token}.yaml | kubectl apply -f -"

    environment = {
      KUBECONFIG = local.kube_config_path
    }
  }

  provisioner "local-exec" {
    command = "rm ${local.kube_config_path}"
  }

  provisioner "local-exec" {
    command = "gcloud auth revoke"
  }
}

