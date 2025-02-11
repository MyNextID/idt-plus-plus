#!/bin/bash

root="certs"
mkdir -p $root
cd $root

# Note: serial number can act as a index in the compact CRL
crtSerialNumber=1

# Generate an elliptic curve private key for the root CA
openssl ecparam -name prime256v1 -genkey -noout -out rootCA.key

# Create a root CA certificate signing request (CSR)
openssl req -new -key rootCA.key -subj "/CN=My Root CA" -out rootCA.csr

# Create a root CA certificate with the crlSign key usage
openssl x509 -req -in rootCA.csr -signkey rootCA.key -out rootCA.crt -days 3650 -sha256 \
  -extfile <(printf "basicConstraints=critical,CA:TRUE\nkeyUsage=critical,keyCertSign,cRLSign\nsubjectKeyIdentifier=hash")

# # Create a self-signed root CA certificate
# openssl req -new -x509 -key rootCA.key -sha256 -days 3650 -subj "/CN=Root CA" -out rootCA.crt
 
# Generate an elliptic curve private key for the org certificate
openssl ecparam -name prime256v1 -genkey -noout -out org.key

# Create a certificate signing request (CSR) for the org certificate
openssl req -new -key org.key -subj "/CN=My Org Cert" -out org.csr

# Generate the org certificate signed by the root CA
openssl x509 -req -in org.csr -CA rootCA.crt -CAkey rootCA.key -set_serial ${crtSerialNumber} -CAcreateserial -out org.crt -days 365 -sha256
