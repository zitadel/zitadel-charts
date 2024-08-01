# GCP Managed Postgres Example

By running the commands below, you deploy a ZITADEL instance configured to use a Google Cloud SQL managed Postgres database. The connection to the database is handled securely using the Cloud SQL Proxy.

> This example uses the Cloud SQL Proxy as a sidecar container to connect to the managed Postgres instance.

## Prerequisites

- A Google Cloud project with a Cloud SQL Postgres instance.
- A Kubernetes service account with the `Cloud SQL Client` role [Cloud SQL Client Role](https://cloud.google.com/sql/docs/mysql/connect-kubernetes-engine#workload-identity)

## Configuration

Update the `values.yaml` file with your Cloud SQL instance connection details e.g.:

```yaml
cloudSqlProxy:
  enabled: true
  imageTag: 2.12
  instanceConnectionName: <gcp-project-id>:<region>:<instance-name>
  args: ["--port=5432", "--structured-logs"]
  resources:
    requests:
      memory: "100Mi"
      cpu: "100m"
      ephemeral-storage: "612Mi"
```

Make sure that the key `serviceAccount` `name` is set to the service account with the `Cloud SQL Client` role. E.g.

```yaml
serviceAccount:
  create: false
  name: <service-account-name>
```	

