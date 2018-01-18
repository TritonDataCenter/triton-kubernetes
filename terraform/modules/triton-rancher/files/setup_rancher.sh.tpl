#!/bin/bash

sudo apt-get install jq -y || sudo yum install jq -y

# Wait for Rancher UI to boot
printf 'Waiting for Rancher to start'
until $(curl --output /dev/null --silent --head --fail ${rancher_host}); do
    printf '.'
    sleep 5
done

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

# Update default registry to private registry if requested
if [ "${rancher_registry}" != "" ]; then
	curl -X PUT \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d '{"id":"registry.default","type":"activeSetting","baseType":"setting","name":"registry.default","activeValue":"","inDb":false,"source":"Code Packaged Defaults","value":"${rancher_registry}"}' \
		'${rancher_host}/v2-beta/settings/registry.default'
fi