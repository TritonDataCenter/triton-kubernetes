provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
  environment     = "${var.azure_environment}"
}

locals {
  rancher_node_role = "${element(keys(var.rancher_host_labels), 0)}"
}

data "template_file" "install_rancher_agent" {
  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.hostname}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_api_url                    = "${var.rancher_api_url}"
    rancher_cluster_registration_token = "${var.rancher_cluster_registration_token}"
    rancher_cluster_ca_checksum        = "${var.rancher_cluster_ca_checksum}"
    rancher_node_role                  = "${local.rancher_node_role == "control" ? "controlplane" : local.rancher_node_role}"
    rancher_agent_image                = "${var.rancher_agent_image}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "azurerm_public_ip" "public_ip" {
  name                         = "${var.hostname}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${var.azure_resource_group_name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_network_interface" "nic" {
  name                = "${var.hostname}"
  location            = "${var.azure_location}"
  resource_group_name = "${var.azure_resource_group_name}"

  network_security_group_id = "${var.azure_network_security_group_id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${var.azure_subnet_id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.public_ip.id}"
  }
}

resource "azurerm_virtual_machine" "host" {
  name                  = "${var.hostname}"
  location              = "${var.azure_location}"
  resource_group_name   = "${var.azure_resource_group_name}"
  network_interface_ids = ["${azurerm_network_interface.nic.id}"]
  vm_size               = "${var.azure_size}"

  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.hostname}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "${var.hostname}-datadisk"
    managed_disk_type = "Standard_LRS"
    create_option     = "Empty"
    lun               = 0
    disk_size_gb      = "1023"
  }

  os_profile {
    computer_name  = "${var.hostname}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${data.template_file.install_rancher_agent.rendered}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}
