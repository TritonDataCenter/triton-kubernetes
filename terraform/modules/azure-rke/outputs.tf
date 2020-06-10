output "rancher_url" {
  value = "https://${var.fqdn}"
}

output "rancher_access_key" {
  value = data.external.rancher_server.result["name"]
}

output "rancher_secret_key" {
  value = data.external.rancher_server.result["token"]
}

output "rke_cluster_yaml" {
  sensitive = true
  value     = rke_cluster.cluster[0].rke_cluster_yaml
}

output "kube_config_yaml" {
  sensitive = true
  value     = rke_cluster.cluster[0].kube_config_yaml
}

