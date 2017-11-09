provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
}

resource "azurerm_resource_group" "resource_group" {
  name     = "${var.azure_resource_group_name}"
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

provider "rancher" {
  api_url    = "${var.api_url}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

resource "rancher_environment" "k8s" {
  name          = "${var.name}"
  orchestration = "kubernetes"
}

resource "rancher_registration_token" "etcd" {
  name           = "etcd_host_tokens"
  description    = "Registration token for ${var.name} etcd hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    etcd = "true"
  }
}

data "template_file" "install_rancher_agent_etcd" {
  count = "${var.etcd_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-etcd-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.etcd.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "azurerm_network_interface" "nic_etcd" {
  count = "${var.etcd_node_count}"

  name                = "${var.name}-etcd-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
  }
}

resource "azurerm_virtual_machine" "etcd" {
  count = "${var.etcd_node_count}"

  name                  = "${var.name}-etcd-${count.index + 1}"
  location              = "${var.azure_location}"
  resource_group_name   = "${azurerm_resource_group.resource_group.name}"
  network_interface_ids = ["${element(azurerm_network_interface.nic_etcd.*.id, count.index)}"]
  vm_size               = "${var.etcd_azure_size}"

  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.name}-etcd-${count.index + 1}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "${var.name}-etcd-${count.index + 1}-datadisk"
    managed_disk_type = "Standard_LRS"
    create_option     = "Empty"
    lun               = 0
    disk_size_gb      = "1023"
  }

  os_profile {
    computer_name  = "${var.name}-etcd-${count.index + 1}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${element(data.template_file.install_rancher_agent_etcd.*.rendered, count.index)}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}

resource "rancher_registration_token" "orchestration" {
  name           = "orchestration_host_tokens"
  description    = "Registration token for ${var.name} orchestration hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    orchestration = "true"
  }
}

data "template_file" "install_rancher_agent_orchestration" {
  count = "${var.orchestration_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-orchestration-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.orchestration.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "azurerm_network_interface" "nic_orchestration" {
  count = "${var.orchestration_node_count}"

  name                = "${var.name}-orchestration-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
  }
}

resource "azurerm_virtual_machine" "orchestration" {
  count = "${var.orchestration_node_count}"

  name                  = "${var.name}-orchestration-${count.index + 1}"
  location              = "${var.azure_location}"
  resource_group_name   = "${azurerm_resource_group.resource_group.name}"
  network_interface_ids = ["${element(azurerm_network_interface.nic_orchestration.*.id, count.index)}"]
  vm_size               = "${var.orchestration_azure_size}"

  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.name}-orchestration-${count.index + 1}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "${var.name}-orchestration-${count.index + 1}-datadisk"
    managed_disk_type = "Standard_LRS"
    create_option     = "Empty"
    lun               = 0
    disk_size_gb      = "1023"
  }

  os_profile {
    computer_name  = "${var.name}-orchestration-${count.index + 1}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${element(data.template_file.install_rancher_agent_orchestration.*.rendered, count.index)}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}

resource "rancher_registration_token" "compute" {
  name           = "compute_host_tokens"
  description    = "Registration token for ${var.name} compute hosts"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    compute = "true"
  }
}

data "template_file" "install_rancher_agent_compute" {
  count = "${var.compute_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-compute-${count.index + 1}"
    rancher_agent_command     = "${rancher_registration_token.compute.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"
  }
}

resource "azurerm_network_interface" "nic_compute" {
  count = "${var.compute_node_count}"

  name                = "${var.name}-compute-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
  }
}

resource "azurerm_virtual_machine" "compute" {
  count = "${var.compute_node_count}"

  name                  = "${var.name}-compute-${count.index + 1}"
  location              = "${var.azure_location}"
  resource_group_name   = "${azurerm_resource_group.resource_group.name}"
  network_interface_ids = ["${element(azurerm_network_interface.nic_compute.*.id, count.index)}"]
  vm_size               = "${var.compute_azure_size}"

  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.name}-compute-${count.index + 1}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "${var.name}-compute-${count.index + 1}-datadisk"
    managed_disk_type = "Standard_LRS"
    create_option     = "Empty"
    lun               = 0
    disk_size_gb      = "1023"
  }

  os_profile {
    computer_name  = "${var.name}-compute-${count.index + 1}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${element(data.template_file.install_rancher_agent_compute.*.rendered, count.index)}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}
