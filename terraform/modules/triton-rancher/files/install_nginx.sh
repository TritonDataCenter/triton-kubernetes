#!/bin/bash

# Installing nginx
sudo apt-get update -y
sudo apt-get install nginx -y


PRIVATE_DIR=/etc/ssl/private
PRIVATE_KEYNAME="nginx-selfsigned.key"
PRIVATE_KEY="${ssl_private_key}"

CERT_DIR=/etc/ssl/certs
CERT_NAME=nginx-selfsigned.crt
CERT="${ssl_cert}"

if [ "$PRIVATE_KEY" != "" ] && [ "$CERT" != "" ]; then
    if [ ! -d "$PRIVATE_DIR" ]; then
        sudo mkdir $PRIVATE_DIR
        sudo chmod 700 $PRIVATE_DIR
    fi

    # Installing the SSL Certificate
    echo "$PRIVATE_KEY" | sudo tee $PRIVATE_DIR/$PRIVATE_KEYNAME > /dev/null
    echo "$CERT" | sudo tee $CERT_DIR/$CERT_NAME > /dev/null
fi

# Removing nginx default config
sudo rm -f /etc/nginx/sites-enabled/default

# Adding nginx rancher_proxy config
NGINX_CONFIG="${nginx_config}"
echo "$NGINX_CONFIG" | sudo tee /etc/nginx/sites-available/rancher_proxy > /dev/null
sudo ln -s /etc/nginx/sites-available/rancher_proxy /etc/nginx/sites-enabled/rancher_proxy

# Restarting nginx to reload config
sudo service nginx restart
