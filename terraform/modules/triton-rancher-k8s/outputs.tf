output "rancher_environment_id" {
  value = "${rancher_environment.k8s.id}"
}

output "name" {
  value = "${var.name}"
}