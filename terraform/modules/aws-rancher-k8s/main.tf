/* Setup our aws provider */
provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.aws_region}"
}

/* Define our vpc */
resource "aws_vpc" "default" {
  cidr_block = "${var.aws_vpc_cidr}"

  tags {
    Name = "${var.name}"
  }
}

provider "rancher" {
  api_url = "${var.rancher_api_url}"
}

data "external" "rancher_environment_template" {
  program = ["bash", "${path.module}/files/rancher_environment_template.sh"]

  query = {
    rancher_api_url     = "${var.rancher_api_url}"
    name                = "${var.name}-kubernetes"
    k8s_plane_isolation = "${var.k8s_plane_isolation}"
    k8s_registry        = "${var.k8s_registry}"
  }
}

resource "rancher_environment" "k8s" {
  name                = "${var.name}"
  project_template_id = "${data.external.rancher_environment_template.result.id}"
}

resource "rancher_registry" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name           = "${var.rancher_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.rancher_registry}"
}

resource "rancher_registry_credential" "rancher_registry" {
  count = "${var.rancher_registry != "" ? 1 : 0}"

  name         = "${var.rancher_registry}"
  registry_id  = "${rancher_registry.rancher_registry.id}"
  email        = "${var.rancher_registry_username}"
  public_value = "${var.rancher_registry_username}"
  secret_value = "${var.rancher_registry_password}"
}

resource "rancher_registry" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name           = "${var.k8s_registry}"
  environment_id = "${rancher_environment.k8s.id}"
  server_address = "${var.k8s_registry}"
}

resource "rancher_registry_credential" "k8s_registry" {
  count = "${var.k8s_registry != "" && var.rancher_registry != var.k8s_registry ? 1 : 0}"

  name         = "${var.k8s_registry}"
  registry_id  = "${rancher_registry.k8s_registry.id}"
  email        = "${var.k8s_registry_username}"
  public_value = "${var.k8s_registry_username}"
  secret_value = "${var.k8s_registry_password}"
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

resource "aws_security_group" "rancher" {
  name        = "${var.name}"
  description = "Security group for rancher hosts in ${var.name} cluster"
  vpc_id      = "${aws_vpc.default.id}"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 500
    to_port     = 500
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 4500
    to_port     = 4500
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = "0"
    to_port   = "0"
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "rancher_registration_token" "etcd" {
  count = "${var.etcd_node_count}"

  name           = "${var.name}-etcd-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-etcd-${count.index + 1} host"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    etcd = "true"
  }
}

data "template_file" "install_rancher_agent_etcd" {
  count = "${var.etcd_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-etcd-${count.index + 1}"
    rancher_agent_command     = "${element(rancher_registration_token.etcd.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "aws_instance" "etcd" {
  count = "${var.etcd_node_count}"

  ami                    = "${var.aws_ami_id}"
  instance_type          = "${var.etcd_aws_instance_type}"
  subnet_id              = "${aws_subnet.public.id}"
  vpc_security_group_ids = ["${aws_security_group.rancher.id}"]
  key_name               = "${var.aws_public_key_path != "" ? aws_key_pair.deployer.key_name : var.aws_key_name}"

  tags = {
    Name = "${var.name}-etcd-${count.index + 1}"
  }

  user_data = "${element(data.template_file.install_rancher_agent_etcd.*.rendered, count.index)}"
}

resource "rancher_registration_token" "orchestration" {
  count = "${var.orchestration_node_count}"

  name           = "${var.name}-orchestration-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-orchestration-${count.index + 1} host"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    orchestration = "true"
  }
}

data "template_file" "install_rancher_agent_orchestration" {
  count = "${var.orchestration_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-orchestration-${count.index + 1}"
    rancher_agent_command     = "${element(rancher_registration_token.orchestration.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "aws_instance" "orchestration" {
  count = "${var.orchestration_node_count}"

  ami                    = "${var.aws_ami_id}"
  instance_type          = "${var.orchestration_aws_instance_type}"
  subnet_id              = "${aws_subnet.public.id}"
  vpc_security_group_ids = ["${aws_security_group.rancher.id}"]
  key_name               = "${var.aws_public_key_path != "" ? aws_key_pair.deployer.key_name : var.aws_key_name}"

  tags = {
    Name = "${var.name}-orchestration-${count.index + 1}"
  }

  user_data = "${element(data.template_file.install_rancher_agent_orchestration.*.rendered, count.index)}"
}

resource "rancher_registration_token" "compute" {
  count = "${var.compute_node_count}"

  name           = "${var.name}-compute-${count.index + 1}_token"
  description    = "Registration token for ${var.name}-compute-${count.index + 1} host"
  environment_id = "${rancher_environment.k8s.id}"

  host_labels {
    compute = "true"
  }
}

data "template_file" "install_rancher_agent_compute" {
  count = "${var.compute_node_count}"

  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.name}-compute-${count.index + 1}"
    rancher_agent_command     = "${element(rancher_registration_token.compute.*.command, count.index)}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "aws_instance" "compute" {
  count = "${var.compute_node_count}"

  ami                    = "${var.aws_ami_id}"
  instance_type          = "${var.compute_aws_instance_type}"
  subnet_id              = "${aws_subnet.public.id}"
  vpc_security_group_ids = ["${aws_security_group.rancher.id}"]
  key_name               = "${var.aws_public_key_path != "" ? aws_key_pair.deployer.key_name : var.aws_key_name}"

  tags = {
    Name = "${var.name}-compute-${count.index + 1}"
  }

  user_data = "${element(data.template_file.install_rancher_agent_compute.*.rendered, count.index)}"
}
