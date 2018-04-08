provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
  environment     = "${var.azure_environment}"
}

provider "rancher" {
  api_url    = "${var.rancher_api_url}"
  access_key = "${var.rancher_access_key}"
  secret_key = "${var.rancher_secret_key}"
}

resource "rancher_registration_token" "token" {
  name           = "${var.hostname}"
  description    = "Registration token for ${var.hostname}"
  environment_id = "${var.rancher_environment_id}"

  host_labels = "${var.rancher_host_labels}"
}

data "template_file" "install_rancher_agent" {
  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.hostname}"
    rancher_agent_command     = "${rancher_registration_token.token.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"

    disk_mount_path = "${var.azure_disk_mount_path}"
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

resource "azurerm_managed_disk" "host_disk" {
  count = "${var.azure_disk_mount_path == "" ? 0 : 1}"

  name                 = "${var.hostname}-managed-disk"
  location             = "${var.azure_location}"
  resource_group_name  = "${var.azure_resource_group_name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "${var.azure_disk_size}"
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
    name              = "${element(coalescelist(azurerm_managed_disk.host_disk.*.name, list("${var.hostname}-datadisk")), 0)}"
    managed_disk_id   = "${element(coalescelist(azurerm_managed_disk.host_disk.*.id, list("")), 0)}"
    managed_disk_type = "Standard_LRS"
    create_option     = "${var.azure_disk_mount_path == "" ? "Empty" : "Attach"}"
    lun               = 0
    disk_size_gb      = "${element(coalescelist(azurerm_managed_disk.host_disk.*.disk_size_gb, list("1023")), 0)}"
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
