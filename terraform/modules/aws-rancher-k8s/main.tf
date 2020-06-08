data "external" "rancher_cluster" {
  program = ["bash", "${path.module}/files/rancher_cluster.sh"]

  query = {
    rancher_api_url       = var.rancher_api_url
    rancher_access_key    = var.rancher_access_key
    rancher_secret_key    = var.rancher_secret_key
    name                  = var.name
    k8s_version           = var.k8s_version
    k8s_network_provider  = var.k8s_network_provider
    k8s_registry          = var.k8s_registry
    k8s_registry_username = var.k8s_registry_username
    k8s_registry_password = var.k8s_registry_password
  }
}

/* Setup our aws provider */
provider "aws" {
  version    = "~> 2.0"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = var.aws_region
}

/* Define our vpc */
resource "aws_vpc" "default" {
  cidr_block = var.aws_vpc_cidr

  tags = {
    Name = var.name
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = aws_vpc.default.id
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.default.id
  cidr_block              = var.aws_subnet_cidr
  map_public_ip_on_launch = true
  depends_on              = [aws_internet_gateway.default]

  tags = {
    Name = "public"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.default.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.default.id
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_key_pair" "deployer" {
  // Only attempt to create the key pair if the public key was provided
  count = var.aws_public_key_path != "" ? 1 : 0

  key_name   = var.aws_key_name
  public_key = file(var.aws_public_key_path)
}

# Firewall requirements taken from:
# https://rancher.com/docs/rancher/v2.0/en/quick-start-guide/
resource "aws_security_group" "rke_ports" {
  name        = var.name
  description = "Security group for rancher hosts in ${var.name} cluster"
  vpc_id      = aws_vpc.default.id

  ingress {
    from_port = "22" # SSH
    to_port   = "22"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "80" # Canal
    to_port   = "80"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "443" # Canal
    to_port   = "443"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "6443" # Canal
    to_port   = "6443"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "2379" # etcd server client API
    to_port   = "2380"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "10250" # kubelet API
    to_port   = "10250"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "10251" # scheduler
    to_port   = "10251"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "10252" # controller
    to_port   = "10252"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "10256" # kubeproxy
    to_port   = "10256"
    protocol  = "tcp"
    self      = true
  }

  ingress {
    from_port = "30000" # NodePort Services
    to_port   = "32767"
    protocol  = "tcp"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

