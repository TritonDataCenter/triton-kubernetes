output "rancher_environment_id" {
  value = "${rancher_environment.k8s.id}"
}

output "gcp_compute_network_name" {
  value = "${google_compute_network.default.name}"
}
