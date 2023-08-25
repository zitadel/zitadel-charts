# Insecure Cockroach Example

By running the commands below, you deploy a simple insecure Cockroach database to your Kubernetes cluster [by using the Cockroach official chart](https://artifacthub.io/packages/helm/cockroachdb/cockroachdb).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]  
> Anybody with network access to the Cockroach database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Cockroach database, see [the secure Cockroach example](../4-cockroach-secure/README.md).

```bash
# Install Cockroach
helm repo add cockroachdb https://charts.cockroachdb.com/
helm install  --wait cockroach cockroachdb/cockroachdb --version 11.1.5 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/3-cocroach-insecure/cockroach-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/3-cockroach-insecure/zitadel-values.yaml
```
