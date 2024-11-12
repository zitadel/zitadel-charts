# Self Signed Example

For setting up a TLS enabled ZITADEL, you can start using a self-signed certificate.
By setting `zitadel.selfSignedCert.enabled` to true, the chart generates a self-signed cert for each zitadel pod on startup.
By running the commands below, you deploy a simple insecure Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]  
> You only pseudo-secure the incoming connections to ZITADEL, not to your database.
> Anybody with network access to the Postgres database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Postgres database, see [the secure Postgres example](../2-postgres-secure/README.md).

```bash
# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait db bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/7-self-signed/postgres-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/7-self-signed/zitadel-values.yaml
```

When ZITADEL is ready, you can access the GUI via port-forwarding:

```bash
kubectl port-forward svc/my-zitadel 8443:8080
```

Now, open https://my-iam.127.0.0.1.sslip.io:8443 in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.my-iam.127.0.0.1.sslip.io
**Password**: Password1!
