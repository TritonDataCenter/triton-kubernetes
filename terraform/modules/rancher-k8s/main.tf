provider "rancher" {
  api_url    = "${var.api_url}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

resource "rancher_environment" "k8s" {
  name          = "${var.name}"
  orchestration = "kubernetes"
}

data "template_file" "setup_rancher_k8s_triton" {
  template = "${file("${path.module}/files/setup_rancher_k8s_triton.sh.tpl")}"

  vars {
    name           = "${var.name}"
    rancher_host   = "${var.api_url}"
    environment_id = "${rancher_environment.k8s.id}"

    etcd_node_count          = "${var.etcd_node_count}"
    orchestration_node_count = "${var.orchestration_node_count}"
    compute_node_count       = "${var.compute_node_count}"

    docker_engine_install_url = "${var.docker_engine_install_url}"

    triton_account                       = "${var.triton_account}"
    triton_key_path                      = "${var.triton_key_path}"
    triton_key_id                        = "${var.triton_key_id}"
    triton_url                           = "${var.triton_url}"
    triton_image_name                    = "${var.triton_image_name}"
    triton_image_version                 = "${var.triton_image_version}"
    triton_ssh_user                      = "${var.triton_ssh_user}"
    etcd_triton_machine_package          = "${var.etcd_triton_machine_package}"
    orchestration_triton_machine_package = "${var.orchestration_triton_machine_package}"
    compute_triton_machine_package       = "${var.compute_triton_machine_package}"
  }
}

resource "null_resource" "setup_rancher_k8s" {
  # Changes to the environment require reprovisioning.
  triggers {
    rancher_environment_id = "${rancher_environment.k8s.id}"
  }

  provisioner "local-exec" {
    command = <<EOF
      ${data.template_file.setup_rancher_k8s_triton.rendered}
      EOF
  }
}
