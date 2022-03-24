locals {
  rancher_master_id = var.host
  rancher_master_ip = var.host
  ssh_user          = var.ssh_user
  key_path          = var.key_path
}

data "template_file" "install_docker" {
  template = file("${path.module}/files/install_docker_rancher.sh.tpl")

  vars = {
    docker_engine_install_url = var.docker_engine_install_url
    rancher_server_image      = var.rancher_server_image
    rancher_registry          = var.rancher_registry
    rancher_registry_username = var.rancher_registry_username
    rancher_registry_password = var.rancher_registry_password
  }
}

resource "null_resource" "install_docker" {
  # Changes to any instance of the cluster requires re-provisioning
  triggers = {
    rancher_master_id = local.rancher_master_id
  }

  connection {
    type        = "ssh"
    user        = local.ssh_user
    host        = local.rancher_master_ip
    private_key = file(local.key_path)
  }

  provisioner "remote-exec" {
    inline = [
      "${data.template_file.install_docker.rendered}"
    ]
  }
}

data "template_file" "install_rancher_master" {
  template = file("${path.module}/files/install_rancher_master.sh.tpl")

  vars = {
    rancher_server_image      = var.rancher_server_image
    rancher_registry          = var.rancher_registry
    rancher_registry_username = var.rancher_registry_username
    rancher_registry_password = var.rancher_registry_password
  }
}

resource "null_resource" "install_rancher_master" {
  depends_on = [null_resource.install_docker]

  # Changes to any instance of the cluster requires re-provisioning
  triggers = {
    rancher_master_id = local.rancher_master_id
  }

  connection {
    type        = "ssh"
    user        = local.ssh_user
    host        = local.rancher_master_ip
    private_key = file(local.key_path)
  }

  provisioner "remote-exec" {
    inline = [
      "${data.template_file.install_rancher_master.rendered}"
    ]
  }
}

data "template_file" "setup_rancher_k8s" {
  template = file("${path.module}/files/setup_rancher.sh.tpl")

  vars = {
    name                   = var.name
    rancher_host           = "https://127.0.0.1"
    host_registration_url  = "https://${local.rancher_master_ip}"
    rancher_admin_password = var.rancher_admin_password
  }
}

resource "null_resource" "setup_rancher_k8s" {
  depends_on = [null_resource.install_rancher_master]

  # Changes to any instance of the cluster requires re-provisioning
  triggers = {
    rancher_master_id = local.rancher_master_id
  }

  connection {
    type        = "ssh"
    user        = local.ssh_user
    host        = local.rancher_master_ip
    private_key = file(local.key_path)
  }

  provisioner "remote-exec" {
    inline = [
      "${data.template_file.setup_rancher_k8s.rendered}"
      ]

  }
}

// The setup_rancher_k8s script will have stored a file with an api key
// We need to retrieve the contents of that file and output it.
// This is a hack to get around the Terraform Rancher provider not having resources for api keys.
data "external" "rancher_server" {
  program = ["bash", "${path.module}/files/rancher_server.sh"]

  query = {
    id        = null_resource.setup_rancher_k8s.id // used to create an implicit dependency
    ssh_host  = local.rancher_master_ip
    ssh_user  = local.ssh_user
    key_path  = local.key_path
    file_path = "~/rancher_api_key"
  }
}

