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

provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
  environment     = "${var.azure_environment}"
}

resource "azurerm_resource_group" "resource_group" {
  name     = "${var.name}-resource_group"
  location = "${var.azure_location}"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "${var.azure_virtual_network_name}"
  address_space       = ["${var.azure_virtual_network_address_space}"]
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

resource "azurerm_subnet" "subnet" {
  name                 = "${var.azure_subnet_name}"
  resource_group_name  = "${azurerm_resource_group.resource_group.name}"
  virtual_network_name = "${azurerm_virtual_network.vnet.name}"
  address_prefix       = "${var.azure_subnet_address_prefix}"
}

resource "azurerm_network_security_group" "firewall" {
  name                = "${var.azurerm_network_security_group_name}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

# Firewall requirements taken from:
# https://rancher.com/docs/rancher/v2.0/en/quick-start-guide/
resource "azurerm_network_security_rule" "rke_ports" {
  name      = "rke_ports"
  priority  = 1000
  direction = "Inbound"
  access    = "Allow"
  protocol  = "Tcp"

  source_port_ranges = [
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

  destination_port_range      = "*"
  source_address_prefix       = "VirtualNetwork"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}
