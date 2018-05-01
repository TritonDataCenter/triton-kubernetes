#!/bin/bash

# Wait for Rancher UI to boot
printf 'Waiting for Rancher to start'
until $(curl --output /dev/null --silent --head --insecure --fail ${rancher_host}); do
    printf '.'
    sleep 5
done

sudo apt-get install jq -y || sudo yum install jq -y

# Login as default admin user
login_response=$(curl -X POST \
	--insecure \
	-d '{"description":"Initial Token", "password":"admin", "ttl": 60000, "username":"admin"}' \
	'${rancher_host}/v3-public/localProviders/local?action=login')
initial_token=$(echo $login_response | jq -r '.token')

# Create token
token_response=$(curl -X POST \
	--insecure \
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
	--insecure \
	-u $access_key:$secret_key \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"currentPassword":"admin","newPassword":"${rancher_admin_password}"}' \
	'${rancher_host}/v3/users?action=changepassword'

# Setup server url
curl -X PUT \
	--insecure \
	-u $access_key:$secret_key \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"baseType": "setting", "id": "server-url", "name": "server-url", "type": "setting", "value": "${host_registration_url}" }' \
	'${rancher_host}/v3/settings/server-url'

# Update graphics
printf 'Updating graphics'
curl -LO https://github.com/joyent/triton-kubernetes/raw/master/static/modified-graphics.tar.gz
tar -xf modified-graphics.tar.gz
for l in `ls modified-graphics/`
do
	sudo docker cp modified-graphics/${l} $(sudo docker ps -q):/usr/share/rancher/ui/assets/images/logos/
done
rm -rf modified-graphics*