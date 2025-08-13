# Secure Postgres Example

By running the commands below, you deploy a TLS and password authentication secured Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).
For creating the TLS certificates, we run a Kubernetes job that creates a self-signed CA certificate and client certificates and keys for the postgres admin DB user and for the zitadel DB user.

> [!INFO]
> The example assumes you already have a running Kubernetes cluster with a working ingress controller.
> If you don't, [run a local KinD cluster](../99-kind-with-traefik/README.md) before executing the follwing commands.

```bash
# Generate TLS certificates
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/certs-job.yaml
kubectl wait --for=condition=complete job/create-certs

# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait db bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/postgres-values.yaml

# Install Zitadel
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/1-postgres-insecure/zitadel-values.yaml
```

When Zitadel is ready, open https://pg-secure.127.0.0.1.sslip.io/ui/console?login_hint=zitadel-admin@zitadel.pg-secure.127.0.0.1.sslip.io in your browser and log in with the password `Password1!`.
