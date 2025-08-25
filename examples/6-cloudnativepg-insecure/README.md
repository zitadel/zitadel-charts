# Insecure CloudNativePG Postgres Example

By running the commands below, you deploy a simple insecure Postgres database to your Kubernetes cluster [by using the CloudNativePG (CNPG)](https://github.com/cloudnative-pg/cloudnative-pg).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]  
> Anybody with network access to the Postgres database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Postgres database, see [the secure Postgres example](../2-postgres-secure/README.md).

> [!INFO]
> The example assumes you already have a running Kubernetes cluster with a working ingress controller.
> If you don't, [run a local KinD cluster](../99-kind-with-traefik/README.md) before executing the follwing commands.

```bash
# Install Postgres
kubectl apply --server-side -f \
  https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.27/releases/cnpg-1.27.0.yaml
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/6-cloudnativepg-insecure/postgres-cluster.yaml


# Install Zitadel
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/zitadel-values.yaml
```

When Zitadel is ready, open https://pg-insecure.127.0.0.1.sslip.io/ui/console?login_hint=zitadel-admin@zitadel.pg-insecure.127.0.0.1.sslip.io in your browser and log in with the password `Password1!`.
