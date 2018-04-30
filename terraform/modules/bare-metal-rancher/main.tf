locals {
  rancher_master_id = "${var.host}"
  rancher_master_ip = "${var.host}"
  ssh_user          = "${var.ssh_user}"
  key_path          = "${var.key_path}"
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

resource "null_resource" "install_docker" {
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
      ${data.template_file.install_docker.rendered}
      EOF
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
  depends_on = ["null_resource.install_docker"]

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
