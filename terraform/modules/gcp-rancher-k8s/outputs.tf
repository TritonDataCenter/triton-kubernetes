output "rancher_cluster_id" {
  value = "${data.external.rancher_cluster.result.cluster_id}"
}

output "rancher_cluster_registration_token" {
  value = "${data.external.rancher_cluster.result.registration_token}"
}

output "rancher_cluster_ca_checksum" {
  value = "${data.external.rancher_cluster.result.ca_checksum}"
}

output "gcp_compute_network_name" {
  value = "${google_compute_network.default.name}"
}

output "gcp_compute_firewall_host_tag" {
  value = "${var.name}-nodes"
}
