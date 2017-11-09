#!/bin/bash

# Install Docker
sudo curl "${docker_engine_install_url}" | sh

# Needed on CentOS, TODO: Replace firewalld with iptables.
sudo service firewalld stop

sudo service docker restart

# Run Rancher docker container
container_id=""
if [ "${ha}" = 1 ]; then
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 -p 9345:9345 rancher/server:stable \
		--db-host ${mysqldb_host} --db-port ${mysqldb_port} --db-user ${mysqldb_user} --db-pass ${mysqldb_password} --db-name ${mysqldb_database_name} \
		--advertise-address $(ip route get 1 | awk '{print $NF;exit}'))
else
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 rancher/server:stable)
fi

# Copy private key to Rancher host
echo "${triton_key_material}" > triton_key
sudo docker exec $container_id mkdir -p /root/.ssh/
sudo docker cp triton_key $container_id:/root/.ssh/id_rsa
sudo docker exec $container_id chmod 0700 /root/.ssh/id_rsa
rm triton_key