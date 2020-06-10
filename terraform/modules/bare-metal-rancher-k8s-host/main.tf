locals {
  rancher_node_role = element(keys(var.rancher_host_labels), 0)
}

data "template_file" "install_rancher_agent" {
  template = file("${path.module}/files/install_rancher_agent.sh.tpl")

  vars = {
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

resource "null_resource" "install_rancher_agent" {
  triggers = {
    host = var.host
  }

  connection {
    type         = "ssh"
    user         = var.ssh_user
    bastion_host = var.bastion_host
    host         = var.host
    private_key  = file(var.key_path)
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.install_rancher_agent.rendered}
      
EOF

  }
}

