output "rancher_environment_id" {
  value = "${rancher_environment.k8s.id}"
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
