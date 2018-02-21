provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

locals {
  rancher_url = "${var.ha ? format("http://%s-proxy.svc.%s.us-east-1.triton.zone", lower(var.name), data.triton_account.main.id) : format("http://%s:8080", element(triton_machine.rancher_master.*.primaryip, 0))}"
}

data "triton_account" "main" {}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
}

data "triton_network" "public" {
  name = "Joyent-SDC-Public"
}

data "triton_network" "private" {
  name = "Joyent-SDC-Private"
}

data "triton_image" "image" {
  name    = "${var.triton_image_name}"
  version = "${var.triton_image_version}"
}

data "triton_image" "mysql_image" {
  name    = "${var.triton_mysql_image_name}"
  version = "${var.triton_mysql_image_version}"
}

data "template_file" "install_rancher_mysqldb" {
  template = "${file("${path.module}/files/install_rancher_mysqldb.sh")}"

  vars {
    mysqldb_port          = "${var.mysqldb_port}"
    mysqldb_user          = "${var.mysqldb_username}"
    mysqldb_password      = "${var.mysqldb_password}"
    mysqldb_database_name = "${var.mysqldb_database_name}"
  }
}

data "template_file" "install_nginx" {
  template = "${file("${path.module}/files/install_nginx.sh")}"

  vars {
    nginx_config = "${replace(data.template_file.rancher_proxy_nginx_conf.rendered, "$", "\\$")}"
  }
}

data "template_file" "rancher_proxy_nginx_conf" {
  template = "${file("${path.module}/files/rancher_proxy_nginx.conf")}"
  vars {
    upstream_config = "${join("\n", formatlist("    server %s:8080;", triton_machine.rancher_master.*.primaryip))}"
  }
}

# Temporarily requiring the user to provide the private network with NAT enabled since
# creating one through terraform results in a network that can't be deleted.
# This is an issue with Triton: https://smartos.org/bugview/NAPI-258
data "triton_network" "gcm_private_network" {
  name = "${var.gcm_private_network_name}"
}

# Creating the nginx proxy for the gcm nodes
resource "triton_machine" "rancher_proxy" {
  # Only make rancher_proxy when HA is configured
  count = "${var.ha ? 2 : 0}"
  # Should package/image be hardcoded to a specific machine package?
  package = "${var.master_triton_machine_package}"
  image = "${data.triton_image.image.id}"
  name = "${var.name}-proxy-${count.index + 1}"

  # Attach to private network and public network
  networks = ["${data.triton_network.public.id}", "${data.triton_network.gcm_private_network.id}"]

  user_script = "${data.template_file.install_nginx.rendered}"

  cns = {
    services = ["${var.name}-proxy"]
  }

  affinity = ["role!=~gcm_proxy"]

  tags = {
    role = "gcm_proxy"
  }
}

# Creating the ssh bastion host for the gcm private network
resource "triton_machine" "rancher_ssh_bastion" {
  count = "${var.ha}"

  # Should package/image be hardcoded to a specific machine package?
  package = "${var.master_triton_machine_package}"
  image = "${data.triton_image.image.id}"
  name = "${var.name}-ssh"

  # Attach to private network and public network
  networks = ["${data.triton_network.public.id}", "${data.triton_network.gcm_private_network.id}"]
}

resource "triton_machine" "rancher_mysqldb" {
  count = "${var.ha}"

  package = "${coalesce(var.mysqldb_triton_machine_package, var.master_triton_machine_package)}"
  image   = "${data.triton_image.mysql_image.id}"
  name    = "${var.name}-mysqldb"

  networks = ["${data.triton_network.gcm_private_network.id}"]

  user_script = "${data.template_file.install_rancher_mysqldb.rendered}"
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

resource "triton_machine" "rancher_master" {
  count = "${var.gcm_node_count}"

  # Set to properly destroy masters before mysql.
  depends_on = ["triton_machine.rancher_mysqldb"]

  package = "${var.master_triton_machine_package}"
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-${count.index + 1}"

  user_script = "${data.template_file.install_docker.rendered}"

  # If HA is enabled, this attaches rancher masters to gcm_private_network
  # Otherwise, the rancher master is attached to triton_network.networks
  # join and split are used because terraform conditionals do not support list or maps
  # https://github.com/hashicorp/terraform/issues/12453
  networks = ["${split(",", var.ha ? data.triton_network.gcm_private_network.id : join(",", data.triton_network.networks.*.id))}"]

  cns = {
    services = ["${var.name}"]
  }

  affinity = ["role!=~gcm"]

  tags = {
    role = "gcm"
  }
}

data "template_file" "install_rancher_master" {
  template = "${file("${path.module}/files/install_rancher_master.sh.tpl")}"

  vars {
    ha = "${var.ha}"

    mysqldb_host              = "${coalesce(join("", triton_machine.rancher_mysqldb.*.primaryip), "")}"
    mysqldb_port              = "${var.mysqldb_port}"
    mysqldb_user              = "${var.mysqldb_username}"
    mysqldb_password          = "${var.mysqldb_password}"
    mysqldb_database_name     = "${var.mysqldb_database_name}"
    rancher_server_image      = "${var.rancher_server_image}"
    rancher_agent_image       = "${var.rancher_agent_image}"
    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "null_resource" "install_rancher_master" {
  count = "${var.gcm_node_count}"

  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_ids = "${join(",", triton_machine.rancher_master.*.id)}"
  }

  connection {
    type        = "ssh"
    user        = "${var.triton_ssh_user}"
    bastion_host = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
    host        = "${element(triton_machine.rancher_master.*.primaryip, count.index)}"
    private_key = "${file(var.triton_key_path)}"
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
    name         = "${var.name}"
    rancher_host = "http://127.0.0.1:8080"
    host_registration_url = "${local.rancher_url}"

    rancher_registry = "${var.rancher_registry}"
  }
}

resource "null_resource" "setup_rancher_k8s" {
  depends_on = ["null_resource.install_rancher_master"]

  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_ids = "${join(",", triton_machine.rancher_master.*.id)}"
  }

  # Setup script can run on any rancher master
  # So we just choose the first in this case
  connection {
    type        = "ssh"
    user        = "${var.triton_ssh_user}"
    bastion_host = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
    host        = "${element(triton_machine.rancher_master.*.primaryip, 0)}"
    private_key = "${file(var.triton_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.setup_rancher_k8s.rendered}
      EOF
  }
}
