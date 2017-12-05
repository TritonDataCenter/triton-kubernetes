#!/bin/bash

# This is a hack to get around the Terraform Rancher provider not having resources for environment templates.
# This script tries to be idempotent by checking if an environment template with the same name already exists.
# This script violates the spirit of data sources in Terraform since it does mutate infrastructure.

# Exit if any of the intermediate steps fail
set -e

# Extract "name" argument from the input into name shell variable.
# jq will ensure that the values are properly quoted
# and escaped for consumption by the shell.
eval "$(jq -r '@sh "rancher_api_url=\(.rancher_api_url) name=\(.name) k8s_plane_isolation=\(.k8s_plane_isolation) k8s_registry=\(.k8s_registry)"')"

kube_template=''
template_already_existed=false
if [ "$(curl --silent $rancher_api_url/v2-beta/projecttemplates/?name=$name | jq -r '.data | length')" != "0" ]; then
	# Look to see if a template exists with the same name
	kube_template=$(curl -X GET \
		-H 'Accept: application/json' \
		"$rancher_api_url/v2-beta/projecttemplates/?name=$name")
	template_already_existed=true
else
	# otherwise clone default kubernetes template
	kube_template=$(curl -X GET \
		-H 'Accept: application/json' \
		"$rancher_api_url/v2-beta/projecttemplates/?name=kubernetes")
fi

kube_template=$(echo $kube_template | jq '.data[0]')
kube_template=$(echo $kube_template | jq --arg k8s_plane_isolation "$k8s_plane_isolation" '(.stacks[] | select(.name == "kubernetes") | .answers.CONSTRAINT_TYPE) |= $k8s_plane_isolation')
kube_template=$(echo $kube_template | jq --arg k8s_registry "$k8s_registry" '(.stacks[] | select(.name == "kubernetes") | .answers.REGISTRY) |= $k8s_registry')


TEMPLATE_ID=''
if [ "$template_already_existed" = true ]; then
	# update existing template
	TEMPLATE_ID=$(echo $kube_template | jq -r '.id')

	output=$(curl -X PUT \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d "$kube_template" \
		"$rancher_api_url/v2-beta/projecttemplates/$TEMPLATE_ID")
else
	# since we're cloning the default kubernetes template, clear out the ids
	kube_template=$(echo $kube_template | jq --arg name "$name" '. + {"name": $name,"description": "","isPublic": true,"uuid":"","id":""}')

	# create new template
	TEMPLATE_ID=$(curl -X POST \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d "$kube_template" \
		"$rancher_api_url/v2-beta/projecttemplates" | jq -r '.id')
fi

# Safely produce a JSON object containing the result value.
# jq will ensure that the value is properly quoted
# and escaped to produce a valid JSON string.
jq -n --arg template_id "$TEMPLATE_ID" '{"id":$template_id}'