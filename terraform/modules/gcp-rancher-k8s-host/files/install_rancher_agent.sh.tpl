#!/bin/sh
# This script just wraps https://raw.githubusercontent.com/joyent/triton-kubernetes/master/scripts/docker/17.03.sh
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

# Mounting the Volume
MOUNT_PATH='${disk_mount_path}'
if [ $$MOUNT_PATH != '' ]; then
	# For GCP Ubuntu instances, the device name for the GCP disk starts with 'sdb'.
	if [ -b /dev/sdb ]; then
		INSTANCE_STORE_BLOCK_DEVICE=/dev/sdb
	fi

	echo $${INSTANCE_STORE_BLOCK_DEVICE}

	if [ -b $${INSTANCE_STORE_BLOCK_DEVICE} ]; then
		sudo mke2fs -F -E nodiscard -L $$MOUNT_PATH -j $${INSTANCE_STORE_BLOCK_DEVICE} &&
		sudo tune2fs -r 0 $${INSTANCE_STORE_BLOCK_DEVICE} &&
		echo "LABEL=$$MOUNT_PATH     $$MOUNT_PATH           ext4    defaults,noatime  1   1" | sudo tee /etc/fstab > /dev/null &&
		sudo mkdir $$MOUNT_PATH &&
		sudo mount $$MOUNT_PATH
	fi
fi

sudo docker run -d --privileged --restart=unless-stopped --net=host -v /etc/kubernetes:/etc/kubernetes -v /var/run:/var/run ${rancher_agent_image} --server ${rancher_api_url} --token ${rancher_cluster_registration_token} --ca-checksum ${rancher_cluster_ca_checksum} --${rancher_node_role}
