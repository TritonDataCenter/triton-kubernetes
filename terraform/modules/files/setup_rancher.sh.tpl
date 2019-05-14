#!/bin/bash

# Wait for Rancher UI to boot
printf 'Waiting for Rancher to start'
until $(curl --output /dev/null --silent --head --insecure --fail ${rancher_host}); do
	printf '.'
	sleep 5
done

# Wait for apt-get to be unlocked
printf 'Waiting for apt-get to unlock'
sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1;
while [ $? -ne 1 ]; do
	printf '.';
	sleep 5;
	sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1;
done

sudo apt-get install jq -y

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
	
# Setup helm
curl -X POST \
	--insecure \
	-u $access_key:$secret_key \
	-H 'Accept: application/json' \
	-H 'Content-Type: application/json' \
	-d '{"branch":"master", "kind":"helm", "name":"helm-charts", "url":"https://github.com/kubernetes/helm.git"}' \
	'${rancher_host}/v3/catalogs'

# Update graphics
printf 'Updating UI ...\n'
# logos
curl -sLO https://github.com/mesoform/triton-kubernetes/raw/master/static/logos.tar.gz -o ~/logos.tar.gz
tar -zxf ~/logos.tar.gz
sudo docker cp ~/logos/ $(sudo docker ps -q):/usr/share/rancher/ui/assets/images/
sudo docker exec -i $(sudo docker ps -q) bash <<EOF
sed -i 's/appName:"Rancher"/appName:"Triton Kubernetes"/g' /usr/share/rancher/ui/assets/ui-*.js
sed -i 's/appName\\":\\"Rancher\\"/appName\\":\\"Triton Kubernetes\\"/g' /usr/share/rancher/ui/assets/ui-*.map
echo 'footer.ember-view{visibility:hidden;display:none;}' >> /usr/share/rancher/ui/assets/vendor.css
echo 'footer.ember-view{visibility:hidden;display:none;}' >> /usr/share/rancher/ui/assets/vendor.rtl.css
EOF
rm -rf logos*