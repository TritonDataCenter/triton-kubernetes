#!/bin/bash

# This is a hack to get around the Terraform Rancher provider not supporting Rancher 2.0.
# This script tries to be idempotent by checking if a cluster with the same name already exists.
# This script violates the spirit of data sources in Terraform since it does mutate infrastructure.

# Exit if any of the intermediate steps fail
set -e

# Extract arguments from the input into shell variables.
# jq will ensure that the values are properly quoted
# and escaped for consumption by the shell.
eval "$(jq -r '@sh "rancher_api_url=\(.rancher_api_url) rancher_access_key=\(.rancher_access_key) rancher_secret_key=\(.rancher_secret_key) name=\(.name) k8s_version=\(.k8s_version) k8s_network_provider=\(.k8s_network_provider) k8s_registry=\(.k8s_registry) k8s_registry_username=\(.k8s_registry_username) k8s_registry_password=\(.k8s_registry_password)"')"

cluster_id=''
cluster_already_existed=false
cluster_search=$(curl -X GET \
	--insecure \
	-u $rancher_access_key:$rancher_secret_key \
	-H 'Accept: application/json' \
	"$rancher_api_url/v3/clusters?name=$name")
# Look to see if a cluster exists with the same name
if [ "$(echo $cluster_search | jq -r '.data | length')" != "0" ]; then
	cluster_already_existed=true
	cluster_id=$(echo $cluster_search | jq -r '.data[0].id')
else
	k8s_registry_json=''
	if [ $k8s_registry != "" ]; then
		k8s_registry_json=',"privateRegistries":[{"url":"'$k8s_registry'","user":"'$k8s_registry_username'","password":"'$k8s_registry_password'"}]'
	fi

	# Create cluster
	cluster_response=$(curl -X POST \
		--insecure \
		-u $rancher_access_key:$rancher_secret_key \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d '{"type":"cluster","googleKubernetesEngineConfig":null,"name":"'$name'","rancherKubernetesEngineConfig":{"ignoreDockerVersion":false,"sshAgentAuth":false,"type":"rancherKubernetesEngineConfig","kubernetesVersion":"'$k8s_version'","authentication":{"type":"authnConfig","strategy":"x509"},"network":{"type":"networkConfig","plugin":"'$k8s_network_provider'"},"services":{"type":"rkeConfigServices","kubeApi":{"podSecurityPolicy":false,"type":"kubeAPIService"}}'$k8s_registry_json'},"id":""}' \
		"$rancher_api_url/v3/cluster")
	cluster_id=$(echo $cluster_response | jq -r '.id')
fi

# Cluster registration token
registration_token='';
if [ "$cluster_already_existed" == true ]; then
	# Get existing registration token
	get_registration_token_response=$(curl -X GET \
		--insecure \
		-u $rancher_access_key:$rancher_secret_key \
		-H 'Accept: application/json' \
		"$rancher_api_url/v3/clusters/$cluster_id/clusterregistrationtokens")

	registration_token=$(echo $get_registration_token_response | jq -r '.data[0].token')
else
	# Create cluster registration token
	create_registration_token_response=$(curl -X POST \
		--insecure \
		-u $rancher_access_key:$rancher_secret_key \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d '{"clusterId":"'$cluster_id'","type":"clusterRegistrationToken"}' \
		"$rancher_api_url/v3/clusterregistrationtoken")

	registration_token=$(echo $create_registration_token_response | jq -r '.token')
fi

# Retrieve CA checksum
cacerts_response=$(curl -X GET \
	--insecure \
	-u $rancher_access_key:$rancher_secret_key \
	-H 'Accept: application/json' \
	"$rancher_api_url/v3/settings/cacerts")
ca_checksum=$(echo $cacerts_response | jq -r .value | shasum -a 256 | awk '{ print $1 }')

# Safely produce a JSON object containing the result value.
# jq will ensure that the value is properly quoted
# and escaped to produce a valid JSON string.
jq -n --arg cluster_id "$cluster_id" \
	--arg registration_token "$registration_token" \
	--arg ca_checksum "$ca_checksum" \
	'{"cluster_id":$cluster_id,"registration_token":$registration_token,"ca_checksum":$ca_checksum}'
