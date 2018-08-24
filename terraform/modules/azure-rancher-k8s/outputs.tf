output "rancher_cluster_id" {
  value = "${lookup(data.external.rancher_cluster.result, "cluster_id")}"
}

output "rancher_cluster_registration_token" {
  value = "${lookup(data.external.rancher_cluster.result, "registration_token")}"
}

output "rancher_cluster_ca_checksum" {
  value = "${lookup(data.external.rancher_cluster.result, "ca_checksum")}"
}

output "azure_resource_group_name" {
  value = "${azurerm_resource_group.resource_group.name}"
}

output "azure_network_security_group_id" {
  value = "${azurerm_network_security_group.firewall.id}"
}

output "azure_subnet_id" {
  value = "${azurerm_subnet.subnet.id}"
}
