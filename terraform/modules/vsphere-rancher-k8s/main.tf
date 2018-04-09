provider "vsphere" {
  user           = "${var.vsphere_user}"
  password       = "${var.vsphere_password}"
  vsphere_server = "${var.vsphere_server}"

  allow_unverified_ssl = ${var.allow_unverified_ssl}
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
    rancher_access_key  = "${var.rancher_access_key}"
    rancher_secret_key  = "${var.rancher_secret_key}"
    name                = "${var.name}-kubernetes"
    k8s_plane_isolation = "${var.k8s_plane_isolation}"
    k8s_registry        = "${var.k8s_registry}"
  }
}

resource "rancher_environment" "k8s" {
  name                = "${var.name}"
  project_template_id = "${data.external.rancher_environment_template.result.id}"

  member {
    external_id      = "1a1"
    external_id_type = "rancher_id"
    role             = "owner"
  }
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