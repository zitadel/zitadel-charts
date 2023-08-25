#!/bin/bash

set -e

NAMESPACE=$1

function createKey() {
  NAME=$1
  openssl genrsa -out ${NAME}.key 2048
  echo "created ${NAME}.key"
}

function createSigningRequest() {
  NAME=$1
  openssl req -new -key ${NAME}.key -extensions 'v3_req' -out ${NAME}.csr -config <(generateServerConfig)
  echo "created ${NAME}.csr"
}

function createKubernetesTLSSecret() {
  NAME=$1
  kubectl --namespace ${NAMESPACE} create secret tls ${NAME}-cert --cert=${NAME}.crt --key=${NAME}.key
}

function generateServerConfig() {
  cat<<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no
[req_distinguished_name]
CN = db-postgresql
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = DNS:postgres,DNS:zitadel,DNS:db-postgresql
EOF
}

function signCertificate() {
  INCSR=$1
  OUTCRT=$2
  CA_CRT=$3
  CA_KEY=$4

  openssl x509 -req -in $INCSR -CA $CA_CRT -CAkey $CA_KEY -CAcreateserial -days 365 -out $OUTCRT -extensions v3_req -extfile <(generateServerConfig)
}

# Create a CA key and cert for signing other certs
createKey myCA
openssl req -x509 -new -nodes -key myCA.key -days 365 -out myCA.crt -subj "/CN=My Custom CA"

createKey postgres
createSigningRequest postgres
signCertificate postgres.csr postgres.crt myCA.crt myCA.key
createKubernetesTLSSecret postgres

createKey zitadel
createSigningRequest zitadel
signCertificate zitadel.csr zitadel.crt myCA.crt myCA.key
createKubernetesTLSSecret zitadel

# The bitnami postgres chart expects the ca.crt to exist in the postgres-cert secret
kubectl --namespace ${NAMESPACE} patch secret postgres-cert --patch="{\"data\":{\"ca.crt\": \"$(cat myCA.crt | base64 --wrap 0)\"}}"
