output "rancher_cluster_id" {
  value = "${data.external.rancher_cluster.result.cluster_id}"
}

output "rancher_cluster_registration_token" {
  value = "${data.external.rancher_cluster.result.registration_token}"
}

output "rancher_cluster_ca_checksum" {
  value = "${data.external.rancher_cluster.result.ca_checksum}"
}

output "vsphere_datacenter_name" {
  value = "${var.vsphere_datacenter_name}"
}

output "vsphere_datastore_name" {
  value = "${var.vsphere_datastore_name}"
}

output "vsphere_resource_pool_name" {
  value = "${var.vsphere_resource_pool_name}"
}

output "vsphere_network_name" {
  value = "${var.vsphere_network_name}"
}
