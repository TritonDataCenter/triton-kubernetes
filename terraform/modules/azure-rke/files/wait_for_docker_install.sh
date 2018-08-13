#!/bin/bash

# Wait for docker to be installed
printf 'Waiting for docker to be installed'
while [ -z "$(command -v docker)" ]; do
	printf '.'
	sleep 5
done

# Let things settle
# TODO Figure out why this is needed
sleep 30