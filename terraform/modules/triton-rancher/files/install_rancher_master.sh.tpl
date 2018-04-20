#!/bin/bash

# Wait for docker to be installed
printf 'Waiting for docker to be installed'
while [ -z "$(command -v docker)" ]; do
	printf '.'
	sleep 5
done

# Wait for rancher_server_image to finish downloading
printf 'Waiting for Rancher Server Image to download\n'
while [ -z "$(sudo docker images -q ${rancher_server_image})" ]; do
	printf '.'
	sleep 5
done

# Run Rancher docker container
sudo docker run -d --restart=unless-stopped -p 80:80 -p 443:443 ${rancher_server_image}
