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
