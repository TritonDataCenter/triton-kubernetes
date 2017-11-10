#!/bin/bash

# Exit if any of the intermediate steps fail
set -e

# Extract "name" argument from the input into name shell variable.
# jq will ensure that the values are properly quoted
# and escaped for consumption by the shell.
eval "$(jq -r '@sh "rancher_api_url=\(.rancher_api_url) name=\(.name)"')"

TEMPLATE_ID=$(curl "$rancher_api_url/v2-beta/projecttemplates/?name=$name" | jq -r '.data[0].id')

# Safely produce a JSON object containing the result value.
# jq will ensure that the value is properly quoted
# and escaped to produce a valid JSON string.
jq -n --arg template_id "$TEMPLATE_ID" '{"id":$template_id}'