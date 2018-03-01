#!/bin/bash

# Installing nginx
sudo apt-get update -y
sudo apt-get install nginx -y
sudo ufw allow 'Nginx Full'

# Removing nginx default config
sudo rm -f /etc/nginx/sites-enabled/default

# Adding nginx rancher_proxy config
NGINX_CONFIG="${nginx_config}"
sudo echo "$NGINX_CONFIG" > /etc/nginx/sites-available/rancher_proxy
sudo ln -s /etc/nginx/sites-available/rancher_proxy /etc/nginx/sites-enabled/rancher_proxy

# Restarting nginx to reload config
sudo service nginx restart
