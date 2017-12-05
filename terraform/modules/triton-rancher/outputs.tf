output "masters" {
  value = "${triton_machine.rancher_master.*.primaryip}"
}
