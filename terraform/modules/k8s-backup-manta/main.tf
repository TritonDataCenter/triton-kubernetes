data "template_file" "minio_manta_deployment" {
  template = "${file("${path.module}/files/minio-manta-deployment.yaml")}"

  vars {
    triton_account  = "${var.triton_account}"
    triton_key_id   = "${var.triton_key_id}"
    triton_key_path = "${var.triton_key_path}"
    manta_subuser   = "${var.manta_subuser}"
  }
}

resource "null_resource" "setup_ark_backup" {
  provisioner "local-exec" {
    command = "curl -LO https://github.com/heptio/ark/releases/download/v0.7.1/ark-v0.7.1-linux-arm64.tar.gz"
  }

  provisioner "local-exec" {
    command = "tar xvf ark-v0.7.1-linux-arm64.tar.gz"
  }

  provisioner "local-exec" {
    command = "curl -LO https://github.com/heptio/ark/archive/v0.7.1.tar.gz"
  }

  provisioner "local-exec" {
    command = "tar xvf v0.7.1.tar.gz"
  }

  provisioner "local-exec" {
    # Get kubernetes config yaml from Rancher, write it to disk
    command = <<EOT
      curl -X POST \
        --insecure \
        -u ${var.rancher_access_key}:${var.rancher_secret_key} \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        -d '' \
        '${var.rancher_api_url}/v3/clusters/${var.rancher_cluster_id}?action=generateKubeconfig' | jq -r '.config' > kubeconfig.yaml
    EOT
  }

  provisioner "local-exec" {
    # Write minio_manta_deployment.yaml to disk
    command = "${format("cat << EOF > minio_manta_deployment.yaml \n%s\nEOF", data.template_file.minio_manta_deployment.rendered)}"
  }

  provisioner "local-exec" {
    command = "kubectl apply -f ark-0.7.1/examples/common/00-prereqs.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "kubectl apply -f minio_manta_deployment.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "kubectl apply -f ark-0.7.1/examples/common/10-deployment.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "rm -rf ark ark-* minio_manta_deployment.yaml kubeconfig.yaml v0.7.1.tar.gz"
  }
}
