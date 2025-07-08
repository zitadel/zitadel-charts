# Insecure Postgres Example

By running the commands below, you deploy a simple insecure Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]  
> Anybody with network access to the Postgres database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Postgres database, see [the secure Postgres example](../2-postgres-secure/README.md).

```bash
# Install Traefik
helm repo add traefik https://traefik.github.io/charts
helm install --wait traefik traefik/traefik --version 36.3.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/traefik-values.yaml

# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait db bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/postgres-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/zitadel-values.yaml
```

When ZITADEL is ready, you can open https://zitadel.127.0.0.1.sslip.io/ui/console in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.127.0.0.1.sslip.io  
**Password**: Password1!
