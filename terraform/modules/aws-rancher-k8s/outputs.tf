output "rancher_cluster_id" {
  value = "${lookup(data.external.rancher_cluster.result, "cluster_id")}"
}

output "rancher_cluster_registration_token" {
  value = "${lookup(data.external.rancher_cluster.result, "registration_token")}"
}

output "rancher_cluster_ca_checksum" {
  value = "${lookup(data.external.rancher_cluster.result, "ca_checksum")}"
}

output "aws_subnet_id" {
  value = "${aws_subnet.public.id}"
}

output "aws_security_group_id" {
  value = "${aws_security_group.rke_ports.id}"
}

output "aws_key_name" {
  value = "${var.aws_key_name}"
}
