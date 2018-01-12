provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
  environment     = "${var.azure_environment}"
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

resource "azurerm_network_security_group" "firewall" {
  name                = "${var.azurerm_network_security_group_name}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

resource "azurerm_network_security_rule" "port500" {
  name                        = "Port500-UdpAllowAny"
  priority                    = 1000
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

resource "azurerm_network_security_rule" "port4500" {
  name                        = "Port4500-UdpAllowAny"
  priority                    = 1001
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

provider "rancher" {
  api_url    = "${var.rancher_api_url}"
  access_key = "${var.rancher_access_key}"
  secret_key = "${var.rancher_secret_key}"
}

data "external" "rancher_environment_template" {
  program = ["bash", "${path.module}/files/rancher_environment_template.sh"]

  query = {
    rancher_api_url     = "${var.rancher_api_url}"
    name                = "${var.name}-kubernetes"
    k8s_plane_isolation = "${var.k8s_plane_isolation}"
    k8s_registry        = "${var.k8s_registry}"
  }
}

resource "rancher_environment" "k8s" {
  name                = "${var.name}"
  project_template_id = "${data.external.rancher_environment_template.result.id}"
}

resource "rancher_registry" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name           = "${var.rancher_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.rancher_registry}"
}

resource "rancher_registry_credential" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name         = "${var.rancher_registry}"
  registry_id  = "${rancher_registry.rancher_registry.id}"
  email        = "${var.rancher_registry_username}"
  public_value = "${var.rancher_registry_username}"
  secret_value = "${var.rancher_registry_password}"
}

resource "rancher_registry" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name           = "${var.k8s_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.k8s_registry}"
}

resource "rancher_registry_credential" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name         = "${var.k8s_registry}"
  registry_id  = "${rancher_registry.k8s_registry.id}"
  email        = "${var.k8s_registry_username}"
  public_value = "${var.k8s_registry_username}"
  secret_value = "${var.k8s_registry_password}"
}

resource "rancher_registration_token" "etcd" {
  count = "${var.etcd_node_count}"

  name           = "${var.name}-etcd-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-etcd-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.etcd.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "azurerm_public_ip" "public_ip_etcd" {
  count = "${var.etcd_node_count}"

  name                         = "${var.name}-etcd-${count.index + 1}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${azurerm_resource_group.resource_group.name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_network_interface" "nic_etcd" {
  count = "${var.etcd_node_count}"

  name                = "${var.name}-etcd-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  network_security_group_id = "${azurerm_network_security_group.firewall.id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.public_ip_etcd.*.id, count.index)}"
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
  count = "${var.orchestration_node_count}"

  name           = "${var.name}-orchestration-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-orchestration-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.orchestration.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "azurerm_public_ip" "public_ip_orchestration" {
  count = "${var.orchestration_node_count}"

  name                         = "${var.name}-orchestration-${count.index + 1}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${azurerm_resource_group.resource_group.name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_network_interface" "nic_orchestration" {
  count = "${var.orchestration_node_count}"

  name                = "${var.name}-orchestration-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  network_security_group_id = "${azurerm_network_security_group.firewall.id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.public_ip_orchestration.*.id, count.index)}"
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
  count = "${var.compute_node_count}"

  name           = "${var.name}-compute-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-compute-${count.index + 1} host"
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
    rancher_agent_command     = "${element(rancher_registration_token.compute.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "azurerm_public_ip" "public_ip_compute" {
  count = "${var.compute_node_count}"

  name                         = "${var.name}-compute-${count.index + 1}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${azurerm_resource_group.resource_group.name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_network_interface" "nic_compute" {
  count = "${var.compute_node_count}"

  name                = "${var.name}-compute-${count.index + 1}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  network_security_group_id = "${azurerm_network_security_group.firewall.id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.public_ip_compute.*.id, count.index)}"
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
