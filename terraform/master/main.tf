resource "triton_machine" "master" {
  name    = "${var.hostname}"
  package = "${var.package}"
  image   = "${var.image}"

  tags = {
    "name" = "${var.hostname}"
  }

  networks             = "${var.networks}"
  root_authorized_keys = "${var.root_authorized_keys}"

  provisioner "remote-exec" {
    connection {
      host        = "${triton_machine.master.primaryip}"
      user        = "ubuntu"
      private_key = "${var.root_authorized_keys}"
      agent       = false
    }

    inline = [
      "sleep 30",
      "sudo cp /home/ubuntu/.ssh/authorized_keys /root/.ssh/",
      "sudo apt-get update",
      "sudo apt-get install python-minimal -y",
    ]
  }

  provisioner "local-exec" {
    command = "echo ${triton_machine.master.primaryip} >> masters.ip"
  }

  tags = {
    hostname = "${var.hostname}"
  }
}
