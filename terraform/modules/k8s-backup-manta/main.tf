locals {
  key_pair    = "${join(":", list(var.rancher_access_key, var.rancher_secret_key))}"
  auth_header = "${join(" ", list("Basic", base64encode(local.key_pair)))}"
  admin_token = "${base64encode(local.auth_header)}"
}

data "template_file" "kubeconfig" {
  template = "${file("${path.module}/files/kubeconfig.yaml")}"

  vars {
    kubernetes_host_ip = "${var.kubernetes_host_ip}"
    cluster_id         = "${var.cluster_id}"
    cluster_name       = "${var.cluster_name}"
    admin_name         = "${var.admin_name}"
    admin_token        = "${local.admin_token}"
  }
}

data "local_file" "triton_private_key" {
  filename = "${var.triton_key_path}"
}

data "template_file" "minio_manta_deployment" {
  template = "${file("${path.module}/files/minio-manta-deployment.yaml")}"

  vars {
    triton_account = "${var.triton_account}"
    triton_key_id = "${var.triton_key_id}"
    triton_key_path = "./triton_id_rsa"
    manta_subuser = "${var.manta_subuser}"
  }
}

data "template_file" "setup_ark_backup_manta" {
  template = "${file("${path.module}/files/setup_ark_backup_manta.sh.tpl")}"

  vars {
    kubeconfig_filedata = "${data.template_file.kubeconfig.rendered}"
    minio_manta_yaml_filedata = "${data.template_file.minio_manta_deployment.rendered}"
    triton_private_key = "${data.local_file.triton_private_key.content}"
    triton_key_path = "./triton_id_rsa"
  }
}

resource "null_resource" "setup_ark_backup_manta" {
  connection {
    type        = "ssh"
    user        = "${var.triton_ssh_user}"
    host        = "${var.triton_ssh_host}"
    private_key = "${file(var.triton_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
            ${data.template_file.setup_ark_backup_manta.rendered}
            EOF
  }
}
