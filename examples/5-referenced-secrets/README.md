# Referenced Secrets Example

To enrich the ZITADEL configuration with secrets, that are already available in the cluster, you can use the `configSecretName` property in the ZITADEL values file.
This is especially handy in case you manage your Kubernetes secrets using a secret manager like [Hashicorp Vault](https://www.vaultproject.io/).

By running the commands below, you deploy a simple insecure Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]
> Anybody with network access to the Postgres database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Postgres database, see [the secure Postgres example](../2-postgres-secure/README.md).

```bash
# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait db bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/5-referenced-secrets/postgres-values.yaml

# Create a secret for arbitrary ZITADEL configuration as well as the ZITADEL masterkey
kubectl apply --filename https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/5-referenced-secrets/zitadel-masterkey.yaml,https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/5-referenced-secrets/zitadel-secrets.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/5-referenced-secrets/zitadel-values.yaml
```

When ZITADEL is ready, you can access the GUI via port-forwarding:

```bash
kubectl port-forward svc/my-zitadel 8080
```

Now, open http://localhost:8080 in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.localhost
**Password**: Password1!
