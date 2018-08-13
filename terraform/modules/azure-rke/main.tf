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

resource "azurerm_virtual_network" "vnet" {
  name                = "${var.azure_virtual_network_name}"
  address_space       = ["${var.azure_virtual_network_address_space}"]
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

resource "azurerm_subnet" "subnet" {
  name                 = "${var.azure_subnet_name}"
  resource_group_name  = "${azurerm_resource_group.resource_group.name}"
  virtual_network_name = "${azurerm_virtual_network.vnet.name}"
  address_prefix       = "${var.azure_subnet_address_prefix}"
}

resource "azurerm_network_security_group" "firewall" {
  name                = "${var.azurerm_network_security_group_name}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

resource "azurerm_application_security_group" "asg" {
  name                = "${var.name}-asg"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

# Firewall requirements taken from:
# https://rancher.com/docs/rke/v0.1.x/en/installation/os/#ports
resource "azurerm_network_security_rule" "public_rancher_ports" {
  name      = "public_rancher_ports_tcp"
  priority  = 1000
  direction = "Inbound"
  access    = "Allow"
  protocol  = "Tcp"

  destination_port_ranges = [
    "22",   # SSH
    "80",   # Rancher UI
    "443",  # Rancher UI
    "6443", # Kubernetes API
  ]

  destination_address_prefix = "*"

  source_port_range     = "*"
  source_address_prefix = "Internet"

  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

resource "azurerm_network_security_rule" "internal_rancher_ports_tcp" {
  name      = "internal_rancher_ports_tcp"
  priority  = 1001
  direction = "Inbound"
  access    = "Allow"
  protocol  = "Tcp"

  destination_port_ranges = [
    "2376",
    "2379",
    "2380",
    "10250",
    "6443",
  ]

  destination_address_prefix = "*"

  source_port_range = "*"

  source_application_security_group_ids = [
    "${azurerm_application_security_group.asg.id}",
  ]

  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

resource "azurerm_network_security_rule" "internal_rancher_ports_udp" {
  name      = "internal_rancher_ports_udp"
  priority  = 1002
  direction = "Inbound"
  access    = "Allow"
  protocol  = "Udp"

  destination_port_ranges = [
    "8472",
  ]

  destination_address_prefix = "*"

  source_port_range = "*"

  source_application_security_group_ids = [
    "${azurerm_application_security_group.asg.id}",
  ]

  resource_group_name         = "${azurerm_resource_group.resource_group.name}"
  network_security_group_name = "${azurerm_network_security_group.firewall.name}"
}

resource "azurerm_public_ip" "public_ip" {
  count = "${var.node_count}"

  name                         = "${var.name}-${count.index}"
  location                     = "${var.azure_location}"
  resource_group_name          = "${azurerm_resource_group.resource_group.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_network_interface" "nic" {
  count = "${var.node_count}"

  name                = "${var.name}-${count.index}"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"

  network_security_group_id = "${azurerm_network_security_group.firewall.id}"

  ip_configuration {
    name                          = "${var.name}-${count.index}-ip"
    subnet_id                     = "${azurerm_subnet.subnet.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${element(azurerm_public_ip.public_ip.*.id, count.index)}"

    application_security_group_ids = [
      "${azurerm_application_security_group.asg.id}",
    ]
  }
}

data "template_file" "install_docker" {
  template = "${file("${path.module}/files/install_docker_rancher.sh.tpl")}"

  vars {
    docker_engine_install_url = "${var.docker_engine_install_url}"

    rancher_server_image      = "${var.rancher_server_image}"
    rancher_registry          = "${var.rancher_registry}"
    rancher_registry_username = "${var.rancher_registry_username}"
    rancher_registry_password = "${var.rancher_registry_password}"
  }
}

resource "azurerm_virtual_machine" "host" {
  count = "${var.node_count}"

  name                  = "${var.name}-${count.index}"
  location              = "${var.azure_location}"
  resource_group_name   = "${azurerm_resource_group.resource_group.name}"
  network_interface_ids = ["${element(azurerm_network_interface.nic.*.id, count.index)}"]
  vm_size               = "${var.azure_size}"

  delete_os_disk_on_termination = true

  storage_image_reference {
    publisher = "${var.azure_image_publisher}"
    offer     = "${var.azure_image_offer}"
    sku       = "${var.azure_image_sku}"
    version   = "${var.azure_image_version}"
  }

  storage_os_disk {
    name              = "${var.name}-${count.index}-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "${var.name}-${count.index}"
    admin_username = "${var.azure_ssh_user}"
    custom_data    = "${data.template_file.install_docker.rendered}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/${var.azure_ssh_user}/.ssh/authorized_keys"
      key_data = "${file(var.azure_public_key_path)}"
    }
  }
}

data "azurerm_public_ip" "public_ip" {
  depends_on = ["azurerm_public_ip.public_ip"]

  count = "${var.node_count}"

  name                = "${element(azurerm_public_ip.public_ip.*.name, count.index)}"
  resource_group_name = "${azurerm_resource_group.resource_group.name}"
}

data "template_file" "wait_for_docker_install" {
  template = "${file("${path.module}/files/wait_for_docker_install.sh")}"
}

resource "null_resource" "wait_for_docker_install" {
  # Changes to any instance of the cluster requires re-provisioning
  triggers {
    node_ids = "${join(",", azurerm_virtual_machine.host.*.id)}"
  }

  count = "${var.node_count}"

  connection {
    type        = "ssh"
    user        = "${var.azure_ssh_user}"
    host        = "${element(data.azurerm_public_ip.public_ip.*.ip_address, count.index)}"
    private_key = "${file(var.azure_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.wait_for_docker_install.rendered}
      EOF
  }
}

data rke_node_parameter "nodes" {
  count = "${var.node_count}"

  internal_address = "${element(azurerm_network_interface.nic.*.private_ip_address, count.index)}"
  address          = "${element(data.azurerm_public_ip.public_ip.*.ip_address, count.index)}"
  user             = "${var.azure_ssh_user}"
  role             = ["controlplane", "etcd", "worker"]
  ssh_key          = "${file(var.azure_private_key_path)}"
}

resource rke_cluster "cluster" {
  depends_on = ["null_resource.wait_for_docker_install"]

  nodes_conf = ["${data.rke_node_parameter.nodes.*.json}"]

  ingress = {
    provider = "nginx"

    extra_args = {
      enable-ssl-passthrough = ""
    }
  }

  addons = <<EOL
---
kind: Namespace
apiVersion: v1
metadata:
  name: cattle-system
---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: cattle-admin
  namespace: cattle-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cattle-crb
  namespace: cattle-system
subjects:
- kind: ServiceAccount
  name: cattle-admin
  namespace: cattle-system
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Secret
metadata:
  name: cattle-keys-ingress
  namespace: cattle-system
type: Opaque
data:
  tls.crt: ${base64encode(file(var.tls_cert_path))}
  tls.key: ${base64encode(file(var.tls_private_key_path))}
---
apiVersion: v1
kind: Service
metadata:
  namespace: cattle-system
  name: cattle-service
  labels:
    app: cattle
spec:
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http
  - port: 443
    targetPort: 443
    protocol: TCP
    name: https
  selector:
    app: cattle
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: cattle-system
  name: cattle-ingress-http
  annotations:
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "1800"   # Max time in seconds for ws to remain shell window open
    nginx.ingress.kubernetes.io/proxy-send-timeout: "1800"   # Max time in seconds for ws to remain shell window open
spec:
  rules:
  - host: ${var.fqdn}
    http:
      paths:
      - backend:
          serviceName: cattle-service
          servicePort: 80
  tls:
  - secretName: cattle-keys-ingress
    hosts:
    - ${var.fqdn}
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  namespace: cattle-system
  name: cattle
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: cattle
    spec:
      serviceAccountName: cattle-admin
      containers:
      - image: ${var.rancher_server_image}
        args:
        - --no-cacerts
        imagePullPolicy: Always
        name: cattle-server
        ports:
        - containerPort: 80
          protocol: TCP
        - containerPort: 443
          protocol: TCP
EOL
}

data "template_file" "setup_rancher_k8s" {
  template = "${file("${path.module}/files/setup_rancher.sh.tpl")}"

  vars {
    name         = "${var.name}"
    fqdn         = "${var.fqdn}"
    rancher_host = "https://${var.fqdn}"

    rancher_admin_password = "${var.rancher_admin_password}"
  }
}

resource "null_resource" "setup_rancher_k8s" {
  depends_on = ["rke_cluster.cluster"]

  triggers {
    node_id = "${azurerm_virtual_machine.host.0.id}"
  }

  connection {
    type        = "ssh"
    user        = "${var.azure_ssh_user}"
    host        = "${data.azurerm_public_ip.public_ip.0.ip_address}"
    private_key = "${file(var.azure_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = <<EOF
      ${data.template_file.setup_rancher_k8s.rendered}
      EOF
  }
}

// The setup_rancher_k8s script will have stored a file with an api key
// We need to retrieve the contents of that file and output it.
// This is a hack to get around the Terraform Rancher provider not having resources for api keys.
data "external" "rancher_server" {
  program = ["bash", "${path.module}/files/rancher_server.sh"]

  query = {
    id        = "${null_resource.setup_rancher_k8s.id}"            // used to create an implicit dependency
    ssh_host  = "${data.azurerm_public_ip.public_ip.0.ip_address}"
    ssh_user  = "${var.azure_ssh_user}"
    key_path  = "${var.azure_private_key_path}"
    file_path = "~/rancher_api_key"
  }
}
