# Admin API Access Example

For accessing the ZITADEL API immediately after installation, the chart provisions a declarative **admin-client** system user (role `IAM_OWNER`).
The chart generates an X.509 keypair, stores it in a `kubernetes.io/tls` Secret named `<release>-admin-service-key` (e.g. `my-zitadel-admin-service-key`), and configures ZITADEL to trust the public certificate.
You read the **private key** out of that Secret and authenticate against the ZITADEL API with a JWT profile — no Secret is created imperatively at runtime, so this works cleanly with GitOps tooling.
With this, you can automate the setup of ZITADEL along with ZITADEL resources like projects, users, and more from scratch.
To achieve a fully declared ZITADEL setup, also check out the [ZITADEL Terraform provider](https://registry.terraform.io/providers/zitadel/zitadel/latest).

> [!NOTE]
> Earlier chart versions (≤ v10) created an `iam-admin` machine-key Secret imperatively. That mechanism is removed in v11; existing `iam-admin` secrets keep working after an upgrade. See the chart's "Upgrade From V10 to V11" notes.

By running the commands below, you deploy a simple, insecure Postgres database to your Kubernetes cluster [by using the Bitnami chart](https://artifacthub.io/packages/helm/bitnami/postgresql).
Also, you deploy [a correctly configured ZITADEL](https://artifacthub.io/packages/helm/zitadel/zitadel).

> [!WARNING]
> Anybody with network access to the Postgres database can connect to it and read and write data.
> Use this example only for testing purposes.
> For deploying a secure Postgres database, see [the secure Postgres example](../2-postgres-secure/README.md).

> [!INFO]
> The example assumes you already have a running Kubernetes cluster with a working ingress controller.
> If you don't, [run a local KinD cluster](../99-kind-with-traefik/README.md) before executing the following commands.

```bash
# Install Postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install --wait db bitnami/postgresql --version 12.10.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-machine-user/postgres-values.yaml

# Install Zitadel
helm repo add zitadel https://charts.zitadel.com
helm install my-zitadel zitadel/zitadel --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/4-machine-user/zitadel-values.yaml
```

When Zitadel is ready, open https://machine.127.0.0.1.sslip.io/ui/console?login_hint=zitadel-admin@zitadel.machine.127.0.0.1.sslip.io in your browser and log in with the password `Password1!`.

To get the admin-client private key for API automation:

```bash
kubectl get secret my-zitadel-admin-service-key -o jsonpath='{.data.tls\.key}' | base64 -d > admin-client.key
```

Sign a JWT with `iss` and `sub` set to `admin-client` and `aud` set to your external ZITADEL URL, then send it as a `Bearer` token to the ZITADEL API.
