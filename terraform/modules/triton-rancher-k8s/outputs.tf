output "rancher_cluster_id" {
  value = lookup(data.external.rancher_cluster.result, "cluster_id")
}

output "rancher_cluster_registration_token" {
  value = lookup(data.external.rancher_cluster.result, "registration_token")
}

output "rancher_cluster_ca_checksum" {
  value = lookup(data.external.rancher_cluster.result, "ca_checksum")
}
