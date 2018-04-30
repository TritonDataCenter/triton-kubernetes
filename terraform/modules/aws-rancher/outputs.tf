output "rancher_url" {
  value = "https://${local.rancher_master_ip}"
}

output "rancher_access_key" {
  value = "${chomp(module.rancher_access_key.stdout)}"
}

output "rancher_secret_key" {
  value = "${chomp(module.rancher_secret_key.stdout)}"
}
