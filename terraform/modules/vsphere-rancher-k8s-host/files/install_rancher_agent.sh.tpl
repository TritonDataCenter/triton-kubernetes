#!/bin/sh
# This script just wraps https://raw.githubusercontent.com/mesoform/triton-kubernetes/master/scripts/docker/17.03.sh
# It disables firewalld on CentOS.
# TODO: Replace firewalld with iptables.

if [ -n "$(command -v firewalld)" ]; then
	sudo systemctl stop firewalld.service
	sudo systemctl disable firewalld.service
fi

sudo curl ${docker_engine_install_url} | sh
sudo service docker stop
sudo bash -c 'echo "{
  \"storage-driver\": \"overlay2\"
}" > /etc/docker/daemon.json'
sudo service docker restart

sudo hostnamectl set-hostname ${hostname}

# Run docker login if requested
if [ "${rancher_registry_username}" != "" ]; then
	sudo docker login -u ${rancher_registry_username} -p ${rancher_registry_password} ${rancher_registry}
fi

# Run Rancher agent container
sudo docker run -d --privileged --restart=unless-stopped --net=host -v /etc/kubernetes:/etc/kubernetes -v /var/run:/var/run ${rancher_agent_image} --server ${rancher_api_url} --token ${rancher_cluster_registration_token} --ca-checksum ${rancher_cluster_ca_checksum} --${rancher_node_role}
