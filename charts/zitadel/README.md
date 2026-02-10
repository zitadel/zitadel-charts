[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# Zitadel

![Version: 9.19.0](https://img.shields.io/badge/Version-9.19.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v4.2.0](https://img.shields.io/badge/AppVersion-v4.2.0-informational?style=flat-square)

## A Better Identity and Access Management Solution

Identity infrastructure, simplified for you.

Learn more about Zitadel by checking out the [source repository on GitHub](https://github.com/zitadel/zitadel)

## What's in the Chart

By default, this chart installs a highly available Zitadel deployment.

The chart deploys a Zitadel init job, a Zitadel setup job and a Zitadel deployment.
By default, the execution order is orchestrated using Helm hooks on installations and upgrades.

## Install the Chart

The easiest way to deploy a Helm release for Zitadel is by following the [Insecure Postgres Example](/examples/1-postgres-insecure/README.md).
For more sophisticated production-ready configurations, follow one of the following examples:

- [Secure Postgres Example](/examples/2-postgres-secure/README.md)
- [Referenced Secrets Example](/examples/3-referenced-secrets/README.md)
- [Machine User Setup Example](/examples/4-machine-user/README.md)
- [Internal TLS Example](/examples/5-internal-tls/README.md)

All the configurations from the examples above are guaranteed to work, because they are directly used in automatic acceptance tests.

## Upgrade From V8 to V9

The v9 charts default Zitadel and login versions reference [Zitadel v4](https://github.com/zitadel/zitadel/releases/tag/v4.0.0).

### Don't Switch to the New Login Deployment

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
> In case something breaks, you can use this PAT to revert your changes or fix the configuration so you can log in to the UI again.

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

## Requirements

Kubernetes: `>= 1.30.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling. Use for advanced pod placement strategies like co-locating pods on the same node (pod affinity), spreading pods across zones (pod anti-affinity), or preferring certain nodes (node affinity). Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| annotations | object | `{}` | Annotations to add to the ZITADEL Deployment resource. Use this for integration with tools like ArgoCD, Flux, or external monitoring systems. |
| cleanupJob | object | `{"activeDeadlineSeconds":60,"annotations":{"helm.sh/hook":"post-delete","helm.sh/hook-delete-policy":"hook-succeeded","helm.sh/hook-weight":"-1"},"backoffLimit":3,"enabled":true,"podAdditionalLabels":{},"podAnnotations":{},"resources":{}}` | Cleanup Job configuration for removing imperatively created resources on helm uninstall. The setup job creates secrets using kubectl that aren't tracked by Helm's lifecycle, causing them to persist after chart removal. This job runs as a post-delete hook to clean up those orphaned resources. |
| cleanupJob.activeDeadlineSeconds | int | `60` | Maximum time in seconds for the cleanup job to complete. After this deadline, the job is terminated even if still running. |
| cleanupJob.annotations | object | `{"helm.sh/hook":"post-delete","helm.sh/hook-delete-policy":"hook-succeeded","helm.sh/hook-weight":"-1"}` | Annotations for the cleanup job. The post-delete hook ensures this runs on helm uninstall, and the delete policy removes the job after completion. |
| cleanupJob.backoffLimit | int | `3` | Number of retries before marking the cleanup job as failed. |
| cleanupJob.enabled | bool | `true` | Enable the cleanup job to remove secrets created by the setup job. Set to false if you want to preserve secrets across reinstalls. |
| cleanupJob.podAdditionalLabels | object | `{}` | Additional labels to add to cleanup job pods. |
| cleanupJob.podAnnotations | object | `{}` | Additional annotations to add to cleanup job pods. |
| cleanupJob.resources | object | `{}` | Resource limits and requests for the cleanup job container. Keep minimal as this job only runs kubectl delete commands. |
| configMap | object | `{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}}` | ConfigMap configuration for ZITADEL's runtime configuration. |
| configMap.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the ZITADEL ConfigMap. The default Helm hooks ensure the ConfigMap is created before the deployment and recreated on upgrades to pick up configuration changes. |
| env | list | `[]` | Additional environment variables for the ZITADEL container. Use this to pass configuration that isn't available through configmapConfig or secretConfig, or to inject values from other Kubernetes resources like ConfigMaps or Secrets. ZITADEL environment variables follow the pattern ZITADEL_<SECTION>_<KEY>. Ref: https://zitadel.com/docs/self-hosting/manage/configure#configure-by-environment-variables |
| envVarsSecret | string | `""` | Name of a Kubernetes Secret containing environment variables to inject into the ZITADEL container. All key-value pairs in the secret will be available as environment variables. This is useful for managing multiple ZITADEL configuration values in a single secret, especially when using external secret management tools like External Secrets Operator or Sealed Secrets. Ref: https://zitadel.com/docs/self-hosting/manage/configure#configure-by-environment-variables |
| extraContainers | list | `[]` | Sidecar containers to run alongside the main ZITADEL container in the Deployment pod. Use this for logging agents, monitoring sidecars, service meshes, or database proxies (e.g., cloud-sql-proxy for Google Cloud SQL). These containers share the pod's network namespace and can access the same volumes as the main container. |
| extraManifests | list | `[]` | Additional Kubernetes manifests to deploy alongside the chart. This allows you to include custom resources without creating a separate chart. Supports Helm templating syntax including .Release, .Values, and template functions. Use this for secrets, configmaps, network policies, or any other resources that ZITADEL depends on. |
| extraVolumeMounts | list | `[]` | Additional volume mounts for the main ZITADEL container. Use this to mount volumes defined in extraVolumes into the container filesystem. Common use cases include mounting custom CA certificates, configuration files, or shared data between containers. |
| extraVolumes | list | `[]` | Additional volumes to add to ZITADEL pods. These volumes can be referenced by extraVolumeMounts to make data available to the ZITADEL container or sidecar containers. Supports all Kubernetes volume types: secrets, configMaps, persistentVolumeClaims, emptyDir, hostPath, etc. |
| fullnameOverride | string | `""` | Completely override the generated resource names (release-name + chart-name). Takes precedence over nameOverride. Use this when you need full control over resource naming, such as when migrating from another chart. |
| image | object | `{"pullPolicy":"IfNotPresent","repository":"ghcr.io/zitadel/zitadel","tag":""}` | Container image configuration for the main ZITADEL application. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. Use "Always" for mutable tags like "latest", or "IfNotPresent" for immutable version tags to reduce network traffic. |
| image.repository | string | `"ghcr.io/zitadel/zitadel"` | Docker image repository for ZITADEL. The default uses GitHub Container Registry. Change this if using a private registry or mirror. |
| image.tag | string | `""` | Image tag. Defaults to the chart's appVersion if not specified. Use a specific version tag (e.g., "v2.45.0") for production deployments to ensure reproducibility and controlled upgrades. |
| imagePullSecrets | list | `[]` |  |
| imageRegistry | string | `""` | Global container registry override for tool images (e.g., wait4x, kubectl). When set, this registry is prepended to tool image repositories for compatibility with CRI-O v1.34+ which enforces fully qualified image names. If left empty, defaults to "docker.io". |
| ingress | object | `{"annotations":{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"},"className":"","controller":"generic","enabled":false,"hosts":[{"paths":[{"path":"/","pathType":"Prefix"}]}],"tls":[]}` | Ingress configuration for exposing the main ZITADEL service. This allows external traffic to reach the ZITADEL API and gRPC endpoints. It can be configured to work with various ingress controllers like NGINX, Traefik, or AWS ALB. |
| ingress.annotations | object | `{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"}` | A map of annotations to apply to the Ingress resource. The default annotation is for NGINX to correctly handle gRPC traffic. |
| ingress.className | string | `""` | The name of the IngressClass resource to use for this Ingress. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
| ingress.controller | string | `"generic"` | A chart-specific setting to enable logic for different controllers. Use "aws" to generate AWS ALB-specific annotations and resources. |
| ingress.enabled | bool | `false` | If true, creates an Ingress resource for the ZITADEL service. |
| ingress.hosts | list | `[{"paths":[{"path":"/","pathType":"Prefix"}]}]` | A list of host rules for the Ingress. Each host can have multiple paths. |
| ingress.tls | list | `[]` | TLS configuration for the Ingress. This allows you to secure the endpoint with HTTPS by referencing a secret that contains the TLS certificate and key. |
| initJob | object | `{"activeDeadlineSeconds":300,"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"1"},"backoffLimit":5,"command":"","enabled":true,"extraContainers":[],"initContainers":[],"podAdditionalLabels":{},"podAnnotations":{},"resources":{}}` | Init Job configuration for database initialization. This Kubernetes Job runs as a Helm pre-install/pre-upgrade hook to prepare the database before ZITADEL starts. It creates the database schema, user, and grants necessary permissions. |
| initJob.activeDeadlineSeconds | int | `300` | Maximum time in seconds for the init job to complete. The job is terminated if it exceeds this deadline, regardless of backoffLimit. |
| initJob.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"1"}` | Annotations for the init job. The Helm hooks ensure this job runs before the main deployment and is recreated on each upgrade. |
| initJob.backoffLimit | int | `5` | Number of retries before marking the init job as failed. Increase this if the database might take longer to become available. |
| initJob.enabled | bool | `true` | Enable or disable the init job. Set to false after the initial installation if you want to manage database initialization externally or if no database changes are expected during upgrades. |
| initJob.extraContainers | list | `[]` | Sidecar containers to run alongside the init container. Useful for logging, proxies (e.g., cloud-sql-proxy), or other supporting services. |
| initJob.initContainers | list | `[]` | Init containers to run before the main init container. Useful for waiting on additional dependencies or performing pre-initialization tasks. |
| initJob.podAdditionalLabels | object | `{}` | Additional labels to add to init job pods. |
| initJob.podAnnotations | object | `{}` | Additional annotations to add to init job pods. |
| initJob.resources | object | `{}` | CPU and memory resource requests and limits for the init job container. The init job typically requires minimal resources as it only runs SQL commands against the database. |
| livenessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5}` | Liveness probe configuration for ZITADEL. The liveness probe determines if a container is still running properly. Failed probes cause the container to be restarted, which can help recover from deadlocks or stuck states. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/ |
| livenessProbe.enabled | bool | `true` | Enable or disable the liveness probe. |
| livenessProbe.failureThreshold | int | `3` | Number of consecutive failures before restarting the container. |
| livenessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting liveness checks after container start. |
| livenessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the liveness check. |
| login | object | `{"affinity":{},"annotations":{},"autoscaling":{"annotations":{},"behavior":{},"enabled":false,"maxReplicas":10,"metrics":[],"minReplicas":3,"targetCPU":null,"targetMemory":null},"configMap":{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}},"customConfigmapConfig":null,"enabled":true,"env":[],"extraContainers":[],"extraVolumeMounts":[],"extraVolumes":[],"fullnameOverride":"","image":{"pullPolicy":"IfNotPresent","repository":"ghcr.io/zitadel/zitadel-login","tag":""},"imagePullSecrets":[],"ingress":{"annotations":{},"className":"","controller":"generic","enabled":false,"hosts":[{"paths":[{"path":"/ui/v2/login","pathType":"Prefix"}]}],"tls":[]},"initContainers":[],"livenessProbe":{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5},"loginClientSecretPrefix":null,"nameOverride":"","nodeSelector":{},"pdb":{"annotations":{},"enabled":false,"maxUnavailable":null,"minAvailable":null},"podAdditionalLabels":{},"podAnnotations":{},"podSecurityContext":{},"readinessProbe":{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5},"replicaCount":1,"resources":{},"revisionHistoryLimit":10,"securityContext":{},"service":{"annotations":{},"appProtocol":"kubernetes.io/http","clusterIP":"","externalTrafficPolicy":"","labels":{},"port":3000,"protocol":"http","scheme":"HTTP","type":"ClusterIP"},"serviceAccount":{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"},"create":true,"name":""},"startupProbe":{"enabled":false,"failureThreshold":30,"periodSeconds":1},"tolerations":[],"topologySpreadConstraints":[]}` | Configuration for the ZITADEL Login UI, a separate Next.js application that provides the user-facing authentication interface. When enabled, it deploys alongside the main ZITADEL service and handles login, registration, and password reset flows. |
| login.affinity | object | `{}` | Affinity rules for pod scheduling. Use for advanced pod placement strategies like co-locating pods or spreading across zones. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| login.annotations | object | `{}` | Annotations to add to the Login UI Deployment resource. |
| login.autoscaling | object | `{"annotations":{},"behavior":{},"enabled":false,"maxReplicas":10,"metrics":[],"minReplicas":3,"targetCPU":null,"targetMemory":null}` | Horizontal Pod Autoscaler configuration for scaling on CPU, memory, or custom metrics. Ref: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/ |
| login.autoscaling.annotations | object | `{}` | Optional map of annotations applied to the HPA object. |
| login.autoscaling.behavior | object | `{}` | Configures the scaling behavior for scaling up and down. Use this to control how quickly the HPA scales pods in response to metric changes. Ref: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior |
| login.autoscaling.enabled | bool | `false` | If true, enables the Horizontal Pod Autoscaler for the login deployment. This will automatically override the `replicaCount` value. |
| login.autoscaling.maxReplicas | int | `10` | The maximum number of pod replicas. |
| login.autoscaling.metrics | list | `[]` | Advanced scaling based on custom metrics exposed by Zitadel. To use these for scaling, you MUST have a metrics server (e.g., Prometheus) and a metrics adapter (e.g., prometheus-adapter) running in your cluster. Ref: https://github.com/kubernetes-sigs/prometheus-adapter |
| login.autoscaling.minReplicas | int | `3` | The minimum number of pod replicas. |
| login.autoscaling.targetCPU | string | `nil` | The target average CPU utilization percentage. |
| login.autoscaling.targetMemory | string | `nil` | The target average memory utilization percentage. |
| login.configMap | object | `{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}}` | ConfigMap configuration for the Login UI environment variables. |
| login.configMap.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the Login UI ConfigMap. The default hooks ensure the ConfigMap is created before the deployment and recreated on upgrades. |
| login.customConfigmapConfig | string | `nil` | Custom environment variables for the Login UI ConfigMap. These override the default values which configure the service user token path, API URL, and custom request headers. Only set this if you need to customize the Login UI behavior beyond the defaults. The defaults are:   ZITADEL_SERVICE_USER_TOKEN_FILE="/login-client/pat"   ZITADEL_API_URL="http://<release>-zitadel:<port>"   CUSTOM_REQUEST_HEADERS="Host:<ExternalDomain>" |
| login.enabled | bool | `true` | Enable or disable the Login UI deployment. When disabled, ZITADEL uses its built-in login interface instead of the separate Login UI application. |
| login.env | list | `[]` | Additional environment variables for the Login UI container. Use this to pass configuration that isn't available through customConfigmapConfig. |
| login.extraContainers | list | `[]` | Sidecar containers to run alongside the Login UI container. Useful for logging agents, proxies, or other supporting services. |
| login.extraVolumeMounts | list | `[]` | Additional volume mounts for the Login UI container. Use this to mount custom certificates, configuration files, or other data into the container. |
| login.extraVolumes | list | `[]` | Additional volumes for the Login UI pod. Define volumes here that are referenced by extraVolumeMounts. |
| login.fullnameOverride | string | `""` | Completely override the generated resource names. Takes precedence over nameOverride when set. |
| login.image | object | `{"pullPolicy":"IfNotPresent","repository":"ghcr.io/zitadel/zitadel-login","tag":""}` | Container image configuration for the Login UI. |
| login.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. Use "Always" for mutable tags like "latest", or "IfNotPresent" for immutable tags. |
| login.image.repository | string | `"ghcr.io/zitadel/zitadel-login"` | Docker image repository for the Login UI. |
| login.image.tag | string | `""` | Image tag. Defaults to the chart's appVersion if not specified. Use a specific version tag for production deployments to ensure reproducibility. |
| login.imagePullSecrets | list | `[]` | References to secrets containing Docker registry credentials for pulling private images. Each entry should be the name of an existing secret. |
| login.ingress | object | `{"annotations":{},"className":"","controller":"generic","enabled":false,"hosts":[{"paths":[{"path":"/ui/v2/login","pathType":"Prefix"}]}],"tls":[]}` | Ingress configuration for exposing the ZITADEL Login UI service. This makes the web-based login interface accessible from outside the cluster. It can be configured independently of the main ZITADEL service Ingress. |
| login.ingress.annotations | object | `{}` | A map of annotations to apply to the Login UI Ingress resource. |
| login.ingress.className | string | `""` | The name of the IngressClass resource to use for this Ingress. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
| login.ingress.controller | string | `"generic"` | A chart-specific setting to enable logic for different controllers. Use "aws" to generate AWS ALB-specific annotations. |
| login.ingress.enabled | bool | `false` | If true, creates an Ingress resource for the Login UI service. |
| login.ingress.hosts | list | `[{"paths":[{"path":"/ui/v2/login","pathType":"Prefix"}]}]` | A list of host rules for the Ingress. The default path targets the login UI. |
| login.ingress.tls | list | `[]` | TLS configuration for the Ingress. Secure the login UI with HTTPS by referencing a secret containing the TLS certificate and key. |
| login.initContainers | list | `[]` | Init containers to run before the Login UI container starts. Useful for waiting on dependencies or performing setup tasks. |
| login.livenessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5}` | Liveness probe configuration. The liveness probe determines if a container is still running. Failed probes cause the container to be restarted. |
| login.livenessProbe.enabled | bool | `true` | Enable or disable the liveness probe. |
| login.livenessProbe.failureThreshold | int | `3` | Number of consecutive failures before restarting the container. |
| login.livenessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting liveness checks after container start. |
| login.livenessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the liveness check. |
| login.loginClientSecretPrefix | string | `nil` | Prefix for the login client secret name. Use this when deploying multiple ZITADEL instances in the same namespace to avoid secret name collisions. When set, the login client secret will be named "{prefix}login-client". |
| login.nameOverride | string | `""` | Override the "login" portion of resource names. Useful when the default naming conflicts with existing resources. |
| login.nodeSelector | object | `{}` | Node labels for pod assignment. Pods will only be scheduled on nodes with matching labels. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ |
| login.pdb | object | `{"annotations":{},"enabled":false,"maxUnavailable":null,"minAvailable":null}` | Pod Disruption Budget configuration for the Login UI deployment. Ensures high availability by limiting the number of pods that can be simultaneously unavailable during voluntary disruptions (e.g., node drains, rolling updates). Either minAvailable or maxUnavailable can be set, but not both. Values can be an integer (e.g., 1) or a percentage (e.g., "50%"). |
| login.pdb.annotations | object | `{}` | Additional annotations to apply to the Pod Disruption Budget resource |
| login.pdb.enabled | bool | `false` | Enable or disable the Pod Disruption Budget for Login UI pods. |
| login.pdb.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailable during disruptions. Cannot be used together with minAvailable. |
| login.pdb.minAvailable | string | `nil` | Minimum number of pods that must remain available during disruptions. Cannot be used together with maxUnavailable. |
| login.podAdditionalLabels | object | `{}` | Additional labels to add to Login UI pods beyond the standard Helm labels. Useful for organizing pods with custom label selectors. |
| login.podAnnotations | object | `{}` | Annotations to add to Login UI pods. Useful for integrations like Prometheus scraping, Istio sidecar injection, or Vault agent injection. |
| login.podSecurityContext | object | `{}` | Optional pod-level security context overrides for Login UI pods. If left empty, the chart-wide podSecurityContext defined below is used instead. Use this to customize security settings specifically for the Login UI. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| login.readinessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5}` | Readiness probe configuration. The readiness probe determines when a pod is ready to receive traffic. Failed probes remove the pod from service endpoints. |
| login.readinessProbe.enabled | bool | `true` | Enable or disable the readiness probe. |
| login.readinessProbe.failureThreshold | int | `3` | Number of consecutive failures before marking the pod as not ready. |
| login.readinessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting readiness checks after container start. |
| login.readinessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the readiness check. |
| login.replicaCount | int | `1` | Number of Login UI pod replicas. A single replica is fine for testing, but production environments should use 3 or more for high availability during rolling updates and node failures. |
| login.resources | object | `{}` | CPU and memory resource requests and limits for the Login UI container. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| login.revisionHistoryLimit | int | `10` | Number of old ReplicaSets to retain for rollback purposes. Set to 0 to disable rollback capability and save cluster resources. |
| login.securityContext | object | `{}` | Optional container-level security context overrides for the Login UI container. If left empty, the chart-wide securityContext defined below is used instead. Use this to customize security settings specifically for the Login UI container. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| login.service | object | `{"annotations":{},"appProtocol":"kubernetes.io/http","clusterIP":"","externalTrafficPolicy":"","labels":{},"port":3000,"protocol":"http","scheme":"HTTP","type":"ClusterIP"}` | Kubernetes Service configuration for the Login UI. |
| login.service.annotations | object | `{}` | Annotations to add to the Service resource. |
| login.service.appProtocol | string | `"kubernetes.io/http"` | Application protocol hint for ingress controllers and service meshes. Helps with protocol detection and routing decisions. |
| login.service.clusterIP | string | `""` | Fixed cluster IP address for ClusterIP services. Leave empty for automatic assignment. Only applicable when type is "ClusterIP". |
| login.service.externalTrafficPolicy | string | `""` | Traffic policy for LoadBalancer services. "Cluster" distributes traffic to all nodes, "Local" only routes to nodes with pods. Only applicable when type is "LoadBalancer". |
| login.service.labels | object | `{}` | Labels to add to the Service resource. |
| login.service.port | int | `3000` | Port number the service exposes. Clients connect to this port. |
| login.service.protocol | string | `"http"` | Protocol identifier used in port naming (e.g., "http", "https", "grpc"). |
| login.service.scheme | string | `"HTTP"` | HTTP scheme for health checks and internal communication. |
| login.service.type | string | `"ClusterIP"` | Service type. Use "ClusterIP" for internal access, "NodePort" for node-level access, or "LoadBalancer" for cloud provider load balancers. |
| login.serviceAccount | object | `{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"},"create":true,"name":""}` | ServiceAccount configuration for Login UI pods. |
| login.serviceAccount.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the Login UI service account. The default Helm hooks ensure it exists before pods are created. Add annotations here for cloud provider integrations (e.g., AWS IAM roles, GCP Workload Identity). |
| login.serviceAccount.create | bool | `true` | Whether to create a dedicated service account for the Login UI. Set to false to use an existing service account or the default account. |
| login.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| login.startupProbe | object | `{"enabled":false,"failureThreshold":30,"periodSeconds":1}` | Startup probe configuration. The startup probe runs before liveness and readiness probes, allowing slow-starting containers to initialize. |
| login.startupProbe.enabled | bool | `false` | Enable or disable the startup probe. When enabled, liveness and readiness probes are disabled until the startup probe succeeds. |
| login.startupProbe.failureThreshold | int | `30` | Number of consecutive failures before marking startup as failed and restarting the container. |
| login.startupProbe.periodSeconds | int | `1` | How often (in seconds) to perform the startup check. |
| login.tolerations | list | `[]` | Tolerations allow pods to be scheduled on nodes with matching taints. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |
| login.topologySpreadConstraints | list | `[]` | Topology spread constraints control how pods are distributed across topology domains (e.g., zones, nodes) for high availability. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/ |
| metrics | object | `{"enabled":false,"serviceMonitor":{"additionalLabels":{},"enabled":false,"honorLabels":false,"honorTimestamps":true,"metricRelabellings":[],"namespace":null,"proxyUrl":null,"relabellings":[],"scheme":null,"scrapeInterval":null,"scrapeTimeout":null,"tlsConfig":{}}}` | Metrics configuration for ZITADEL monitoring. When enabled, ZITADEL exposes Prometheus-compatible metrics at /debug/metrics on the main service port. These metrics include Go runtime statistics, gRPC metrics, and ZITADEL-specific operational metrics like projection event counts. |
| metrics.enabled | bool | `false` | Enable metrics scraping annotations on ZITADEL pods. When true, adds prometheus.io/* annotations that enable automatic discovery by Prometheus. |
| metrics.serviceMonitor | object | `{"additionalLabels":{},"enabled":false,"honorLabels":false,"honorTimestamps":true,"metricRelabellings":[],"namespace":null,"proxyUrl":null,"relabellings":[],"scheme":null,"scrapeInterval":null,"scrapeTimeout":null,"tlsConfig":{}}` | ServiceMonitor configuration for Prometheus Operator. A ServiceMonitor is a custom resource that tells Prometheus Operator how to scrape metrics from ZITADEL. Requires the Prometheus Operator to be installed in the cluster. Ref: https://github.com/prometheus-operator/prometheus-operator |
| metrics.serviceMonitor.additionalLabels | object | `{}` | Additional labels to add to the ServiceMonitor. Use this to match Prometheus Operator's serviceMonitorSelector if configured. |
| metrics.serviceMonitor.honorLabels | bool | `false` | If true, use metric labels from ZITADEL instead of relabeling them to match Prometheus conventions. |
| metrics.serviceMonitor.honorTimestamps | bool | `true` | If true, preserve original scrape timestamps from ZITADEL instead of using Prometheus server time. |
| metrics.serviceMonitor.metricRelabellings | list | `[]` | Relabeling rules applied to individual metrics. Use to rename metrics, drop expensive metrics, or modify metric labels. |
| metrics.serviceMonitor.namespace | string | `nil` | Namespace where the ServiceMonitor should be created. If null, uses the release namespace. Set this if Prometheus watches a specific namespace. |
| metrics.serviceMonitor.proxyUrl | string | `nil` | HTTP proxy URL for scraping. Use if Prometheus needs to access ZITADEL through a proxy. |
| metrics.serviceMonitor.relabellings | list | `[]` | Relabeling rules applied before ingestion. Use to modify, filter, or drop labels before metrics are stored. |
| metrics.serviceMonitor.scheme | string | `nil` | HTTP scheme to use for scraping. Set to "https" if ZITADEL has internal TLS enabled. If null, defaults to "http". |
| metrics.serviceMonitor.scrapeInterval | string | `nil` | How often Prometheus should scrape metrics from ZITADEL. If null, uses Prometheus's default scrape interval (typically 30s). |
| metrics.serviceMonitor.scrapeTimeout | string | `nil` | Timeout for scrape requests. If null, uses Prometheus's default timeout. Should be less than scrapeInterval. |
| metrics.serviceMonitor.tlsConfig | object | `{}` | TLS configuration for scraping HTTPS endpoints. Configure this if ZITADEL has internal TLS enabled and you need to verify certificates. |
| nameOverride | string | `""` | Override the "zitadel" portion of resource names. Useful when the default naming would conflict with existing resources or when deploying multiple instances with different configurations. |
| nodeSelector | object | `{}` | Node labels for pod assignment. Pods will only be scheduled on nodes that have all the specified labels. Use this to target specific node pools (e.g., high-memory nodes, nodes in specific zones). Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ |
| pdb | object | `{"annotations":{},"enabled":false,"maxUnavailable":null,"minAvailable":null}` | Pod Disruption Budget configuration for the main ZITADEL deployment. Ensures high availability by limiting the number of pods that can be simultaneously unavailable during voluntary disruptions (e.g., node drains, rolling updates). Either minAvailable or maxUnavailable can be set, but not both. Values can be an integer (e.g., 1) or a percentage (e.g., "50%"). |
| pdb.annotations | object | `{}` | Additional annotations to apply to the Pod Disruption Budget resource |
| pdb.enabled | bool | `false` | Enable or disable the Pod Disruption Budget for ZITADEL pods |
| pdb.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailable during disruptions. Cannot be used together with minAvailable. |
| pdb.minAvailable | string | `nil` | Minimum number of pods that must remain available during disruptions. Cannot be used together with maxUnavailable. |
| podAdditionalLabels | object | `{}` | Additional labels to add to ZITADEL pods beyond the standard Helm labels. Useful for organizing pods with custom label selectors, network policies, or pod security policies. |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{"fsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` | Chart-wide pod security context applied to all pods (ZITADEL, Login UI, jobs) by default. Individual components can override these via their own podSecurityContext settings (e.g., login.podSecurityContext). Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| podSecurityContext.fsGroup | int | `1000` | Group ID for volume ownership and file creation. Files created in mounted volumes will be owned by this group. |
| podSecurityContext.runAsNonRoot | bool | `true` | Require containers to run as a non-root user. This is a security best practice that prevents privilege escalation attacks. |
| podSecurityContext.runAsUser | int | `1000` | User ID to run the container processes. 1000 is a common non-root UID that matches the ZITADEL container's default user. |
| readinessProbe | object | `{"enabled":true,"failureThreshold":3,"initialDelaySeconds":0,"periodSeconds":5}` | Readiness probe configuration for ZITADEL. The readiness probe determines when a pod is ready to receive traffic. Failed probes remove the pod from service endpoints, preventing traffic from being routed to unhealthy pods. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/ |
| readinessProbe.enabled | bool | `true` | Enable or disable the readiness probe. |
| readinessProbe.failureThreshold | int | `3` | Number of consecutive failures before marking the pod as not ready. |
| readinessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting readiness checks after container start. Set higher if ZITADEL needs time to initialize before accepting traffic. |
| readinessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the readiness check. |
| replicaCount | int | `1` | Number of ZITADEL pod replicas. While a single replica is fine for testing, production environments should use 3 or more to prevent downtime during rolling updates, node failures, and ensure high availability. |
| resources | object | `{}` | CPU and memory resource requests and limits for the ZITADEL container. Setting appropriate resources ensures predictable performance and prevents resource starvation. Requests affect scheduling; limits enforce caps. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| securityContext | object | `{"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1000}` | Chart-wide container security context applied to all containers by default. Individual components can override these via their own securityContext settings (e.g., login.securityContext). Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| securityContext.privileged | bool | `false` | Prevent the container from running in privileged mode. Privileged containers have access to all host devices and capabilities. |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount the container's root filesystem as read-only. This prevents malicious processes from writing to the filesystem and is a security best practice for immutable containers. |
| securityContext.runAsNonRoot | bool | `true` | Require the container to run as a non-root user. |
| securityContext.runAsUser | int | `1000` | User ID to run the container process. |
| service | object | `{"annotations":{"traefik.ingress.kubernetes.io/service.serversscheme":"h2c"},"appProtocol":"kubernetes.io/h2c","clusterIP":"","externalTrafficPolicy":"","labels":{},"port":8080,"protocol":"http2","scheme":"HTTP","type":"ClusterIP"}` | Kubernetes Service configuration for the main ZITADEL application. This service exposes ZITADEL's HTTP/2 API which serves both REST and gRPC traffic. |
| service.annotations | object | `{"traefik.ingress.kubernetes.io/service.serversscheme":"h2c"}` | Annotations to add to the Service resource. The default annotation tells Traefik to use HTTP/2 when communicating with ZITADEL backends. |
| service.appProtocol | string | `"kubernetes.io/h2c"` | Application protocol hint for ingress controllers and service meshes. "kubernetes.io/h2c" indicates HTTP/2 over cleartext (without TLS). |
| service.clusterIP | string | `""` | Fixed cluster IP address for ClusterIP services. Leave empty for automatic assignment by Kubernetes. Only applicable when type is "ClusterIP". Setting a fixed IP is useful when other services need a stable endpoint. |
| service.externalTrafficPolicy | string | `""` | Traffic policy for LoadBalancer and NodePort services. "Cluster" distributes traffic across all nodes (default), "Local" only routes to nodes with pods, preserving client source IP but potentially causing uneven load distribution. |
| service.labels | object | `{}` | Labels to add to the Service resource. Use for organizing services or matching service selectors in network policies. |
| service.port | int | `8080` | Port number the service exposes. Clients and ingress controllers connect to this port. ZITADEL uses 8080 by default for its HTTP/2 server. |
| service.protocol | string | `"http2"` | Protocol identifier used in port naming. ZITADEL uses HTTP/2 (h2c) for both REST API and gRPC traffic on the same port. |
| service.scheme | string | `"HTTP"` | HTTP scheme for health checks and internal communication. Use "HTTP" when TLS termination happens at the ingress/load balancer, or "HTTPS" when ZITADEL is configured with internal TLS. |
| service.type | string | `"ClusterIP"` | Service type. Use "ClusterIP" for internal-only access (typical when using an ingress controller), "NodePort" for direct node-level access, or "LoadBalancer" for cloud provider load balancers. |
| serviceAccount | object | `{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"},"create":true,"name":""}` | ServiceAccount configuration for ZITADEL pods. The service account controls what Kubernetes API permissions the pods have. |
| serviceAccount.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations to add to the service account. The default Helm hooks ensure the service account exists before pods are created. Add annotations here for cloud provider integrations (e.g., AWS IAM roles, GCP Workload Identity). |
| serviceAccount.create | bool | `true` | Whether to create a dedicated service account for ZITADEL. Set to false if you want to use an existing service account or the default account. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. Set this to use a specific existing service account when create is false. |
| setupJob | object | `{"activeDeadlineSeconds":300,"additionalArgs":["--init-projections=true"],"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"2"},"backoffLimit":5,"extraContainers":[],"initContainers":[],"machinekeyWriter":{"image":{"repository":"","tag":""},"resources":{}},"podAdditionalLabels":{},"podAnnotations":{},"resources":{}}` | Setup Job configuration for ZITADEL application setup. This Kubernetes Job runs as a Helm pre-install/pre-upgrade hook after the init job to configure ZITADEL itself. It creates the first organization, admin user, machine users, and generates authentication credentials stored as Kubernetes Secrets. |
| setupJob.activeDeadlineSeconds | int | `300` | Maximum time in seconds for the setup job to complete. The job is terminated if it exceeds this deadline. Increase this for slow environments. |
| setupJob.additionalArgs | list | `["--init-projections=true"]` | Additional command-line arguments to pass to the ZITADEL setup command. The default enables projection initialization for better startup performance. |
| setupJob.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"2"}` | Annotations for the setup job. The Helm hooks ensure this job runs after the init job (weight "2" > "1") and before the main deployment. |
| setupJob.backoffLimit | int | `5` | Number of retries before marking the setup job as failed. |
| setupJob.extraContainers | list | `[]` | Sidecar containers to run alongside the setup container. Useful for logging, proxies (e.g., cloud-sql-proxy), or other supporting services. |
| setupJob.initContainers | list | `[]` | Init containers to run before the main setup container. Useful for waiting on additional dependencies or performing pre-setup tasks. |
| setupJob.machinekeyWriter | object | `{"image":{"repository":"","tag":""},"resources":{}}` | Configuration for the sidecar container that writes machine keys and PATs to Kubernetes Secrets. This container runs alongside the setup container and uses kubectl to create secrets from the generated credentials. |
| setupJob.machinekeyWriter.image | object | `{"repository":"","tag":""}` | Image configuration for the machinekey writer container. |
| setupJob.machinekeyWriter.image.repository | string | `""` | Override the default kubectl image repository. Leave empty to use the value from tools.kubectl.image.repository. |
| setupJob.machinekeyWriter.image.tag | string | `""` | Override the default kubectl image tag. Leave empty to use the value from tools.kubectl.image.tag (which defaults to cluster version). |
| setupJob.machinekeyWriter.resources | object | `{}` | CPU and memory resource requests and limits for the machinekey writer container. This container only runs kubectl commands and needs minimal resources. |
| setupJob.podAdditionalLabels | object | `{}` | Additional labels to add to setup job pods. |
| setupJob.podAnnotations | object | `{}` | Additional annotations to add to setup job pods. |
| setupJob.resources | object | `{}` | CPU and memory resource requests and limits for the setup job container. The setup job performs more work than init, including generating keys and creating initial data. |
| startupProbe | object | `{"enabled":true,"failureThreshold":30,"periodSeconds":1}` | Startup probe configuration for ZITADEL. The startup probe runs before liveness and readiness probes begin, allowing slow-starting containers to fully initialize. This is especially useful for ZITADEL which may need time to complete database migrations on first start. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/ |
| startupProbe.enabled | bool | `true` | Enable or disable the startup probe. When enabled, liveness and readiness probes are disabled until the startup probe succeeds. |
| startupProbe.failureThreshold | int | `30` | Number of consecutive failures before marking startup as failed and restarting the container. With periodSeconds=1 and failureThreshold=30, the container has 30 seconds to start. |
| startupProbe.periodSeconds | int | `1` | How often (in seconds) to perform the startup check. |
| tolerations | list | `[]` | Tolerations allow pods to be scheduled on nodes with matching taints. Taints are used to repel pods from nodes; tolerations allow exceptions. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |
| tools | object | `{"kubectl":{"image":{"pullPolicy":"","repository":"alpine/k8s","tag":""}},"wait4x":{"image":{"pullPolicy":"","repository":"wait4x/wait4x","tag":"3.6"},"resources":{}}}` | Configuration for helper tools used by init containers and jobs. These images are used by components such as wait-for-zitadel, wait-for-postgres, and the setup and cleanup jobs. Each tool follows the standard image configuration pattern with registry, repository, tag, pull policy, and pull secrets. |
| tools.kubectl | object | `{"image":{"pullPolicy":"","repository":"alpine/k8s","tag":""}}` | Configuration for the kubectl helper image used by init containers and jobs for lightweight Kubernetes API operations. This image is used by the setup job's machinekey containers and the cleanup job. |
| tools.kubectl.image.pullPolicy | string | `""` | The pull policy for the kubectl image. If left empty, Kubernetes applies its default policy depending on whether the tag is mutable or fixed. |
| tools.kubectl.image.repository | string | `"alpine/k8s"` | The name of the image repository that contains the kubectl image. The chart automatically prepends the registry (docker.io by default) for compatibility with CRI-O v1.34+ which enforces fully qualified names. |
| tools.kubectl.image.tag | string | `""` | The image tag to use for the kubectl image. It should be left empty to automatically default to the Kubernetes cluster version |
| tools.wait4x | object | `{"image":{"pullPolicy":"","repository":"wait4x/wait4x","tag":"3.6"},"resources":{}}` | Configuration for the wait4x image used for readiness and dependency checks in init containers. Values are intentionally left empty and should be set by the user when overriding this configuration. |
| tools.wait4x.image.pullPolicy | string | `""` | The pull policy for the wait4x image. If left empty, the chart defaults to the Kubernetes default pull policy for the given tag. |
| tools.wait4x.image.repository | string | `"wait4x/wait4x"` | The name of the image repository that contains the wait4x image. The chart automatically prepends the registry (docker.io by default) for compatibility with CRI-O v1.34+ which enforces fully qualified names. |
| tools.wait4x.image.tag | string | `"3.6"` | The image tag to use for the wait4x image. Leave empty to require the user to set a specific version explicitly. |
| tools.wait4x.resources | object | `{}` | CPU and memory resource requests and limits for wait4x init containers. These resources apply to all init containers using the wait4x tool, including wait-for-zitadel and wait-for-postgres. Setting equal requests and limits enables the "Guaranteed" QoS class when combined with resource settings on the main container. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| topologySpreadConstraints | list | `[]` | Topology spread constraints control how pods are distributed across topology domains (e.g., zones, nodes, regions) for high availability. Unlike affinity, these constraints provide more granular control over pod distribution. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/ |
| zitadel.autoscaling | object | `{"annotations":{},"behavior":{},"enabled":false,"maxReplicas":10,"metrics":[],"minReplicas":3,"targetCPU":null,"targetMemory":null}` | Horizontal Pod Autoscaler configuration for scaling on CPU, memory, or custom metrics. Ref: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/ |
| zitadel.autoscaling.annotations | object | `{}` | Optional map of annotations applied to the HPA object. |
| zitadel.autoscaling.enabled | bool | `false` | If true, enables the Horizontal Pod Autoscaler for the Zitadel deployment. This will automatically override the `replicaCount` value. |
| zitadel.autoscaling.maxReplicas | int | `10` | The maximum number of pod replicas. |
| zitadel.autoscaling.metrics | list | `[]` | Advanced scaling based on custom metrics exposed by Zitadel. Zitadel exposes standard Go runtime metrics. To use these for scaling, you MUST have a metrics server (e.g., Prometheus) and a metrics adapter (e.g., prometheus-adapter) running in your cluster. Ref: https://github.com/kubernetes-sigs/prometheus-adapter |
| zitadel.autoscaling.minReplicas | int | `3` | The minimum number of pod replicas. |
| zitadel.autoscaling.targetCPU | string | `nil` | The target average CPU utilization percentage. |
| zitadel.autoscaling.targetMemory | string | `nil` | The target average memory utilization percentage. |
| zitadel.configSecretKey | string | `"config-yaml"` | The key within the configSecretName secret that contains the ZITADEL configuration YAML. The default "config-yaml" matches the expected format. |
| zitadel.configSecretName | string | `nil` | Name of an existing Kubernetes Secret containing ZITADEL configuration. Use this when you want to manage ZITADEL configuration externally (e.g., via External Secrets Operator, Sealed Secrets, or GitOps). The secret should contain YAML configuration in the same format as configmapConfig. |
| zitadel.configmapConfig | object | `{"Database":{"Postgres":{"Host":"","Port":5432}},"ExternalDomain":"","ExternalSecure":true,"FirstInstance":{"LoginClientPatPath":null,"MachineKeyPath":null,"Org":{"LoginClient":{"Machine":{"Name":"Automatically Initialized IAM Login Client","Username":"login-client"},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Machine":{"Machine":{"Name":"Automatically Initialized IAM Admin","Username":"iam-admin"},"MachineKey":{"ExpirationDate":"2029-01-01T00:00:00Z","Type":1},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Skip":null},"PatPath":null,"Skip":false},"Machine":{"Identification":{"Hostname":{"Enabled":true},"Webhook":{"Enabled":false}}},"TLS":{"Enabled":false}}` | ZITADEL runtime configuration written to a Kubernetes ConfigMap. These values are passed directly to the ZITADEL binary and control its behavior. For the complete list of available configuration options, see: https://github.com/zitadel/zitadel/blob/main/cmd/defaults.yaml |
| zitadel.configmapConfig.Database | object | `{"Postgres":{"Host":"","Port":5432}}` | Database connection configuration. ZITADEL requires PostgreSQL 14+ or CockroachDB 22+ as its backing database. |
| zitadel.configmapConfig.Database.Postgres.Host | string | `""` | PostgreSQL server hostname or IP address. Leave empty if providing via configSecretName or environment variables. |
| zitadel.configmapConfig.Database.Postgres.Port | int | `5432` | PostgreSQL server port number. |
| zitadel.configmapConfig.ExternalDomain | string | `""` | The domain name used for external access to ZITADEL (e.g., "auth.example.com"). This is used for OIDC issuer URLs, redirect URIs, and cookie domains. Must match the domain configured in your ingress or load balancer. |
| zitadel.configmapConfig.ExternalSecure | bool | `true` | Whether external connections use HTTPS. Set to true when ZITADEL is behind a TLS-terminating proxy or ingress controller. This affects URL generation for redirects, OIDC discovery, and other external-facing endpoints. |
| zitadel.configmapConfig.FirstInstance | object | `{"LoginClientPatPath":null,"MachineKeyPath":null,"Org":{"LoginClient":{"Machine":{"Name":"Automatically Initialized IAM Login Client","Username":"login-client"},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Machine":{"Machine":{"Name":"Automatically Initialized IAM Admin","Username":"iam-admin"},"MachineKey":{"ExpirationDate":"2029-01-01T00:00:00Z","Type":1},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Skip":null},"PatPath":null,"Skip":false}` | Configures the initial instance created by the ZITADEL setup job. The values defined here are used to bootstrap the first organization and its users. |
| zitadel.configmapConfig.FirstInstance.MachineKeyPath | string | `nil` | IMPORTANT: The following fields are managed automatically by the Helm chart. Setting these manually is not supported and will cause the deployment to fail. The chart automatically generates and manages these paths internally. |
| zitadel.configmapConfig.FirstInstance.Org.LoginClient | object | `{"Machine":{"Name":"Automatically Initialized IAM Login Client","Username":"login-client"},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}}` | Defines the internal machine user required by the ZITADEL Login UI. This user communicates with the main ZITADEL API to handle login flows. It is a specialized, internal-use-only account and is not intended for general administrative use. A PAT is generated and stored in a Kubernetes secret named 'login-client'. |
| zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat | object | `{"ExpirationDate":"2029-01-01T00:00:00Z"}` | The expiration date for the Login UI's Personal Access Token (PAT). |
| zitadel.configmapConfig.FirstInstance.Org.Machine | object | `{"Machine":{"Name":"Automatically Initialized IAM Admin","Username":"iam-admin"},"MachineKey":{"ExpirationDate":"2029-01-01T00:00:00Z","Type":1},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}}` | Defines an administrative machine user created with the IAM_OWNER role. This user is intended for initial automation and administrative tasks right after the Helm chart is deployed.  By default, a JWT Machine Key is generated and stored in a Kubernetes secret named after the `Username` (e.g., 'iam-admin').  If the `Pat` block below is defined, a Personal Access Token (PAT) is also generated and stored in a separate Kubernetes secret. This secret will be named after the `Username` with a '-pat' suffix (e.g., 'iam-admin-pat'). |
| zitadel.configmapConfig.FirstInstance.Org.Machine.Pat | object | `{"ExpirationDate":"2029-01-01T00:00:00Z"}` | If this Pat object is present, a PAT will also be created. |
| zitadel.configmapConfig.FirstInstance.Org.Skip | string | `nil` | IMPORTANT: Setting Skip here is not supported. Use FirstInstance.Skip instead. This field exists for validation purposes only. |
| zitadel.configmapConfig.Machine | object | `{"Identification":{"Hostname":{"Enabled":true},"Webhook":{"Enabled":false}}}` | Machine identification settings for ZITADEL instances in a cluster. |
| zitadel.configmapConfig.Machine.Identification.Hostname | object | `{"Enabled":true}` | Use the pod hostname for machine identification. Recommended for Kubernetes deployments where each pod has a unique hostname. |
| zitadel.configmapConfig.Machine.Identification.Webhook | object | `{"Enabled":false}` | Use a webhook for machine identification. Alternative to hostname-based identification for specialized deployment scenarios. |
| zitadel.configmapConfig.TLS | object | `{"Enabled":false}` | Internal TLS configuration for ZITADEL's HTTP server. When enabled, ZITADEL serves HTTPS directly instead of relying on a proxy for TLS termination. |
| zitadel.configmapConfig.TLS.Enabled | bool | `false` | Enable HTTPS on ZITADEL's internal server. Requires serverSslCrtSecret or selfSignedCert to be configured with valid certificates. |
| zitadel.dbSslAdminCrtSecret | string | `""` | Name of a Kubernetes Secret containing the admin user's client certificate for mutual TLS (mTLS) authentication to the database. The secret must contain keys "tls.crt" (certificate) and "tls.key" (private key). Used by the init job for database setup operations that require elevated privileges. |
| zitadel.dbSslCaCrt | string | `""` | PEM-encoded CA certificate for verifying the database server's TLS certificate. Use this when your PostgreSQL/CockroachDB server uses a self-signed certificate or a certificate signed by a private CA. The certificate is stored in a Kubernetes Secret and mounted into ZITADEL pods at /db-ssl-ca-crt/ca.crt. Either provide the certificate inline here, or reference an existing secret using dbSslCaCrtSecret instead. |
| zitadel.dbSslCaCrtAnnotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the dbSslCaCrt Secret when created from the inline certificate. The default Helm hooks ensure the secret exists before pods start. |
| zitadel.dbSslCaCrtSecret | string | `""` | Name of an existing Kubernetes Secret containing the database CA certificate at key "ca.crt". Use this instead of dbSslCaCrt when the certificate is managed externally (e.g., by cert-manager or an operator). The secret must exist in the same namespace as the ZITADEL release. |
| zitadel.dbSslUserCrtSecret | string | `""` | Name of a Kubernetes Secret containing the application user's client certificate for mutual TLS (mTLS) authentication to the database. The secret must contain keys "tls.crt" (certificate) and "tls.key" (private key). Used by the main ZITADEL deployment and setup job for normal database operations. |
| zitadel.debug | object | `{"annotations":{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-weight":"1"},"enabled":false,"extraContainers":[],"initContainers":[]}` | Debug pod configuration. When enabled, creates a standalone pod with the ZITADEL binary and configuration mounted, useful for troubleshooting and running ad-hoc ZITADEL commands. The pod starts in sleep mode, allowing you to exec into it and run commands interactively. Usage: kubectl exec -it <debug-pod-name> -- /bin/sh Then run: zitadel --help |
| zitadel.debug.annotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-weight":"1"}` | Annotations for the debug pod. The Helm hooks ensure it's created during install/upgrade and cleaned up appropriately. |
| zitadel.debug.enabled | bool | `false` | Enable or disable the debug pod. Only enable for troubleshooting; disable in production environments. |
| zitadel.debug.extraContainers | list | `[]` | Sidecar containers to run alongside the debug container. |
| zitadel.debug.initContainers | list | `[]` | Init containers to run before the debug container starts. |
| zitadel.extraContainers | list | `[]` | Global sidecar containers added to all ZITADEL workloads (Deployment, init job, setup job, and debug pod when enabled). Use this for shared services like database proxies (e.g., cloud-sql-proxy) that all workloads need to connect to the database. |
| zitadel.initContainers | list | `[]` | Global init containers added to all ZITADEL workloads (Deployment, init job, setup job, and debug pod when enabled). Use this for shared dependencies like database readiness checks or certificate initialization that all workloads need. |
| zitadel.masterkey | string | `""` | ZITADEL's masterkey for symmetric encryption of sensitive data like private keys and tokens. Must be exactly 32 characters (256 bits). Generate with: tr -dc A-Za-z0-9 </dev/urandom | head -c 32 IMPORTANT: Store this value securely. Loss of the masterkey means loss of all encrypted data. Either set this value or use masterkeySecretName. |
| zitadel.masterkeyAnnotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the masterkey Secret when created from zitadel.masterkey. The default Helm hooks ensure the secret exists before pods start. |
| zitadel.masterkeySecretName | string | `""` | Name of an existing Kubernetes Secret containing the masterkey at key "masterkey". Use this for production deployments to avoid storing the masterkey in values files. The secret must exist before chart installation. Note: Either zitadel.masterkey or zitadel.masterkeySecretName must be set. |
| zitadel.podSecurityContext | object | `{}` | Optional overrides for the pod and container security contexts used by Zitadel pods. If left empty, the chart-wide podSecurityContext and securityContext defined below are used. |
| zitadel.revisionHistoryLimit | int | `10` | Number of old ReplicaSets to retain for rollback purposes Set to 0 to not keep any old ReplicaSets |
| zitadel.secretConfig | string | `nil` | Sensitive ZITADEL configuration values written to a Kubernetes Secret instead of a ConfigMap. Use this for database passwords, API keys, SMTP credentials, and other values that should not be stored in plain text. The secret is mounted alongside the ConfigMap and both are merged by ZITADEL at startup. Structure follows the same format as configmapConfig. See all options: https://github.com/zitadel/zitadel/blob/main/cmd/defaults.yaml Example:   secretConfig:     Database:       Postgres:         User:           Password: "my-secure-password" |
| zitadel.secretConfigAnnotations | object | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the secretConfig Secret. The default Helm hooks ensure the secret is created before the deployment and recreated on upgrades. |
| zitadel.securityContext | object | `{}` |  |
| zitadel.selfSignedCert | object | `{"additionalDnsName":"","enabled":false}` | Generate a self-signed TLS certificate using Helm's native functions. If enabled, this creates a Kubernetes secret containing a self-signed certificate and mounts it to /etc/tls/ in the ZITADEL pod.  This is useful for quick demos or test environments without an ingress controller. It is NOT recommended for production use.  The certificate will be valid for the `ExternalDomain`, `localhost`, and any host specified in `additionalDnsName`. It can't use dynamic values like the Pod IP, which the previous initContainer method did. |
| zitadel.selfSignedCert.enabled | bool | `false` | Enable generation of a self-signed TLS certificate. |
| zitadel.serverSslCrtSecret | string | `""` | Name of a Kubernetes Secret containing the TLS certificate for ZITADEL's internal HTTPS server. The secret must contain keys "tls.crt" (certificate) and "tls.key" (private key). Use this when ZITADEL should serve HTTPS directly instead of relying on TLS termination at an ingress controller or load balancer. Requires configmapConfig.TLS.Enabled to be true. |

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
kind create cluster --config ./test/acceptance/kindConfig.yaml

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
