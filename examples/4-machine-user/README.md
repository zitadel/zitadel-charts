# Machine User Setup Example

Instead of having ZITADEL set up with a human user with the `IAM_ADMIN` role, you can also set up ZITADEL with a machine user.
For accessing the ZITADEL API immediately after the installation, you can have ZITADEL either create a personal access token or a machine key of type JSON.
The chart creates a Kubernetes secret with the machine user credentials on the first installation of the chart for further use.
With this, you can automate the setup of ZITADEL along with ZITADEL resources like projects, users, and more from scratch.
To achieve a fully declared ZITADEL setup, also check out the [ZITADEL Terraform provider](https://registry.terraform.io/providers/zitadel/zitadel/latest).

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
