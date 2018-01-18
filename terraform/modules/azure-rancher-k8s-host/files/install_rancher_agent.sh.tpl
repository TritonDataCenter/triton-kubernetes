#!/bin/sh
# This script just wraps https://releases.rancher.com/install-docker/1.12.sh
# It disables firewalld on CentOS.
# TODO: Replace firewalld with iptables.

if [ -n "$(command -v firewalld)" ]; then
	sudo systemctl stop firewalld.service
	sudo systemctl disable firewalld.service
fi

sudo curl ${docker_engine_install_url} | sh

sudo service docker restart

sudo hostnamectl set-hostname ${hostname}
sudo bash -c 'echo "127.0.0.1 ${hostname}" >> /etc/hosts'

# Run docker login if requested
if [ "${rancher_registry_username}" != "" ]; then
	sudo docker login -u ${rancher_registry_username} -p ${rancher_registry_password} ${rancher_registry}
fi

# Run Rancher agent container
${rancher_agent_command}