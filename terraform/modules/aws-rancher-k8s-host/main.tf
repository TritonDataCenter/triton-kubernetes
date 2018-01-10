provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.aws_region}"
}

provider "rancher" {
  api_url = "${var.rancher_api_url}"
}

resource "rancher_registration_token" "token" {
  name           = "${var.hostname}"
  description    = "Registration token for ${var.hostname}"
  environment_id = "${var.rancher_environment_id}"

  host_labels = "${var.rancher_host_labels}"
}

data "template_file" "install_rancher_agent" {
  template = "${file("${path.module}/files/install_rancher_agent.sh.tpl")}"

  vars {
    hostname                  = "${var.hostname}"
    rancher_agent_command     = "${rancher_registration_token.token.command}"
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "aws_instance" "host" {
  ami                    = "${var.aws_ami_id}"
  instance_type          = "${var.aws_instance_type}"
  subnet_id              = "${var.aws_subnet_id}"
  vpc_security_group_ids = ["${var.aws_security_group_id}"]
  key_name               = "${var.aws_key_name}"

  tags = {
    Name = "${var.hostname}"
  }

  user_data = "${data.template_file.install_rancher_agent.rendered}"
}
