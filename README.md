[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# ZITADEL

## A Better Identity and Access Management Solution

Identity infrastructure, simplified for you.

Learn more about ZITADEL by checking out the [source repository on GitHub](https://github.com/zitadel/zitadel)

## What's in the Chart

By default, this chart installs a highly available ZITADEL deployment.

## Install the Chart

Either follow the [guide for deploying ZITADEL on Kubernetes](https://zitadel.com/docs/self-hosting/deploy/kubernetes) or follow one of the example guides:

- [Insecure Postgres Example](examples/1-postgres-insecure/README.md)
- [Secure Postgres Example](examples/2-postgres-secure/README.md)
- [Insecure Cockroach Example](examples/3-cockroach-insecure/README.md)
- [Secure Cockroach Example](examples/4-cockroach-secure/README.md)
- [Referenced Secrets Example](examples/5-referenced-secrets/README.md)
- [Machine User Setup Example](examples/6-machine-user/README.md)

## Upgrade from v7

The default ZITADEL version is now >= v2.55.
[This requires Cockroach DB to be at >= v23.2](https://zitadel.com/docs/support/advisory/a10009)
If you are using an older version of Cockroach DB, please upgrade it before upgrading ZITADEL.

Note that in order to upgrade cockroach, you should not jump minor versions.
For example:

```bash
# install Cockroach DB v23.1.14
helm upgrade db cockroachdb/cockroachdb --version 11.2.4 --reuse-values
# install Cockroach DB v23.2.5
helm upgrade db cockroachdb/cockroachdb --version 12.0.5 --reuse-values
# install Cockroach DB v24.1.1
helm upgrade db cockroachdb/cockroachdb --version 13.0.1 --reuse-values
# install ZITADEL v2.55.0
helm upgrade my-zitadel zitadel/zitadel --version 8.0.0 --reuse-values
```

Please refer to the docs by Cockroach Labs. The ZITADEL tests run against the [official CockroachDB chart](https://artifacthub.io/packages/helm/cockroachdb/cockroachdb).

(Credits to @panapol-p and @kleberbaum :pray:)

## Upgrade from v6

- Now, you have the flexibility to define resource requests and limits separately for the machineKeyWriter,
  distinct from the setupJob.
  If you don't specify resource requests and limits for the machineKeyWriter,
  it will automatically inherit the values used by the setupJob.

- To maintain consistency in the structure of the values.yaml file, certain properties have been renamed.
  If you are using any of the following properties, kindly review the updated names and adjust the values accordingly:

  | Old Value                                   | New Value                                    |
  |---------------------------------------------|----------------------------------------------|
  | `setupJob.machinekeyWriterImage.repository` | `setupJob.machinekeyWriter.image.repository` |
  | `setupJob.machinekeyWriterImage.tag`        | `setupJob.machinekeyWriter.image.tag`        |

## Upgrade from v5

- CockroachDB is not in the default configuration anymore.
  If you use CockroachDB, please check the host and ssl mode in your ZITADEL Database configuration section.

- The properties for database certificates are renamed and the defaults are removed.
  If you use one of the following properties, please check the new names and set the values accordingly:

  | Old Value                      | New Value                     |
  |--------------------------------|-------------------------------|
  | `zitadel.dbSslRootCrt`         | `zitadel.dbSslCaCrt`          | 
  | `zitadel.dbSslRootCrtSecret`   | `zitadel.dbSslCaCrtSecret`    |
  | `zitadel.dbSslClientCrtSecret` | `zitadel.dbSslAdminCrtSecret` |
  | `-`                            | `zitadel.dbSslUserCrtSecret`  |

## Uninstalling the Chart

The ZITADEL chart uses Helm hooks,
[which are not garbage collected by helm uninstall, yet](https://helm.sh/docs/topics/charts_hooks/#hook-resources-are-not-managed-with-corresponding-releases).
Therefore, to also remove hooks installed by the ZITADEL Helm chart,
delete them manually:

```bash
helm uninstall my-zitadel
for k8sresourcetype in job configmap secret rolebinding role serviceaccount; do
    kubectl delete $k8sresourcetype --selector app.kubernetes.io/name=zitadel,app.kubernetes.io/managed-by=Helm
done
```

## Troubleshooting

### Debug Pod

For troubleshooting, you can deploy a debug pod by setting the `zitadel.debug.enabled` property to `true`.
You can then use this pod to inspect the ZITADEL configuration and run zitadel commands using the zitadel binary.
For more information, print the debug pods logs using something like the following command:

```bash 
kubectl logs rs/my-zitadel-debug
``` 

### migration already started, will check again in 5 seconds

If you see this error message in the logs of the setup job, you need to reset the last migration step once you resolved the issue.
To do so, start a [debug pod](#debug-pod) and run something like the following command:

```bash
kubectl exec -it my-zitadel-debug -- zitadel setup cleanup --config /config/zitadel-config-yaml
```

## Contributing

Lint the chart:

```bash
docker run -it --network host --workdir=/data --rm --volume $(pwd):/data quay.io/helmpack/chart-testing:v3.5.0 ct lint --charts charts/zitadel --target-branch main
```

Test the chart:

```bash
# Create a local Kubernetes cluster
kind create cluster --image kindest/node:v1.27.2

# Test the chart
go test ./...
```

Watch the Kubernetes pods if you want to see progress.

```bash
kubectl get pods --all-namespaces --watch

# Or if you have the watch binary installed
watch -n .1 "kubectl get pods --all-namespaces"
```

## Contributors

<a href="https://github.com/zitadel/zitadel-charts/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=zitadel/zitadel-charts" />
</a>
