# Postgresql Subchart Example

By running the commands below, you deploy a TLS and password authentication secured Postgres database to your Kubernetes cluster [by using a correctly configured ZITADEL chart](https://artifacthub.io/packages/helm/zitadel/zitadel).
For creating the TLS certificates, we run a Kubernetes job that creates a self-signed CA certificate and client certificates and keys for the postgres admin DB user and for the zitadel DB user.

```bash
# Generate TLS certificates (optional)
kubectl apply -f https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/8-postgres-subchart/certs-job.yaml
kubectl wait --for=condition=complete job/create-certs
```
or use `zitadel.certs.job.enabled: true` in your `values.yaml` file

```bash
# Install ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values values.yaml # example: https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/8-postgres-subchart/values.yaml
```
once the installations is complete, you can make it accessible from outside kubernetes using and Ingress rule either in the helm chart, or by creating your own:

```yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: nameOfClusterIssuer #if using cert-manager
  name: zitadel-ingress
  namespace: zitadel
spec:
  rules:
  - host: zitadel.example.com # replace this with whatever hostname you want it to be exposed as
    http:
      paths:
      - pathType: Prefix
        path: /
        backend:
          service:
            name: zitadel-server # ensure this is the same service name created by the helmchart that has the zitadel server as the endpoint
            port:
              number: 80 # ensure the port number here matches the port that the service exposes
  tls:
  - hosts:
    - zitadel.example.com # again, make sure that the hostname here matches the host name above
    secretName: myingress-cer
---
```


