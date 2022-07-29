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
- the password for the CockroachDB user (ZITADEL creates the user if it doesn't exist already)

Checkout the [values.yaml for all configuration options](https://github.com/zitadel/zitadel-charts/blob/main/charts/zitadel/values.yaml).

## Install the Chart

```bash
# Add the helm repository
helm repo add zitadel https://charts.zitadel.com

# generate masterkey and database user password (store them securely)
ZITADEL_MASTERKEY=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32)
ZITADEL_CRDB_PASSWORD=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32)

# install a zitadel release that is accessible via port forwarding to localhost
helm install --namespace zitadel --create-namespace my-zitadel zitadel/zitadel \
  --set zitadel.masterkey=${ZITADEL_MASTERKEY} \
  --set zitadel.secretConfig.Database.cockroach.User.Password=${ZITADEL_CRDB_PASSWORD} 
```

Enjoy watching a highly available and secure ZITADEL instance starting up in less than a minute.
The following GIF was made with a local KinD cluster on a 32 RAM and 8 CPU cores machine.
![watch pods](https://github.com/zitadel/zitadel-charts/raw/main/watch-pods.gif "Watch Pods")

## Login

Use the zitadel admin user for the initial login:
- *username*: zitadel-admin@zitadel.localhost
- *password*: Password1!

## Configuring ZITADEL

For learning how to run ZITADEL in production, please read the following docs:
- [configuration docs](https://docs.zitadel.com/docs/guides/manage/self-hosted/configure)
- [load balancing example docs](https://docs.zitadel.com/docs/guides/deploy/loadbalancing-example)

