[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# ZITADEL

## A Better Identity and Access Management Solution

Identity infrastructure, simplified for you.
It is built for the serverless era.

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
