#!/bin/bash

# HA setup requires use of FQDN. However DNS on FQDN might not be setup so we need to set it in /etc/hosts
sudo sed -i '$ a 127.0.0.1 ${fqdn}' /etc/hosts

# Wait for Rancher UI to boot
printf 'Waiting for Rancher to start'
until $(curl --output /dev/null --silent --head --fail ${rancher_host}); do
    printf '.'
    sleep 5
done

sudo apt-get install jq -y || sudo yum install jq -y

# Login as default admin user
login_response=$(curl -X POST \
	-d '{"description":"Initial Token", "password":"admin", "ttl": 60000, "username":"admin"}' \
	'${rancher_host}/v3-public/localProviders/local?action=login')
initial_token=$(echo $login_response | jq -r '.token')

# Create token
token_response=$(curl -X POST \
	-u $initial_token \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"expired":false,"isDerived":false,"ttl":0,"type":"token","description":"Managed by Terraform","name":"triton-kubernetes"}' \
	'${rancher_host}/v3/token')
echo $token_response > ~/rancher_api_key
access_key=$(echo $token_response | jq -r '.name')
secret_key=$(echo $token_response | jq -r '.token' | cut -d: -f2)

# Change default admin password
curl -X POST \
	-u $access_key:$secret_key \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"currentPassword":"admin","newPassword":"${rancher_admin_password}"}' \
	'${rancher_host}/v3/users?action=changepassword'

# Setup server url
curl -X PUT \
	-u $access_key:$secret_key \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"baseType": "setting", "id": "server-url", "name": "server-url", "type": "setting", "value": "${rancher_host}" }' \
	'${rancher_host}/v3/settings/server-url'

# Remove FQDN reference to localhost
sudo sed -i '$ d' /etc/hosts