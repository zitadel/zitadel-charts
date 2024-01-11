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
# Beware that the "--wait" option doesn't work with the CockroachDB chart (read more in note below).
helm install db cockroachdb/cockroachdb --version 11.2.1 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/3-cockroach-insecure/cockroach-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/3-cockroach-insecure/zitadel-values.yaml
```

> [!NOTE]
> The --wait option doesn't work with the CockroachDB chart, because its init job is done by a post-install hook.
> With --wait, the installation is only done when the Cockroach pods are ready and the init job only starts when the installation is done.
> This is a deadlock, because the CockroachDB pods don't become ready without the CockroachDB init job having completed.

When ZITADEL is ready, you can access the GUI via port-forwarding:

```bash
kubectl port-forward svc/my-zitadel 8080
```

Now, open http://127.0.0.1.sslip.io:8080 in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.127.0.0.1.sslip.io
**Password**: Password1!
