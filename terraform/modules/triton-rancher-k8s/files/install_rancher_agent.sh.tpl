#!/bin/sh
# This script just wraps https://releases.rancher.com/install-docker/1.12.sh
# It disables firewalld on CentOS.
# TODO: Replace firewalld with iptables.

if [ -n "$(command -v firewalld)" ]; then
	sudo systemctl stop firewalld.service
	sudo systemctl disable firewalld.service
fi

sudo curl ${docker_engine_install_url} | sh

sudo service docker stop
sudo mkdir /etc/systemd/system/docker.service.d
cat >>/home/ubuntu/docker.conf <<EOF
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd --graph="/mnt/docker"
EOF
sudo mkdir /etc/systemd/system/docker.service.d/
sudo mv /home/ubuntu/docker.conf /etc/systemd/system/docker.service.d/
sudo chown root:root /etc/systemd/system/docker.service.d/docker.conf
sudo mkdir /mnt/docker
sudo bash -c "mv /var/lib/docker/* /mnt/docker/"
sudo rm -rf /var/lib/docker
sudo systemctl daemon-reload
sudo service docker restart

sudo hostnamectl set-hostname ${hostname}

# Run docker login if requested
if [ "${rancher_registry_username}" != "" ]; then
	sudo docker login -u ${rancher_registry_username} -p ${rancher_registry_password} ${rancher_registry}
fi

# Run Rancher agent container
${rancher_agent_command}
