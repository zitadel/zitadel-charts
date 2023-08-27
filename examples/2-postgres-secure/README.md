# Secure Postgres Example

By running the commands below, you deploy a TLS and password authentication secured Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).
For creating the TLS certificates, we run a Kubernetes job that creates a self-signed CA certificate and client certificates and keys for the postgres admin DB user and for the zitadel DB user.

```bash
# Generate TLS certificates
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/certs-job.yaml
kubectl wait --for=condition=complete job/create-certs

# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait postgres bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/postgres-values.yaml

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/2-postgres-secure/zitadel-values.yaml
```

When ZITADEL is ready, you can access the GUI via port-forwarding:

```bash
kubectl port-forward svc/my-zitadel 8080
```

Now, open http://zitadel.default.127.0.0.1.sslip.io:8080 in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.zitadel.default.127.0.0.1.sslip.io  
**Password**: Password1!
