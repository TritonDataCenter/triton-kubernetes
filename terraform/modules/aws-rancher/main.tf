provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.aws_region}"
}

resource "aws_vpc" "default" {
  cidr_block = "${var.aws_vpc_cidr}"

  tags {
    Name = "${var.name}"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"
}

resource "aws_subnet" "public" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "${var.aws_subnet_cidr}"
  map_public_ip_on_launch = true
  depends_on              = ["aws_internet_gateway.default"]

  tags {
    Name = "public"
  }
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.default.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.default.id}"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = "${aws_subnet.public.id}"
  route_table_id = "${aws_route_table.public.id}"
}

resource "aws_key_pair" "deployer" {
  // Only attempt to create the key pair if the public key was provided
  count = "${var.aws_public_key_path != "" ? 1 : 0}"

  key_name   = "${var.aws_key_name}"
  public_key = "${file("${var.aws_public_key_path}")}"
}

# Firewall requirements taken from:
# https://rancher.com/docs/rancher/v2.0/en/quick-start-guide/
resource "aws_security_group" "rke_ports" {
  name        = "${var.name}"
  description = "Security group for rancher hosts in ${var.name} cluster"
  vpc_id      = "${aws_vpc.default.id}"

  ingress {
    from_port   = "22"          # SSH
    to_port     = "22"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = "80"          # Rancher UI
    to_port     = "80"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = "443"         # Rancher UI
    to_port     = "443"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "host" {
  ami                    = "${var.aws_ami_id}"
  instance_type          = "${var.aws_instance_type}"
  subnet_id              = "${aws_subnet.public.id}"
  vpc_security_group_ids = ["${aws_security_group.rke_ports.id}"]
  key_name               = "${var.aws_key_name}"

  tags = {
    Name = "${var.name}"
  }

  user_data = "${data.template_file.install_docker.rendered}"
}

locals {
  rancher_master_id = "${aws_instance.host.id}"
  rancher_master_ip = "${aws_instance.host.public_ip}"
  ssh_user          = "${var.aws_ssh_user}"
  key_path          = "${var.aws_private_key_path}"
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
data "external" "rancher_server" {
  program = ["bash", "${path.module}/files/rancher_server.sh"]

  query = {
    id        = "${null_resource.setup_rancher_k8s.id}" // used to create an implicit dependency
    ssh_host  = "${local.rancher_master_ip}"
    ssh_user  = "${local.ssh_user}"
    key_path  = "${local.key_path}"
    file_path = "~/rancher_api_key"
  }
}
