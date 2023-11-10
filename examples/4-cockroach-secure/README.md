# Secure Cockroach Example

By running the commands below, you deploy a TLS and password authentication secured Cockroach database to your Kubernetes cluster [by using the Cockroach official chart](https://artifacthub.io/packages/helm/cockroachdb/cockroachdb).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).
The cockroach chart is configured to create a CA certificate and a client certificate and key for the root DB user.
However, you also want to secure the runtime connections between ZITADEL and Cockroach for the zitadel user.
Therefore, you create a Kubernetes job that creates a client certificate and key for the zitadel DB user by executing the `cockroach cert` command.

```bash
# Install Cockroach
helm repo add cockroachdb https://charts.cockroachdb.com/
helm install db cockroachdb/cockroachdb --version 11.2.1 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-cockroach-secure/cockroach-values.yaml

# Generate a TLS certificate for the zitadel DB user
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-cockroach-secure/zitadel-cert-job.yaml
kubectl wait --for=condition=complete job/create-zitadel-cert

# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-cockroach-secure/zitadel-values.yaml
```

When ZITADEL is ready, you can access the GUI via port-forwarding:

```bash
kubectl port-forward svc/my-zitadel 8080
```

Now, open http://127.0.0.1.sslip.io:8080 in your browser and log in with the following credentials:

**Username**: zitadel-admin@zitadel.127.0.0.1.sslip.io
**Password**: Password1!
