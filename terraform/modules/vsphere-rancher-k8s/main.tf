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

provider "vsphere" {
  user           = "${var.vsphere_user}"
  password       = "${var.vsphere_password}"
  vsphere_server = "${var.vsphere_server}"

  allow_unverified_ssl = "true"
}

data "vsphere_datacenter" "dc" {
  name = "${var.vsphere_datacenter_name}"
}

data "vsphere_datastore" "datastore" {
  name          = "${var.vsphere_datastore_name}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_resource_pool" "pool" {
  name          = "${var.vsphere_resource_pool_name}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_network" "network" {
  name          = "${var.vsphere_network_name}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}
