output "masters" {
  value = "${triton_machine.rancher_master.*.primaryip}"
}

output "rancher_url" {
  depends_on = ["triton_machine.rancher_proxy"]
  value = "${local.rancher_url}"
}

output "ssh_bastion_ip" {
  value = "${element(coalescelist(triton_machine.rancher_ssh_bastion.*.primaryip, list("")), 0)}"
}
