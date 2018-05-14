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
    command = <<EOT
      cat > credentials-ark <<EOF
        [default]
        aws_access_key_id=${var.aws_access_key}
        aws_secret_access_key=${var.aws_secret_key}
      EOF
    EOT
  }

  provisioner "local-exec" {
    command = "kubectl apply -f ark-0.7.1/examples/common/00-prereqs.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "kubectl create secret generic cloud-credentials --namespace $ARK_SERVER_NAMESPACE --from-file cloud=credentials-ark --kubeconfig=kubeconfig.yaml --dry-run -o yaml | kubectl apply --kubeconfig=kubeconfig.yaml -f -"

    environment {
      ARK_SERVER_NAMESPACE = "heptio-ark-server"
    }
  }

  provisioner "local-exec" {
    command = <<EOT
      sed -i '.original' 's/<YOUR_BUCKET>/${var.aws_s3_bucket}/g' ark-0.7.1/examples/aws/00-ark-config.yaml
      sed -i '.original' 's/<YOUR_REGION>/${var.aws_region}/g' ark-0.7.1/examples/aws/00-ark-config.yaml
    EOT
  }

  provisioner "local-exec" {
    command = "kubectl apply -f ark-0.7.1/examples/aws/00-ark-config.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "kubectl apply -f ark-0.7.1/examples/common/10-deployment.yaml --kubeconfig=kubeconfig.yaml"
  }

  provisioner "local-exec" {
    command = "rm -rf ark ark-* credentials-ark kubeconfig.yaml v0.7.1.tar.gz"
  }
}
