# Secure Postgres Example

By running the commands below, you deploy a TLS and password authentication secured Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).
For creating the TLS certificates, we run a Kubernetes job that creates a self-signed CA certificate and client certificates and keys for the postgres admin DB user and for the zitadel DB user.

```bash
# Generate TLS certificates
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/certs-job.yaml

# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait postgres bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/postgres-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/zitadel-values.yaml
```
