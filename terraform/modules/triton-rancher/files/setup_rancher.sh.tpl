#!/bin/bash

# Wait for Rancher UI to boot
printf 'Waiting for Rancher to start'
until $(curl --output /dev/null --silent --head --fail ${rancher_host}); do
    printf '.'
    sleep 5
done

sudo apt-get install jq -y || sudo yum install jq -y

# Install docker-machine-driver-triton
driver_id=$(curl -X POST \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"url":"${docker_machine_driver_triton_url}", "uiUrl":"${rancher_ui_driver_triton}", "builtin":false, "activateOnCreate":true, "checksum":"${docker_machine_driver_triton_checksum}"}' \
	'${rancher_host}/v2-beta/machinedrivers' | jq -r '.id')

# Wait for docker-machine-driver-triton to become active
printf 'Waiting for docker-machine-driver-triton to become active'
while [ "$(curl --silent ${rancher_host}/v2-beta/machinedrivers/$driver_id | jq -r '.state')" != "active" ]; do
	printf '.'
	sleep 5
done

# Setup api.host
curl -X PUT \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"id":"api.host","type":"activeSetting","baseType":"setting","name":"api.host","activeValue":null,"inDb":false,"source":null,"value":"http://${primary_ip}:8080"}' \
	'${rancher_host}/v2-beta/settings/api.host'

# Modify kubernetes template plane isolation
sleep 5
kube_template=$(curl -X GET \
	-H 'Accept: application/json' \
	'${rancher_host}/v2-beta/projecttemplates/?name=kubernetes')

kube_template_id=$(echo $kube_template | jq -r '.data[0].id')
kube_template=$(echo $kube_template | jq '.data[0]')
kube_template=$(echo $kube_template | jq '(.stacks[] | select(.name == "kubernetes") | .answers) |= {"CONSTRAINT_TYPE":"required"}')

curl -X PUT \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d "$kube_template" \
	'${rancher_host}/v2-beta/projecttemplates/'$kube_template_id