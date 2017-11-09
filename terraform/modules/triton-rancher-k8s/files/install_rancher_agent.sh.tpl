#!/bin/sh
# This script just wraps https://releases.rancher.com/install-docker/1.12.sh
# It disables firewalld on CentOS.
# TODO: Replace firewalld with iptables.

if [ -n "$(command -v firewalld)" ]; then
	sudo systemctl stop firewalld.service
	sudo systemctl disable firewalld.service
fi

sudo curl "https://releases.rancher.com/install-docker/1.12.sh" | sh

sudo service docker restart

sudo hostnamectl set-hostname ${hostname}

# Run Rancher agent container
${rancher_agent_command}