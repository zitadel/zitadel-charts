#!/bin/bash

set -e

NAMESPACE=$1

function createKey() {
  DBUSER=$1
  openssl genrsa -out ${DBUSER}.key 2048
  echo "created ${DBUSER}.key"
}

function createSigningRequest() {
  DBUSER=$1
  openssl req -new -key ${DBUSER}.key -subj "/CN=${DBUSER}" -out ${DBUSER}.csr
  echo "created ${DBUSER}.csr"
}

function createKubernetesTLSSecret() {
  DBUSER=$1
  kubectl --namespace ${NAMESPACE} create secret tls ${DBUSER}-cert --cert=${DBUSER}.crt --key=${DBUSER}.key
}

createKey postgres
createSigningRequest postgres
openssl x509 -req -in postgres.csr -signkey postgres.key -days 365 -out postgres.crt
createKubernetesTLSSecret postgres

createKey zitadel
createSigningRequest zitadel
openssl x509 -req -in zitadel.csr -CA postgres.crt -CAkey postgres.key -CAcreateserial -days 365 -out zitadel.crt
createKubernetesTLSSecret zitadel

kubectl --namespace ${NAMESPACE} create secret generic postgres-ca --from-file=ca.crt=postgres.crt
