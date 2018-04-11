output "masters" {
  value = "${triton_machine.rancher_master.primaryip}"
}

output "rancher_url" {
  value = "https://${triton_machine.rancher_master.primaryip}"
}

output "rancher_access_key" {
  value = "${chomp(module.rancher_access_key.stdout)}"
}

output "rancher_secret_key" {
  value = "${chomp(module.rancher_secret_key.stdout)}"
}
