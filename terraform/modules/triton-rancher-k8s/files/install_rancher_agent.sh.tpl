#!/bin/sh
# Disable firewalld on CentOS.
# TODO: Replace firewalld with iptables.

if [ -n "$(command -v firewalld)" ]; then
	sudo systemctl stop firewalld.service
	sudo systemctl disable firewalld.service
fi

sudo curl ${docker_engine_install_url} | sh

sudo service docker stop
DOCKER_SERVICE=$(systemctl status docker.service --no-pager | grep Loaded | sed 's~\(.*\)loaded (\(.*\)docker.service\(.*\)$~\2docker.service~g')
sed 's~ExecStart=/usr/bin/dockerd -H\(.*\)~ExecStart=/usr/bin/dockerd --graph="/mnt/docker" -H\1~g' $DOCKER_SERVICE > /home/ubuntu/docker.conf && sudo mv /home/ubuntu/docker.conf $DOCKER_SERVICE
sudo mkdir /mnt/docker
sudo bash -c "mv /var/lib/docker/* /mnt/docker/"
sudo rm -rf /var/lib/docker
sudo bash -c 'echo "{
  \"storage-driver\": \"overlay2\"
}" > /etc/docker/daemon.json'
sudo systemctl daemon-reload
sudo systemctl restart docker

sudo hostnamectl set-hostname ${hostname}

# Run docker login if requested
if [ "${rancher_registry_username}" != "" ]; then
	sudo docker login -u ${rancher_registry_username} -p ${rancher_registry_password} ${rancher_registry}
fi

# Run Rancher agent container
${rancher_agent_command}
