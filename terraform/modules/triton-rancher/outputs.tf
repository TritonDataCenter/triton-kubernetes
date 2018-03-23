output "masters" {
  value = "${triton_machine.rancher_master.*.primaryip}"
}

output "rancher_url" {
  depends_on = ["triton_machine.rancher_proxy"]
  value      = "${local.rancher_url}"
}

output "ssh_bastion_ip" {
  value = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
}

output "rancher_internal_url" {
  value = "${local.rancher_internal_url}"
}

output "rancher_access_key" {
  value = "${chomp(module.rancher_access_key.stdout)}"
}

output "rancher_secret_key" {
  value = "${chomp(module.rancher_secret_key.stdout)}"
}

output "rancher_admin_username" {
  value = "${var.rancher_admin_username}"
}
