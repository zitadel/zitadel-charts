[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# ZITADEL

## A Better Identity and Access Management Solution

ZITADEL combines the best of Auth0 and Keycloak.
It is built for the serverless era.

Learn more about ZITADEL by checking out the [source repository on GitHub](https://github.com/zitadel/zitadel)

## What's in the Chart

By default, this chart installs a highly available ZITADEL deployment.

## Install the Chart

Follow the [guide for deploying ZITADEL on Kubernetes](https://docs.zitadel.com/docs/guides/deploy/kubernetes).

## Upgrading from v3

:::Note
Apart from breaking changes in this chart, v4 also update the
default ZITADEL version to v2.14.2. For upgrading
ZITADEL, please refer to the
[ZITADEL release notes](https://github.com/zitadel/zitadel/releases/tag/v2.14.0).

This section is only relevant for existing releases with the
values property cockroachdb.enabled not set to false.

In v4, the cockroachdb chart dependency is removed.
We decided to go this way because:
- Maintaining two separate releases is easier, especially in production.
- We can use Helm hooks specific to ZITADEL.
- ZITADEL doesn't only support CockroachDB.

If you have cockroachdb.enabled=true in your values.yaml,
you need to make sure, that the cockroachdb chart is not
managed by the zitadel release anymore. The following
example for doing so uninstalls your entire zitadel
release, reinstalls cockroach using a dedicated release,
and then installs the new zitadel chart version.
The new cockroach release will take over the PersistentVolumeClaims
from the uninstalled chart, so no data migration is needed.
Nevertheless, we highly recommend making and testing a backup before upgrading.
Also note, that you will have downtime when
following the example while zitadel is uninstalled.

```bash
helm repo add cockroachdb https://charts.cockroachdb.com/
helm repo update cockroachdb zitadel
helm uninstall my-zitadel
helm install crdb cockroachdb/cockroachdb --version 8.1.8 --set fullnameOverride=crdb
helm install my-zitadel zitadel/zitadel --values ./my-zitadel-values.yaml
```


## Contributors

<a href="https://github.com/zitadel/zitadel-charts/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=zitadel/zitadel-charts" />
</a>
