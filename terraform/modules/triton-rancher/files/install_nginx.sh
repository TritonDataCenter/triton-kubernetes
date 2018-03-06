#!/bin/bash

# Installing nginx
sudo apt-get update -y
sudo apt-get install nginx -y

# Installing the SSL private key
PRIVATE_DIR=/etc/ssl/private
if [ ! -d "$PRIVATE_DIR" ]; then
    sudo mkdir $PRIVATE_DIR
    sudo chmod 700 $PRIVATE_DIR
fi

PRIVATE_KEYNAME="nginx-selfsigned.key"
PRIVATE_KEY="${ssl_private_key}"
echo "$PRIVATE_KEY" | sudo tee $PRIVATE_DIR/$PRIVATE_KEYNAME > /dev/null

# Installing the SSL Certificate
CERT_DIR=/etc/ssl/certs
CERT_NAME=nginx-selfsigned.crt
CERT="${ssl_cert}"
echo "$CERT" | sudo tee $CERT_DIR/$CERT_NAME > /dev/null

# Removing nginx default config
sudo rm -f /etc/nginx/sites-enabled/default

# Adding nginx rancher_proxy config
NGINX_CONFIG="${nginx_config}"
echo "$NGINX_CONFIG" | sudo tee /etc/nginx/sites-available/rancher_proxy > /dev/null
sudo ln -s /etc/nginx/sites-available/rancher_proxy /etc/nginx/sites-enabled/rancher_proxy

# Restarting nginx to reload config
sudo service nginx restart
