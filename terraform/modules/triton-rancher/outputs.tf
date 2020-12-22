output "rancher_url" {
  value = "https://${local.rancher_master_ip}"
}

output "rancher_access_key" {
  value = lookup(data.external.rancher_server.result, "name")
}

output "rancher_secret_key" {
  value = lookup(data.external.rancher_server.result, "token")
}
