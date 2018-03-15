provider "triton" {
  account      = "${var.triton_account}"
  key_material = "${file(var.triton_key_path)}"
  key_id       = "${var.triton_key_id}"
  url          = "${var.triton_url}"
}

locals {
  using_custom_tls_cert = "${var.rancher_tls_private_key_path != "" && var.rancher_tls_cert_path != ""}"
  rancher_fqdn          = "${var.ha ? format("%s-proxy.svc.%s.us-east-1.triton.zone", lower(var.name), data.triton_account.main.id) : element(triton_machine.rancher_master.*.primaryip, 0)}"
  rancher_internal_url  = "${local.using_custom_tls_cert ? format("https://%s", local.rancher_fqdn) : format("http://%s", local.rancher_fqdn)}"
  rancher_url           = "${local.using_custom_tls_cert ? format("https://%s", var.rancher_domain_name) : format("http://%s", local.rancher_fqdn)}"
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

data "local_file" "custom_tls_private_key" {
  count    = "${local.using_custom_tls_cert ? 1 : 0}"
  filename = "${var.rancher_tls_private_key_path}"
}

data "local_file" "custom_tls_cert" {
  count    = "${local.using_custom_tls_cert ? 1 : 0}"
  filename = "${var.rancher_tls_cert_path}"
}

data "template_file" "install_nginx" {
  template = "${file("${path.module}/files/install_nginx.sh")}"

  vars {
    nginx_config = "${local.using_custom_tls_cert ?
      replace(data.template_file.rancher_proxy_nginx_https_conf.rendered, "$", "\\$") :
      replace(data.template_file.rancher_proxy_nginx_http_conf.rendered, "$", "\\$")}"

    triton_ssh_user = "${var.triton_ssh_user}"
    ssl_private_key = "${element(coalescelist(data.local_file.custom_tls_private_key.*.content, tls_private_key.generated_ssl_key.*.private_key_pem, list("")), 0)}"
    ssl_cert        = "${element(coalescelist(data.local_file.custom_tls_cert.*.content, tls_self_signed_cert.generated_ssl_cert.*.cert_pem, list("")), 0)}"
  }
}

data "template_file" "rancher_proxy_nginx_http_conf" {
  template = "${file("${path.module}/files/rancher_proxy_nginx_http.conf")}"

  vars {
    # For HA, the upstream servers point to the rancher master ip addresses
    # For non-HA, the nginx proxy is on the same box as the rancher master so upstream is set to 127.0.0.1
    upstream_config = "${var.ha ? join("\n", formatlist("    server %s:8080;", triton_machine.rancher_master.*.primaryip)) : "    server 127.0.0.1:8080;"}"
  }
}

data "template_file" "rancher_proxy_nginx_https_conf" {
  template = "${file("${path.module}/files/rancher_proxy_nginx_https.conf")}"

  vars {
    # For HA, the upstream servers point to the rancher master ip addresses
    # For non-HA, the nginx proxy is on the same box as the rancher master so upstream is set to 127.0.0.1
    upstream_config = "${var.ha ? join("\n", formatlist("    server %s:8080;", triton_machine.rancher_master.*.primaryip)) : "    server 127.0.0.1:8080;"}"

    rancher_domain_name = "${var.rancher_domain_name}"
  }
}

# Self signed SSL Certificate and Key
# NOTE: Rancher Agents currently do not work with Rancher masters that 
# use self-signed SSL Certificates.
# Temporarily disabling self-signing certificates until a solution is found.
resource "tls_private_key" "generated_ssl_key" {
  # count = "${local.using_custom_tls_cert ? 0 : 1}"
  count     = 0
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_self_signed_cert" "generated_ssl_cert" {
  # count = "${local.using_custom_tls_cert ? 0 : 1}"
  count                 = 0
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.generated_ssl_key.private_key_pem}"
  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]

  subject {
    common_name         = "${local.rancher_fqdn}"
    organization        = "Joyent"
    organizational_unit = "triton-kubernetes"
    locality            = "San Francisco"
    province            = "California"
    country             = "US"
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
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-proxy-${count.index + 1}"

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
  image   = "${data.triton_image.image.id}"
  name    = "${var.name}-ssh"

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

# For non-HA Rancher, this installs nginx on the rancher master itself since there
# are no nginx proxy instances.
resource "null_resource" "install_rancher_master_nginx" {
  depends_on = ["null_resource.install_rancher_master"]

  count = "${var.ha ? 0 : 1}"

  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_ids = "${join(",", triton_machine.rancher_master.*.id)}"
  }

  connection {
    type         = "ssh"
    user         = "${var.triton_ssh_user}"
    bastion_host = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
    host         = "${element(triton_machine.rancher_master.*.primaryip, count.index)}"
    private_key  = "${file(var.triton_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.install_nginx.rendered}
      EOF
  }
}

resource "null_resource" "install_rancher_master" {
  count = "${var.gcm_node_count}"

  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    rancher_master_ids = "${join(",", triton_machine.rancher_master.*.id)}"
  }

  connection {
    type         = "ssh"
    user         = "${var.triton_ssh_user}"
    bastion_host = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
    host         = "${element(triton_machine.rancher_master.*.primaryip, count.index)}"
    private_key  = "${file(var.triton_key_path)}"
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
    rancher_host          = "http://127.0.0.1:8080"
    host_registration_url = "${local.rancher_url}"

    rancher_registry       = "${var.rancher_registry}"
    rancher_admin_username = "${var.rancher_admin_username}"
    rancher_admin_password = "${var.rancher_admin_password}"
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
    type         = "ssh"
    user         = "${var.triton_ssh_user}"
    bastion_host = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
    host         = "${element(triton_machine.rancher_master.*.primaryip, 0)}"
    private_key  = "${file(var.triton_key_path)}"
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
locals {
  ssh_proxy_command_arg = "-o ProxyCommand='ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${var.triton_key_path} -W %h:%p ${var.triton_ssh_user}@${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}'"
}

module "rancher_access_key" {
  source  = "matti/outputs/shell"
  version = "0.0.1"

  // We ssh into the remote box and cat the file.
  // We echo the output from null_resource.setup_rancher_k8s to setup an implicit dependency.
  command = "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ${var.ha ? local.ssh_proxy_command_arg : ""} -i ${var.triton_key_path} ${var.triton_ssh_user}@${element(triton_machine.rancher_master.*.primaryip, 0)} 'echo ${null_resource.setup_rancher_k8s.id} > /dev/null; cat ~/rancher_api_key | jq -r .publicValue'"
}

module "rancher_secret_key" {
  source  = "matti/outputs/shell"
  version = "0.0.1"

  // We ssh into the remote box and cat the file.
  // We echo the output from null_resource.setup_rancher_k8s to setup an implicit dependency.
  command = "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ${var.ha ? local.ssh_proxy_command_arg : ""} -i ${var.triton_key_path} ${var.triton_ssh_user}@${element(triton_machine.rancher_master.*.primaryip, 0)} 'echo ${null_resource.setup_rancher_k8s.id} > /dev/null; cat ~/rancher_api_key | jq -r .secretValue'"
}
