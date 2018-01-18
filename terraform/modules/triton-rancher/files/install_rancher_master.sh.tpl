#!/bin/bash

# Install Docker
sudo curl "${docker_engine_install_url}" | sh

# Needed on CentOS, TODO: Replace firewalld with iptables.
sudo service firewalld stop

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

# Run docker login if requested
if [ "${rancher_registry_username}" != "" ]; then
	sudo docker login -u ${rancher_registry_username} -p ${rancher_registry_password} ${rancher_registry}
fi

# Run Rancher docker container
container_id=""
if [ "${ha}" = 1 ]; then
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 -p 9345:9345 -e CATTLE_BOOTSTRAP_REQUIRED_IMAGE=${rancher_agent_image} ${rancher_server_image} \
		--db-host ${mysqldb_host} --db-port ${mysqldb_port} --db-user ${mysqldb_user} --db-pass ${mysqldb_password} --db-name ${mysqldb_database_name} \
		--advertise-address $(ip route get 1 | awk '{print $NF;exit}'))
else
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 -e CATTLE_BOOTSTRAP_REQUIRED_IMAGE=${rancher_agent_image} ${rancher_server_image})
fi

# Copy private key to Rancher host
echo "${triton_key_material}" > triton_key
sudo docker exec $container_id mkdir -p /root/.ssh/
sudo docker cp triton_key $container_id:/root/.ssh/id_rsa
sudo docker exec $container_id chmod 0700 /root/.ssh/id_rsa
rm triton_key
