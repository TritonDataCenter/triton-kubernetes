#!/bin/bash

# Wait for docker to be installed
printf 'Waiting for docker to be installed'
while [ -z "$(command -v docker)" ]; do
	printf '.'
	sleep 5
done

# Wait for rancher_server_image to finish downloading
printf 'Waiting for Rancher Server Image to download'
while [ -z "$(sudo docker images -q ${rancher_server_image})" ]; do
	printf '.'
	sleep 5
done

# Run Rancher docker container
container_id=""
if [ "${ha}" = 1 ]; then
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 -p 9345:9345 -e CATTLE_BOOTSTRAP_REQUIRED_IMAGE=${rancher_agent_image} ${rancher_server_image} \
		--db-host ${mysqldb_host} --db-port ${mysqldb_port} --db-user ${mysqldb_user} --db-pass ${mysqldb_password} --db-name ${mysqldb_database_name} \
		--advertise-address $(ip route get 1 | awk '{print $NF;exit}'))
else
	container_id=$(sudo docker run -d --restart=unless-stopped -p 8080:8080 -e CATTLE_BOOTSTRAP_REQUIRED_IMAGE=${rancher_agent_image} ${rancher_server_image})
fi
