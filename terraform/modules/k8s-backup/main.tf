locals {
    key_pair = "${join(":", list(var.rancher_access_key, var.rancher_secret_key))}"
    auth_header = "${join(" ", list("Basic", base64encode(local.key_pair)))}"
    admin_token = "${base64encode(local.auth_header)}"
}

data "template_file" "kubeconfig" {
    template = "${file("${path.module}/files/kubeconfig.yaml")}"

    vars {
        kubernetes_host_ip = "${var.kubernetes_host_ip}"
        cluster_id = "${var.cluster_id}"
        cluster_name = "${var.cluster_name}"
        admin_name = "${var.admin_name}"
        admin_token = "${local.admin_token}"
    }
}

data "template_file" "setup_ark_backup" {
    template = "${file("${path.module}/files/setup_ark_backup.sh.tpl")}"

    vars {
        kubeconfig_filedata = "${data.template_file.kubeconfig.rendered}"
    }
}

resource "null_resource" "setup_ark_backup" {
    connection {
        type = "ssh"
        user = "${var.triton_ssh_user}"
        host = "${var.triton_ssh_host}"
        private_key = "${file(var.triton_key_path)}"
    }

    provisioner "remote-exec" {
        inline = <<EOF
            ${data.template_file.setup_ark_backup.rendered}
            EOF
    }
}
