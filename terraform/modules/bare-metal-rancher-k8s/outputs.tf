output "rancher_cluster_id" {
  value = "${data.external.rancher_cluster.result.cluster_id}"
}

output "rancher_cluster_registration_token" {
  value = "${data.external.rancher_cluster.result.registration_token}"
}

output "rancher_cluster_ca_checksum" {
  value = "${data.external.rancher_cluster.result.ca_checksum}"
}
