[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# Zitadel

## A Better Identity and Access Management Solution

Identity infrastructure, simplified for you.

Learn more about Zitadel by checking out the [source repository on GitHub](https://github.com/zitadel/zitadel)

## What's in the Chart

By default, this chart installs a highly available Zitadel deployment.

The chart deploys a Zitadel init job, a Zitadel setup job and a Zitadel deployment.
By default, the execution order is orchestrated using Helm hooks on installations and upgrades.

## Install the Chart

The easiest way to deploy a Helm release for Zitadel is by following the [Insecure Postgres Example](/examples/1-postgres-insecure/README.md).
For more sofisticated production-ready configurations, follow one of the following examples:

- [Secure Postgres Example](/examples/2-postgres-secure/README.md)
- [Referenced Secrets Example](/examples/3-referenced-secrets/README.md)
- [Machine User Setup Example](/examples/4-machine-user/README.md)
- [Internal TLS Example](/examples/5-internal-tls/README.md)

All the configurations from the examples above are guaranteed to work, because they are directly used in automatic acceptance tests.

## Upgrade From V8 to V9

The v9 charts default Zitadel and login versions reference [Zitadel v4](https://github.com/zitadel/zitadel/releases/tag/v4.0.0).

### Don’t Switch to the New Login Deployment

Use `login.enabled: false` to omit deploying the new login.

### Switch to the New Login Deployment

By default, a new deployment for the login v2 is configured and created.
For new installations, the setup job automatically creates a user of type machine with role `IAM_LOGIN_CLIENT`.
It writes the users personal access token into a Kubernetes secret which is then mounted into the login pods.

For existing installations, the setup job doesn't create this login client user.
Therefore, the Kubernetes secret has to be created manually before upgrading to v9:

1. Create a user of type machine
2. Make the user an instance administrator with role `IAM_LOGIN_CLIENT`
3. Create a personal access token for the user 
4. Create a secret with that token: `kubectl --namespace <my-namespace> create secret generic login-client --from-file=pat=<my-local-path-to-the-downloaded-pat-file>`

To make the login externally accessible, you need to route traffic with the path prefix `/ui/v2/login` to the login service.
If you use an ingress controller, you can enable the login ingress with `login.ingress.enabled: true`

> [!CAUTION]
> Don't Lock Yourself Out of Your Instance  
> Before you change your Zitadel configuration, we highly recommend you to create a service user with a personal access token (PAT) and the IAM_OWNER role.  
> In case something breaks, you can use this PAT to revert your changes or fix the configuration so you can use a login UI again.  

To actually use the new login, enable the loginV2 feature on the instance.
Leave the base URI empty to use the default or explicitly configure it to `/ui/v2/login`.
If you enable this feature, the login will be used for every application configured in your Zitadel instance.

### Other Breaking Changes

- Default Traefik and NGINX annotations for internal unencrypted HTTP/2 traffic to the Zitadel pods are added.
- The default value `localhost` is removed from the Zitadel ingresses `host` field. Instead, the `host` fields for the Zitadel and login ingresses default to `zitadel.configmapConfig.ExternalDomain`.
- The following Kubernetes versions are tested:
  - v1.33.1
  - v1.32.5
  - v1.31.9
  - v1.30.13

## Upgrade from v7

> [!WARNING] The chart version 8 doesn't get updates to the default Zitadel version anymore as this might break environments that use CockroachDB.
> Please set the version explicitly using the appVersion variable if you need a newer Zitadel version.
> The upcoming version 9 will include the latest Zitadel version by default (Zitadel v3).

The default Zitadel version is now >= v2.55.
[This requires Cockroach DB to be at >= v23.2](https://zitadel.com/docs/support/advisory/a10009)
If you are using an older version of Cockroach DB, please upgrade it before upgrading Zitadel.

Note that in order to upgrade cockroach, you should not jump minor versions.
For example:

```bash
# install Cockroach DB v23.1.14
helm upgrade db cockroachdb/cockroachdb --version 11.2.4 --reuse-values
# install Cockroach DB v23.2.5
helm upgrade db cockroachdb/cockroachdb --version 12.0.5 --reuse-values
# install Cockroach DB v24.1.1
helm upgrade db cockroachdb/cockroachdb --version 13.0.1 --reuse-values
# install Zitadel v2.55.0
helm upgrade my-zitadel zitadel/zitadel --version 8.0.0 --reuse-values
```

Please refer to the docs by Cockroach Labs. The Zitadel tests run against the [official CockroachDB chart](https://artifacthub.io/packages/helm/cockroachdb/cockroachdb).

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
  If you use CockroachDB, please check the host and ssl mode in your Zitadel Database configuration section.

- The properties for database certificates are renamed and the defaults are removed.
  If you use one of the following properties, please check the new names and set the values accordingly:

  | Old Value                      | New Value                     |
  |--------------------------------|-------------------------------|
  | `zitadel.dbSslRootCrt`         | `zitadel.dbSslCaCrt`          | 
  | `zitadel.dbSslRootCrtSecret`   | `zitadel.dbSslCaCrtSecret`    |
  | `zitadel.dbSslClientCrtSecret` | `zitadel.dbSslAdminCrtSecret` |
  | `-`                            | `zitadel.dbSslUserCrtSecret`  |

## Uninstalling the Chart

The Zitadel chart uses Helm hooks,
[which are not garbage collected by helm uninstall, yet](https://helm.sh/docs/topics/charts_hooks/#hook-resources-are-not-managed-with-corresponding-releases).
Therefore, to also remove hooks installed by the Zitadel Helm chart,
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
You can then use this pod to inspect the Zitadel configuration and run zitadel commands using the zitadel binary.
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

### Multiple Releases in Single Namespace

Read the comment for the value login.loginClientSecretPrefix

## Contributing

### Editor Configuration

This repository uses an `.editorconfig` file to maintain consistent coding styles across different editors and IDEs.
Make sure your IDE supports the `/.editorconfig` file to automatically apply the correct formatting rules.

Many IDEs support EditorConfig out of the box (including IntelliJ IDEA, PyCharm, WebStorm, and others), while [some require you to install a plugin](https://editorconfig.org/#editor-plugins), like VSCode.

The `.editorconfig` file in this repository defines:
- Line endings (LF)
- Character encoding (UTF-8)
- Indentation style (spaces for most files, tabs for Go)
- Trailing whitespace handling
- Final newline requirements

If your editor doesn't automatically pick up these settings, please install the appropriate EditorConfig plugin for your development environment.

#### Lint the chart:

```bash
docker run -it --network host --workdir=/data --rm --volume $(pwd):/data quay.io/helmpack/chart-testing:v3.5.0 ct lint --charts charts/zitadel --target-branch main
```

#### Validate Helm Charts

Validate the Helm chart manifests against Kubernetes API schemas using kubeconform:

```bash
make check
```

This renders the chart templates and validates them for correctness without requiring a Kubernetes cluster.

### Test the chart:

```bash
# Create KinD cluster
kind create cluster --config ./charts/zitadel/acceptance_test/kindConfig.yaml

# Test the chart
make test
```

Watch the Kubernetes pods if you want to see progress.

```bash
kubectl get pods --all-namespaces --watch

# Or if you have the watch binary installed
watch -n .1 "kubectl get pods --all-namespaces"
```

## Contributors

<a href="https://github.com/zitadel/zitadel-charts/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=zitadel/zitadel-charts"  alt="Contributors"/>
</a>
