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
resource "azurerm_network_security_rule" "rancher_ports" {
  name      = "rancher_ports"
  priority  = 1000
  direction = "Inbound"
  access    = "Allow"
  protocol  = "Tcp"

  source_port_ranges = [
    "22",  # SSH
    "80",  # Rancher UI
    "443", # Rancher UI
  ]

  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

resource "azurerm_public_ip" "public_ip" {
  name                         = "${var.name}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${azurerm_resource_group.resource_group.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_network_interface" "nic" {
  name                = "${var.name}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  network_security_group_id = "${azurerm_network_security_group.firewall.id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurerm_public_ip.public_ip.id}"
  }
}

resource "azurerm_virtual_machine" "host" {
  name                  = "${var.name}"
  location              = "${var.azure_location}"
  resource_group_name   = "${azurerm_resource_group.resource_group.name}"
  network_interface_ids = ["${azurerm_network_interface.nic.id}"]
  vm_size               = "${var.azure_size}"

  delete_os_disk_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.name}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "${var.name}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${data.template_file.install_docker.rendered}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}

data "azurerm_public_ip" "public_ip" {
  depends_on = ["azurerm_public_ip.public_ip"]

  name                = "${azurerm_public_ip.public_ip.name}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

locals {
  rancher_master_id = "${azurerm_virtual_machine.host.id}"
  rancher_master_ip = "${data.azurerm_public_ip.public_ip.ip_address}"
  ssh_user          = "${var.azure_ssh_user}"
  key_path          = "${var.azure_private_key_path}"
}

data "template_file" "install_docker" {
  template = "${file("${path.module}/files/install_docker_rancher.sh.tpl")}"

  vars {
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_server_image      = "${var.rancher_server_image}"
    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

data "template_file" "install_rancher_master" {
  template = "${file("${path.module}/files/install_rancher_master.sh.tpl")}"

  vars {
    rancher_server_image      = "${var.rancher_server_image}"
    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "null_resource" "install_rancher_master" {
  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_id = "${local.rancher_master_id}"
  }

  connection {
    type        = "ssh"
    user        = "${local.ssh_user}"
    host        = "${local.rancher_master_ip}"
    private_key = "${file(local.key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.install_rancher_master.rendered}
      EOF
  }
}

data "template_file" "setup_rancher_k8s" {
  template = "${file("${path.module}/files/setup_rancher.sh.tpl")}"

  vars {
    name                  = "${var.name}"
    rancher_host          = "https://127.0.0.1"
    host_registration_url = "https://${local.rancher_master_ip}"

    rancher_admin_password = "${var.rancher_admin_password}"
  }
}

resource "null_resource" "setup_rancher_k8s" {
  depends_on = ["null_resource.install_rancher_master"]

  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_id = "${local.rancher_master_id}"
  }

  connection {
    type        = "ssh"
    user        = "${local.ssh_user}"
    host        = "${local.rancher_master_ip}"
    private_key = "${file(local.key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.setup_rancher_k8s.rendered}
      EOF
  }
}

// The setup_rancher_k8s script will have stored a file with an api key
// We need to retrieve the contents of that file and output it.
// This is a hack to get around the Terraform Rancher provider not having resources for api keys.
module "rancher_access_key" {
  source  = "matti/outputs/shell"
  version = "0.0.1"

  // We ssh into the remote box and cat the file.
  // We echo the output from null_resource.setup_rancher_k8s to setup an implicit dependency.
  command = "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i ${local.key_path} ${local.ssh_user}@${local.rancher_master_ip} 'echo ${null_resource.setup_rancher_k8s.id} > /dev/null; cat ~/rancher_api_key | jq -r .name'"
}

module "rancher_secret_key" {
  source  = "matti/outputs/shell"
  version = "0.0.1"

  // We ssh into the remote box and cat the file.
  // We echo the output from null_resource.setup_rancher_k8s to setup an implicit dependency.
  command = "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i ${local.key_path} ${local.ssh_user}@${local.rancher_master_ip} 'echo ${null_resource.setup_rancher_k8s.id} > /dev/null; cat ~/rancher_api_key | jq -r .token | cut -d: -f2'"
}
