# Secure Cockroach Example

By running the commands below, you deploy a TLS and password authentication secured Cockroach database to your Kubernetes cluster [by using the Cockroach official chart](https://artifacthub.io/packages/helm/cockroachdb/cockroachdb).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).
The cockroach chart is configured to create a CA certificate and a client certificate and key for the root DB user.
However, you also want to secure the runtime connections between ZITADEL and Cockroach for the zitadel user.
Therefore, you create a Kubernetes job that creates a client certificate and key for the zitadel DB user by executing the `cockroach cert` command.

```bash
# Install Cockroach
helm repo add cockroachdb https://charts.cockroachdb.com/
helm install cockroach cockroachdb/cockroachdb --version 11.1.5

# Wait for the cockroach CA certificate to be created
until kubectl -n zitadel-test-4-cockroach-secure-dq8tfz get secret db-cockroachdb-ca-secret; do echo "awaiting cockroach ca"; sleep 1; done

# Generate a TLS certificate for the zitadel DB user
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-cockroach-secure/zitadel-cert-job.yaml

# Install ZITADEL
helm repo add zitadel https://zitadel.github.io/zitadel/charts
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-cockroach-secure/zitadel-values.yaml
```
