#!/bin/bash

PRIVATE_KEY_NAME=nginx-selfsigned.key
CERT_NAME=nginx-selfsigned.crt

# Generating the SSL private key and certificate and
# leaving them in the HOME directory.
sudo openssl req -x509 -nodes -days 365 \
    -newkey rsa:2048 \
    -subj "/C=US/ST=California/L=Los Angeles/O=joyent/OU=triton-kubernetes/CN=${common_name}" \
    -keyout $PRIVATE_KEY_NAME \
    -out $CERT_NAME
