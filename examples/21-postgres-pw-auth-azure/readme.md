This example uses [`helmfile`](https://helmfile.readthedocs.io/) with [`helm-secrets`](https://github.com/jkroepke/helm-secrets) in combination with our [Kubernetes Deploy Guides](https://zitadel.com/docs/self-hosting/deploy/kubernetes) and the [PostgreSQL Database](https://zitadel.com/docs/self-hosting/manage/database#postgres) to provision a production-eligible ZITADEL instance. You can make this example work without `helmfile` by mirroring the `zitadel` configration to your declarative setup.

The secrets are safely retrieved from Azure KeyVault in this example but you can also use other secret stores supported by `helm-secrets` or provide the secrets directly within the `values`.

> **Warning**
> Even though this examples strives to be complete, make sure to read our [Production Guide](https://zitadel.com/docs/self-hosting/manage/production) before you decide to use it as reference.

You can provision the example via a `helmfile` command like this:

```bash
helmfile -f zitadel.yaml diff
helmfile -f zitadel.yaml apply
helmfile -f zitadel.yaml destroy
```