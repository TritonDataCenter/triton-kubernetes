provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

data "triton_network" "networks" {
  count = "${length(var.triton_network_names)}"
  name  = "${element(var.triton_network_names, count.index)}"
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

resource "triton_machine" "rancher_mysqldb" {
  count = "${var.ha}"

  package = "${coalesce(var.mysqldb_triton_machine_package, var.master_triton_machine_package)}"
  image   = "${data.triton_image.mysql_image.id}"
  name    = "${var.name}-mysqldb"

  networks = ["${data.triton_network.networks.*.id}"]

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

  networks = ["${data.triton_network.networks.*.id}"]

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
    primary_ip   = "${element(triton_machine.rancher_master.*.primaryip, 0)}"

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
    host        = "${element(triton_machine.rancher_master.*.primaryip, 0)}"
    private_key = "${file(var.triton_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.setup_rancher_k8s.rendered}
      EOF
  }
}
