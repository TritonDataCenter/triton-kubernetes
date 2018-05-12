data "external" "rancher_cluster" {
  program = ["bash", "${path.module}/files/rancher_cluster_import.sh"]

  query = {
    rancher_api_url    = "${var.rancher_api_url}"
    rancher_access_key = "${var.rancher_access_key}"
    rancher_secret_key = "${var.rancher_secret_key}"
    name               = "${var.name}"
  }
}

provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id       = "${var.azure_client_id}"
  client_secret   = "${var.azure_client_secret}"
  tenant_id       = "${var.azure_tenant_id}"
  environment     = "${var.azure_environment}"
}

resource "azurerm_resource_group" "resource_group" {
  name     = "${var.name}-resource_group"
  location = "${var.azure_location}"
}

resource "azurerm_kubernetes_cluster" "primary" {
  name                = "${var.name}"
  location            = "${azurerm_resource_group.resource_group.location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
  dns_prefix          = "${var.name}"

  kubernetes_version = "${var.k8s_version}"

  linux_profile {
    admin_username = "${var.azure_ssh_user}"

    ssh_key {
      key_data = "${file(var.azure_public_key_path)}"
    }
  }

  agent_pool_profile {
    name    = "default"
    count   = "${var.node_count}"
    vm_size = "${var.azure_size}"
  }

  service_principal {
    client_id     = "${var.azure_client_id}"
    client_secret = "${var.azure_client_secret}"
  }
}

locals {
  kube_config_path = "./${var.name}_config"
}

# Bootstrap rancher in aks environment
resource "null_resource" "import_rancher" {
  triggers {
    cluster = "${azurerm_kubernetes_cluster.primary.id}"
  }

  provisioner "local-exec" {
    command = "${format("cat << EOF > %s \n%s\nEOF", local.kube_config_path, azurerm_kubernetes_cluster.primary.kube_config_raw)}"
  }

  provisioner "local-exec" {
    command = "curl --insecure -sfL ${var.rancher_api_url}/v3/import/${data.external.rancher_cluster.result.registration_token}.yaml | kubectl apply -f -"

    environment {
      KUBECONFIG = "${local.kube_config_path}"
    }
  }

  provisioner "local-exec" {
    command = "rm ${local.kube_config_path}"
  }
}
