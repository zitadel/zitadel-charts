[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# ZITADEL

## A Better Identity and Access Management Solution

ZITADEL combines the best of Auth0 and Keycloak.
It is built for the serverless era.

Learn more about ZITADEL by checking out the [source repository on GitHub](https://github.com/zitadel/zitadel)

## What the Chart Brings

By default, this chart installs a highly available ZITADEL instance.
Also it installs a highly available and secure CockroachDB by default.

## What the Chart Needs You to Bring
For the ZITADEL startup configuration, you need to bring:
- the masterkey that ZITADEL uses for symmetric encryption
- the password for the CockroachDB user (ZITADEL creates the user if it doesn't exist)
- the domain where you will publish your ZITADEL instance

Checkout the values.yaml for all configuration options.
<!--TODO: link values.yaml -->

## Install the Chart

```bash
# Add the helm repository
helm repo add zitadel https://zitadel.github.io/zitadel-charts

# generate keys (store them securely)
ZITADEL_MASTERKEY=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32)
ZITADEL_CRDB_PASSWORD=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32)

# configure your domain
ZITADEL_DOMAIN=zitadel.mydomain.com

helm install --namespace zitadel --create-namespace my-zitadel zitadel/zitadel \
  --set zitadel.masterkey=${ZITADEL_MASTERKEY} \
  --set zitadel.secretConfig.Database.User.Password=${ZITADEL_CRDB_PASSWORD} \
  --set zitadel.configmapConfig.ExternalDomain=${ZITADEL_DOMAIN} 
```

Enjoy watching a highly available and secure ZITADEL instance starting up.
<!--TODO: gif here -->

## Route traffic to ZITADEL

You can use the Ingress resource from the ZITADEL chart to route traffic to ZITADEL.
Make sure your network sends end-to-end HTTP/2 traffic.

