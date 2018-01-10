output "rancher_environment_id" {
  value = "${rancher_environment.k8s.id}"
}

output "aws_subnet_id" {
  value = "${aws_subnet.public.id}"
}

output "aws_security_group_id" {
  value = "${aws_security_group.rancher.id}"
}

output "aws_key_name" {
  value = "${aws_key_pair.deployer.key_name}"
}
