provider "vsphere" {
  user           = var.vsphere_user
  password       = var.vsphere_password
  vsphere_server = var.vsphere_server

  allow_unverified_ssl = true
}

locals {
  rancher_node_role = element(keys(var.rancher_host_labels), 0)
}

data "template_file" "install_rancher_agent" {
  template = file("${path.module}/files/install_rancher_agent.sh.tpl")

  vars = {
    hostname                           = var.hostname
    docker_engine_install_url          = var.docker_engine_install_url
    rancher_api_url                    = var.rancher_api_url
    rancher_cluster_registration_token = var.rancher_cluster_registration_token
    rancher_cluster_ca_checksum        = var.rancher_cluster_ca_checksum
    rancher_node_role                  = local.rancher_node_role == "control" ? "controlplane" : local.rancher_node_role
    rancher_agent_image                = var.rancher_agent_image
    rancher_registry                   = var.rancher_registry
    rancher_registry_username          = var.rancher_registry_username
    rancher_registry_password          = var.rancher_registry_password
  }
}

data "vsphere_datacenter" "dc" {
  name = var.vsphere_datacenter_name
}

data "vsphere_datastore" "datastore" {
  name          = var.vsphere_datastore_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_resource_pool" "pool" {
  name          = var.vsphere_resource_pool_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "network" {
  name          = var.vsphere_network_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_virtual_machine" "template" {
  name          = var.vsphere_template_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_virtual_machine" "vm" {
  name             = var.hostname
  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.datastore.id

  num_cpus = 2
  memory   = 2048
  guest_id = data.vsphere_virtual_machine.template.guest_id

  scsi_type = data.vsphere_virtual_machine.template.scsi_type

  network_interface {
    network_id   = data.vsphere_network.network.id
    adapter_type = data.vsphere_virtual_machine.template.network_interface_types[0]
  }

  disk {
    label            = "disk0"
    size             = data.vsphere_virtual_machine.template.disks[0].size
    eagerly_scrub    = data.vsphere_virtual_machine.template.disks[0].eagerly_scrub
    thin_provisioned = data.vsphere_virtual_machine.template.disks[0].thin_provisioned
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
  }
}

resource "null_resource" "install_rancher_agent" {
  triggers = {
    vsphere_virtual_machine_id = vsphere_virtual_machine.vm.id
  }

  connection {
    type = "ssh"
    user = var.ssh_user

    host        = vsphere_virtual_machine.vm.default_ip_address
    private_key = file(var.key_path)
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.install_rancher_agent.rendered}
      
EOF

  }
}

