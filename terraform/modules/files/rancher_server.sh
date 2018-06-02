#!/bin/bash

# SSH into a remote box and cat a file that contains the rancher server information.

# Exit if any of the intermediate steps fail
set -e

# Extract arguments from the input into shell variables.
# jq will ensure that the values are properly quoted
# and escaped for consumption by the shell.
eval "$(jq -r '@sh "id=\(.id) ssh_host=\(.ssh_host) ssh_user=\(.ssh_user) key_path=\(.key_path) file_path=\(.file_path)"')"

result=$(ssh -o UserKnownHostsFile=/dev/null \
	-o StrictHostKeyChecking=no \
	-i $key_path \
	$ssh_user@$ssh_host \
	"cat $file_path")

name=""
token=""
if [ "$result" != "" ]; then
	name=$(echo $result | jq -r .name)
	token=$(echo $result | jq -r .token | cut -d: -f2)
fi

# Safely produce a JSON object containing the result value.
# jq will ensure that the value is properly quoted
# and escaped to produce a valid JSON string.
jq -n --arg name "$name" \
	--arg token "$token" \
	'{"name":$name,"token":$token}'
