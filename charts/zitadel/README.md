[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/zitadel)](https://artifacthub.io/packages/search?repo=zitadel)

# Zitadel

![Version: 9.23.0](https://img.shields.io/badge/Version-9.23.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v4.10.1](https://img.shields.io/badge/AppVersion-v4.10.1-informational?style=flat-square)

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
| affinity | Affinity | `{}` | Affinity rules for pod scheduling. Use for advanced pod placement strategies like co-locating pods on the same node (pod affinity), spreading pods across zones (pod anti-affinity), or preferring certain nodes (node affinity). Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| annotations | map[string]string | `{}` | Annotations to add to the ZITADEL Deployment resource. Use this for integration with tools like ArgoCD, Flux, or external monitoring systems. |
| cleanupJob.activeDeadlineSeconds | int | `60` | Maximum time in seconds for the cleanup job to complete. After this deadline, the job is terminated even if still running. |
| cleanupJob.annotations | map[string]string | `{"helm.sh/hook":"post-delete","helm.sh/hook-delete-policy":"hook-succeeded","helm.sh/hook-weight":"-1"}` | Annotations for the cleanup job. The post-delete hook ensures this runs on helm uninstall, and the delete policy removes the job after completion. |
| cleanupJob.backoffLimit | int | `3` | Number of retries before marking the cleanup job as failed. |
| cleanupJob.enabled | bool | `true` | Enable the cleanup job to remove secrets created by the setup job. Set to false if you want to preserve secrets across reinstalls. |
| cleanupJob.podAdditionalLabels | map[string]string | `{}` | Additional labels to add to cleanup job pods. |
| cleanupJob.podAnnotations | map[string]string | `{}` | Additional annotations to add to cleanup job pods. |
| cleanupJob.resources | ResourceRequirements | `{}` | Resource limits and requests for the cleanup job container. Keep minimal as this job only runs kubectl delete commands. |
| configMap.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the ZITADEL ConfigMap. The default Helm hooks ensure the ConfigMap is created before the deployment and recreated on upgrades to pick up configuration changes. |
| env | []EnvVar | `[]` | Additional environment variables for the ZITADEL container. Use this to pass configuration that isn't available through configmapConfig or secretConfig, or to inject values from other Kubernetes resources like ConfigMaps or Secrets. ZITADEL environment variables follow the pattern ZITADEL_<SECTION>_<KEY>. Ref: https://zitadel.com/docs/self-hosting/manage/configure#configure-by-environment-variables |
| envVarsSecret | string | `""` | Name of a Kubernetes Secret containing environment variables to inject into the ZITADEL container. All key-value pairs in the secret will be available as environment variables. This is useful for managing multiple ZITADEL configuration values in a single secret, especially when using external secret management tools like External Secrets Operator or Sealed Secrets. Ref: https://zitadel.com/docs/self-hosting/manage/configure#configure-by-environment-variables |
| extraContainers | []Container | `[]` | Sidecar containers to run alongside the main ZITADEL container in the Deployment pod. Use this for logging agents, monitoring sidecars, service meshes, or database proxies (e.g., cloud-sql-proxy for Google Cloud SQL). These containers share the pod's network namespace and can access the same volumes as the main container. |
| extraManifests | []object | `[]` | Additional Kubernetes manifests to deploy alongside the chart. This allows you to include custom resources without creating a separate chart. Supports Helm templating syntax including .Release, .Values, and template functions. Use this for secrets, configmaps, network policies, or any other resources that ZITADEL depends on. |
| extraVolumeMounts | []VolumeMount | `[]` | Additional volume mounts for the main ZITADEL container. Use this to mount volumes defined in extraVolumes into the container filesystem. Common use cases include mounting custom CA certificates, configuration files, or shared data between containers. |
| extraVolumes | []Volume | `[]` | Additional volumes to add to ZITADEL pods. These volumes can be referenced by extraVolumeMounts to make data available to the ZITADEL container or sidecar containers. Supports all Kubernetes volume types: secrets, configMaps, persistentVolumeClaims, emptyDir, hostPath, etc. |
| fullnameOverride | string | `""` | Completely override the generated resource names (release-name + chart-name). Takes precedence over nameOverride. Use this when you need full control over resource naming, such as when migrating from another chart. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. Use "Always" for mutable tags like "latest", or "IfNotPresent" for immutable version tags to reduce network traffic. |
| image.repository | string | `"ghcr.io/zitadel/zitadel"` | Docker image repository for ZITADEL. The default uses GitHub Container Registry. Change this if using a private registry or mirror. |
| image.tag | string | `""` | Image tag. Defaults to the chart's appVersion if not specified. Use a specific version tag (e.g., "v2.45.0") for production deployments to ensure reproducibility and controlled upgrades. |
| imagePullSecrets | []LocalObjectReference | `[]` | References to secrets containing Docker registry credentials for pulling private ZITADEL images. Each entry should be the name of an existing secret of type kubernetes.io/dockerconfigjson. Example:   imagePullSecrets:     - name: my-registry-secret |
| imageRegistry | string | `""` | Global container registry override for tool images (e.g., wait4x, kubectl). When set, this registry is prepended to tool image repositories for compatibility with CRI-O v1.34+ which enforces fully qualified image names. If left empty, defaults to "docker.io". |
| ingress.annotations | map[string]string | `{"nginx.ingress.kubernetes.io/backend-protocol":"GRPC"}` | Annotations to apply to the Ingress resource. The default annotation is for NGINX to correctly handle gRPC traffic. |
| ingress.className | string | `""` | The name of the IngressClass resource to use for this Ingress. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
| ingress.controller | string | `"generic"` | A chart-specific setting to enable logic for different controllers. Use "aws" to generate AWS ALB-specific annotations and resources. |
| ingress.enabled | bool | `false` | If true, creates an Ingress resource for the ZITADEL service. |
| ingress.hosts | list | `[{"paths":[{"path":"/","pathType":"Prefix"}]}]` | A list of host rules for the Ingress. Each host can have multiple paths. |
| ingress.tls | []IngressTLS | `[]` | TLS configuration for the Ingress. This allows you to secure the endpoint with HTTPS by referencing a secret that contains the TLS certificate and key. |
| initJob.activeDeadlineSeconds | int | `300` | Maximum time in seconds for the init job to complete. The job is terminated if it exceeds this deadline, regardless of backoffLimit. |
| initJob.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"1"}` | Annotations for the init job. The Helm hooks ensure this job runs before the main deployment and is recreated on each upgrade. |
| initJob.backoffLimit | int | `5` | Number of retries before marking the init job as failed. Increase this if the database might take longer to become available. |
| initJob.command | string | `""` |  |
| initJob.enabled | bool | `true` | Enable or disable the init job. Set to false after the initial installation if you want to manage database initialization externally or if no database changes are expected during upgrades. |
| initJob.extraContainers | []Container | `[]` | Sidecar containers to run alongside the init container. Useful for logging, proxies (e.g., cloud-sql-proxy), or other supporting services. |
| initJob.initContainers | []Container | `[]` | Init containers to run before the main init container. Useful for waiting on additional dependencies or performing pre-initialization tasks. |
| initJob.podAdditionalLabels | map[string]string | `{}` | Additional labels to add to init job pods. |
| initJob.podAnnotations | map[string]string | `{}` | Additional annotations to add to init job pods. |
| initJob.resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for the init job container. The init job typically requires minimal resources as it only runs SQL commands against the database. |
| livenessProbe.enabled | bool | `true` | Enable or disable the liveness probe. |
| livenessProbe.failureThreshold | int | `3` | Number of consecutive failures before restarting the container. |
| livenessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting liveness checks after container start. |
| livenessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the liveness check. |
| login.affinity | Affinity | `{}` | Affinity rules for pod scheduling. Use for advanced pod placement strategies like co-locating pods or spreading across zones. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| login.annotations | map[string]string | `{}` | Annotations to add to the Login UI Deployment resource. |
| login.autoscaling.annotations | map[string]string | `{}` | Annotations applied to the HPA object. |
| login.autoscaling.behavior | HorizontalPodAutoscalerBehavior | `{}` | Configures the scaling behavior for scaling up and down. Use this to control how quickly the HPA scales pods in response to metric changes. Ref: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior |
| login.autoscaling.enabled | bool | `false` | If true, enables the Horizontal Pod Autoscaler for the login deployment. This will automatically override the `replicaCount` value. |
| login.autoscaling.maxReplicas | int | `10` | The maximum number of pod replicas. |
| login.autoscaling.metrics | []MetricSpec | `[]` | Advanced scaling based on custom metrics exposed by Zitadel. To use these for scaling, you MUST have a metrics server (e.g., Prometheus) and a metrics adapter (e.g., prometheus-adapter) running in your cluster. Ref: https://github.com/kubernetes-sigs/prometheus-adapter |
| login.autoscaling.minReplicas | int | `3` | The minimum number of pod replicas. |
| login.autoscaling.targetCPU | string | `nil` | The target average CPU utilization percentage. |
| login.autoscaling.targetMemory | string | `nil` | The target average memory utilization percentage. |
| login.configMap.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the Login UI ConfigMap. The default hooks ensure the ConfigMap is created before the deployment and recreated on upgrades. |
| login.customConfigmapConfig | string | `nil` | Custom environment variables for the Login UI ConfigMap. These override the default values which configure the service user token path, API URL, and custom request headers. Only set this if you need to customize the Login UI behavior beyond the defaults. The defaults are:   ZITADEL_SERVICE_USER_TOKEN_FILE="/login-client/pat"   ZITADEL_API_URL="http://<release>-zitadel:<port>"   CUSTOM_REQUEST_HEADERS="Host:<ExternalDomain>" |
| login.enabled | bool | `true` | Enable or disable the Login UI deployment. When disabled, ZITADEL uses its built-in login interface instead of the separate Login UI application. |
| login.env | []EnvVar | `[]` | Additional environment variables for the Login UI container. Use this to pass configuration that isn't available through customConfigmapConfig. |
| login.extraContainers | []Container | `[]` | Sidecar containers to run alongside the Login UI container. Useful for logging agents, proxies, or other supporting services. |
| login.extraVolumeMounts | []VolumeMount | `[]` | Additional volume mounts for the Login UI container. Use this to mount custom certificates, configuration files, or other data into the container. |
| login.extraVolumes | []Volume | `[]` | Additional volumes for the Login UI pod. Define volumes here that are referenced by extraVolumeMounts. |
| login.fullnameOverride | string | `""` | Completely override the generated resource names. Takes precedence over nameOverride when set. |
| login.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. Use "Always" for mutable tags like "latest", or "IfNotPresent" for immutable tags. |
| login.image.repository | string | `"ghcr.io/zitadel/zitadel-login"` | Docker image repository for the Login UI. |
| login.image.tag | string | `""` | Image tag. Defaults to the chart's appVersion if not specified. Use a specific version tag for production deployments to ensure reproducibility. |
| login.imagePullSecrets | []LocalObjectReference | `[]` | References to secrets containing Docker registry credentials for pulling private images. Each entry should be the name of an existing secret. |
| login.ingress.annotations | map[string]string | `{}` | Annotations to apply to the Login UI Ingress resource. |
| login.ingress.className | string | `""` | The name of the IngressClass resource to use for this Ingress. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
| login.ingress.controller | string | `"generic"` | A chart-specific setting to enable logic for different controllers. Use "aws" to generate AWS ALB-specific annotations. |
| login.ingress.enabled | bool | `false` | If true, creates an Ingress resource for the Login UI service. |
| login.ingress.hosts | list | `[{"paths":[{"path":"/ui/v2/login","pathType":"Prefix"}]}]` | A list of host rules for the Ingress. The default path targets the login UI. |
| login.ingress.tls | []IngressTLS | `[]` | TLS configuration for the Ingress. Secure the login UI with HTTPS by referencing a secret containing the TLS certificate and key. |
| login.initContainers | []Container | `[]` | Init containers to run before the Login UI container starts. Useful for waiting on dependencies or performing setup tasks. |
| login.livenessProbe.enabled | bool | `true` | Enable or disable the liveness probe. |
| login.livenessProbe.failureThreshold | int | `3` | Number of consecutive failures before restarting the container. |
| login.livenessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting liveness checks after container start. |
| login.livenessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the liveness check. |
| login.loginClientSecretPrefix | string | `nil` | Prefix for the login client secret name. Use this when deploying multiple ZITADEL instances in the same namespace to avoid secret name collisions. When set, the login client secret will be named "{prefix}login-client". |
| login.nameOverride | string | `""` | Override the "login" portion of resource names. Useful when the default naming conflicts with existing resources. |
| login.nodeSelector | map[string]string | `{}` | Node labels for pod assignment. Pods will only be scheduled on nodes with matching labels. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ |
| login.pdb.annotations | map[string]string | `{}` | Additional annotations to apply to the Pod Disruption Budget resource. |
| login.pdb.enabled | bool | `false` | Enable or disable the Pod Disruption Budget for Login UI pods. |
| login.pdb.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailable during disruptions. Cannot be used together with minAvailable. |
| login.pdb.minAvailable | string | `nil` | Minimum number of pods that must remain available during disruptions. Cannot be used together with maxUnavailable. |
| login.podAdditionalLabels | map[string]string | `{}` | Additional labels to add to Login UI pods beyond the standard Helm labels. Useful for organizing pods with custom label selectors. |
| login.podAnnotations | map[string]string | `{}` | Annotations to add to Login UI pods. Useful for integrations like Prometheus scraping, Istio sidecar injection, or Vault agent injection. |
| login.podSecurityContext | PodSecurityContext | `{}` | Optional pod-level security context overrides for Login UI pods. If left empty, the chart-wide podSecurityContext defined below is used instead. Use this to customize security settings specifically for the Login UI. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| login.readinessProbe.enabled | bool | `true` | Enable or disable the readiness probe. |
| login.readinessProbe.failureThreshold | int | `3` | Number of consecutive failures before marking the pod as not ready. |
| login.readinessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting readiness checks after container start. |
| login.readinessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the readiness check. |
| login.replicaCount | int | `1` | Number of Login UI pod replicas. A single replica is fine for testing, but production environments should use 3 or more for high availability during rolling updates and node failures. |
| login.resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for the Login UI container. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| login.revisionHistoryLimit | int | `10` | Number of old ReplicaSets to retain for rollback purposes. Set to 0 to disable rollback capability and save cluster resources. |
| login.securityContext | SecurityContext | `{}` | Optional container-level security context overrides for the Login UI container. If left empty, the chart-wide securityContext defined below is used instead. Use this to customize security settings specifically for the Login UI container. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| login.service.annotations | map[string]string | `{}` | Annotations to add to the Service resource. |
| login.service.appProtocol | string | `"kubernetes.io/http"` | Application protocol hint for ingress controllers and service meshes. Helps with protocol detection and routing decisions. |
| login.service.clusterIP | string | `""` | Fixed cluster IP address for ClusterIP services. Leave empty for automatic assignment. Only applicable when type is "ClusterIP". |
| login.service.externalTrafficPolicy | string | `""` | Traffic policy for LoadBalancer services. "Cluster" distributes traffic to all nodes, "Local" only routes to nodes with pods. Only applicable when type is "LoadBalancer". |
| login.service.labels | map[string]string | `{}` | Labels to add to the Service resource. |
| login.service.port | int | `3000` | Port number the service exposes. Clients connect to this port. |
| login.service.protocol | string | `"http"` | Protocol identifier used in port naming (e.g., "http", "https", "grpc"). |
| login.service.scheme | string | `"HTTP"` | HTTP scheme for health checks and internal communication. |
| login.service.type | string | `"ClusterIP"` | Service type. Use "ClusterIP" for internal access, "NodePort" for node-level access, or "LoadBalancer" for cloud provider load balancers. |
| login.serviceAccount.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the Login UI service account. The default Helm hooks ensure it exists before pods are created. Add annotations here for cloud provider integrations (e.g., AWS IAM roles, GCP Workload Identity). |
| login.serviceAccount.create | bool | `true` | Whether to create a dedicated service account for the Login UI. Set to false to use an existing service account or the default account. |
| login.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| login.startupProbe.enabled | bool | `false` | Enable or disable the startup probe. When enabled, liveness and readiness probes are disabled until the startup probe succeeds. |
| login.startupProbe.failureThreshold | int | `30` | Number of consecutive failures before marking startup as failed and restarting the container. |
| login.startupProbe.periodSeconds | int | `1` | How often (in seconds) to perform the startup check. |
| login.tolerations | []Toleration | `[]` | Tolerations allow pods to be scheduled on nodes with matching taints. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |
| login.topologySpreadConstraints | []TopologySpreadConstraint | `[]` | Topology spread constraints control how pods are distributed across topology domains (e.g., zones, nodes) for high availability. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/ |
| metrics.enabled | bool | `false` | Enable metrics scraping annotations on ZITADEL pods. When true, adds prometheus.io/* annotations that enable automatic discovery by Prometheus. |
| metrics.serviceMonitor.additionalLabels | map[string]string | `{}` | Additional labels to add to the ServiceMonitor. Use this to match Prometheus Operator's serviceMonitorSelector if configured. |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| metrics.serviceMonitor.honorLabels | bool | `false` | If true, use metric labels from ZITADEL instead of relabeling them to match Prometheus conventions. |
| metrics.serviceMonitor.honorTimestamps | bool | `true` | If true, preserve original scrape timestamps from ZITADEL instead of using Prometheus server time. |
| metrics.serviceMonitor.metricRelabellings | []RelabelConfig | `[]` | Relabeling rules applied to individual metrics. Use to rename metrics, drop expensive metrics, or modify metric labels. |
| metrics.serviceMonitor.namespace | string | `nil` | Namespace where the ServiceMonitor should be created. If null, uses the release namespace. Set this if Prometheus watches a specific namespace. |
| metrics.serviceMonitor.proxyUrl | string | `nil` | HTTP proxy URL for scraping. Use if Prometheus needs to access ZITADEL through a proxy. |
| metrics.serviceMonitor.relabellings | []RelabelConfig | `[]` | Relabeling rules applied before ingestion. Use to modify, filter, or drop labels before metrics are stored. |
| metrics.serviceMonitor.scheme | string | `nil` | HTTP scheme to use for scraping. Set to "https" if ZITADEL has internal TLS enabled. If null, defaults to "http". |
| metrics.serviceMonitor.scrapeInterval | string | `nil` | How often Prometheus should scrape metrics from ZITADEL. If null, uses Prometheus's default scrape interval (typically 30s). |
| metrics.serviceMonitor.scrapeTimeout | string | `nil` | Timeout for scrape requests. If null, uses Prometheus's default timeout. Should be less than scrapeInterval. |
| metrics.serviceMonitor.tlsConfig | TLSConfig | `{}` | TLS configuration for scraping HTTPS endpoints. Configure this if ZITADEL has internal TLS enabled and you need to verify certificates. |
| nameOverride | string | `""` | Override the "zitadel" portion of resource names. Useful when the default naming would conflict with existing resources or when deploying multiple instances with different configurations. |
| nodeSelector | map[string]string | `{}` | Node labels for pod assignment. Pods will only be scheduled on nodes that have all the specified labels. Use this to target specific node pools (e.g., high-memory nodes, nodes in specific zones). Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ |
| pdb.annotations | map[string]string | `{}` | Additional annotations to apply to the Pod Disruption Budget resource. |
| pdb.enabled | bool | `false` | Enable or disable the Pod Disruption Budget for ZITADEL pods |
| pdb.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailable during disruptions. Cannot be used together with minAvailable. |
| pdb.minAvailable | string | `nil` | Minimum number of pods that must remain available during disruptions. Cannot be used together with maxUnavailable. |
| podAdditionalLabels | map[string]string | `{}` | Additional labels to add to ZITADEL pods beyond the standard Helm labels. Useful for organizing pods with custom label selectors, network policies, or pod security policies. |
| podAnnotations | map[string]string | `{}` | Annotations to add to ZITADEL pods. Use this for integrations like Prometheus scraping, Istio sidecar injection, Vault agent injection, or any other annotation-based configuration. Example:   podAnnotations:     prometheus.io/scrape: "true"     sidecar.istio.io/inject: "true" |
| podSecurityContext.fsGroup | int | `1000` | Group ID for volume ownership and file creation. Files created in mounted volumes will be owned by this group. |
| podSecurityContext.runAsNonRoot | bool | `true` | Require containers to run as a non-root user. This is a security best practice that prevents privilege escalation attacks. |
| podSecurityContext.runAsUser | int | `1000` | User ID to run the container processes. 1000 is a common non-root UID that matches the ZITADEL container's default user. |
| readinessProbe.enabled | bool | `true` | Enable or disable the readiness probe. |
| readinessProbe.failureThreshold | int | `3` | Number of consecutive failures before marking the pod as not ready. |
| readinessProbe.initialDelaySeconds | int | `0` | Seconds to wait before starting readiness checks after container start. Set higher if ZITADEL needs time to initialize before accepting traffic. |
| readinessProbe.periodSeconds | int | `5` | How often (in seconds) to perform the readiness check. |
| replicaCount | int | `1` | Number of ZITADEL pod replicas. While a single replica is fine for testing, production environments should use 3 or more to prevent downtime during rolling updates, node failures, and ensure high availability. |
| resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for the ZITADEL container. Setting appropriate resources ensures predictable performance and prevents resource starvation. Requests affect scheduling; limits enforce caps. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| securityContext.privileged | bool | `false` | Prevent the container from running in privileged mode. Privileged containers have access to all host devices and capabilities. |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount the container's root filesystem as read-only. This prevents malicious processes from writing to the filesystem and is a security best practice for immutable containers. |
| securityContext.runAsNonRoot | bool | `true` | Require the container to run as a non-root user. |
| securityContext.runAsUser | int | `1000` | User ID to run the container process. |
| service.annotations | map[string]string | `{"traefik.ingress.kubernetes.io/service.serversscheme":"h2c"}` | Annotations to add to the Service resource. The default annotation tells Traefik to use HTTP/2 when communicating with ZITADEL backends. |
| service.appProtocol | string | `"kubernetes.io/h2c"` | Application protocol hint for ingress controllers and service meshes. "kubernetes.io/h2c" indicates HTTP/2 over cleartext (without TLS). |
| service.clusterIP | string | `""` | Fixed cluster IP address for ClusterIP services. Leave empty for automatic assignment by Kubernetes. Only applicable when type is "ClusterIP". Setting a fixed IP is useful when other services need a stable endpoint. |
| service.externalTrafficPolicy | string | `""` | Traffic policy for LoadBalancer and NodePort services. "Cluster" distributes traffic across all nodes (default), "Local" only routes to nodes with pods, preserving client source IP but potentially causing uneven load distribution. |
| service.labels | map[string]string | `{}` | Labels to add to the Service resource. Use for organizing services or matching service selectors in network policies. |
| service.port | int | `8080` | Port number the service exposes. Clients and ingress controllers connect to this port. ZITADEL uses 8080 by default for its HTTP/2 server. |
| service.protocol | string | `"http2"` | Protocol identifier used in port naming. ZITADEL uses HTTP/2 (h2c) for both REST API and gRPC traffic on the same port. |
| service.scheme | string | `"HTTP"` | HTTP scheme for health checks and internal communication. Use "HTTP" when TLS termination happens at the ingress/load balancer, or "HTTPS" when ZITADEL is configured with internal TLS. |
| service.type | string | `"ClusterIP"` | Service type. Use "ClusterIP" for internal-only access (typical when using an ingress controller), "NodePort" for direct node-level access, or "LoadBalancer" for cloud provider load balancers. |
| serviceAccount.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations to add to the service account. The default Helm hooks ensure the service account exists before pods are created. Add annotations here for cloud provider integrations (e.g., AWS IAM roles, GCP Workload Identity). |
| serviceAccount.create | bool | `true` | Whether to create a dedicated service account for ZITADEL. Set to false if you want to use an existing service account or the default account. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. Set this to use a specific existing service account when create is false. |
| setupJob.activeDeadlineSeconds | int | `300` | Maximum time in seconds for the setup job to complete. The job is terminated if it exceeds this deadline. Increase this for slow environments. |
| setupJob.additionalArgs | []string | `["--init-projections=true"]` | Additional command-line arguments to pass to the ZITADEL setup command. The default enables projection initialization for better startup performance. |
| setupJob.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"2"}` | Annotations for the setup job. The Helm hooks ensure this job runs after the init job (weight "2" > "1") and before the main deployment. |
| setupJob.backoffLimit | int | `5` | Number of retries before marking the setup job as failed. |
| setupJob.extraContainers | []Container | `[]` | Sidecar containers to run alongside the setup container. Useful for logging, proxies (e.g., cloud-sql-proxy), or other supporting services. |
| setupJob.initContainers | []Container | `[]` | Init containers to run before the main setup container. Useful for waiting on additional dependencies or performing pre-setup tasks. |
| setupJob.machinekeyWriter.image.repository | string | `""` | Override the default kubectl image repository. Leave empty to use the value from tools.kubectl.image.repository. |
| setupJob.machinekeyWriter.image.tag | string | `""` | Override the default kubectl image tag. Leave empty to use the value from tools.kubectl.image.tag (which defaults to cluster version). |
| setupJob.machinekeyWriter.resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for the machinekey writer container. This container only runs kubectl commands and needs minimal resources. |
| setupJob.podAdditionalLabels | map[string]string | `{}` | Additional labels to add to setup job pods. |
| setupJob.podAnnotations | map[string]string | `{}` | Additional annotations to add to setup job pods. |
| setupJob.resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for the setup job container. The setup job performs more work than init, including generating keys and creating initial data. |
| startupProbe.enabled | bool | `true` | Enable or disable the startup probe. When enabled, liveness and readiness probes are disabled until the startup probe succeeds. |
| startupProbe.failureThreshold | int | `30` | Number of consecutive failures before marking startup as failed and restarting the container. With periodSeconds=1 and failureThreshold=30, the container has 30 seconds to start. |
| startupProbe.periodSeconds | int | `1` | How often (in seconds) to perform the startup check. |
| tolerations | []Toleration | `[]` | Tolerations allow pods to be scheduled on nodes with matching taints. Taints are used to repel pods from nodes; tolerations allow exceptions. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |
| tools.kubectl.image.pullPolicy | string | `""` | The pull policy for the kubectl image. If left empty, Kubernetes applies its default policy depending on whether the tag is mutable or fixed. |
| tools.kubectl.image.repository | string | `"alpine/k8s"` | The name of the image repository that contains the kubectl image. The chart automatically prepends the registry (docker.io by default) for compatibility with CRI-O v1.34+ which enforces fully qualified names. |
| tools.kubectl.image.tag | string | `""` | The image tag to use for the kubectl image. It should be left empty to automatically default to the Kubernetes cluster version |
| tools.wait4x.image.pullPolicy | string | `""` | The pull policy for the wait4x image. If left empty, the chart defaults to the Kubernetes default pull policy for the given tag. |
| tools.wait4x.image.repository | string | `"wait4x/wait4x"` | The name of the image repository that contains the wait4x image. The chart automatically prepends the registry (docker.io by default) for compatibility with CRI-O v1.34+ which enforces fully qualified names. |
| tools.wait4x.image.tag | string | `"3.6"` | The image tag to use for the wait4x image. Leave empty to require the user to set a specific version explicitly. |
| tools.wait4x.resources | ResourceRequirements | `{}` | CPU and memory resource requests and limits for wait4x init containers. These resources apply to all init containers using the wait4x tool, including wait-for-zitadel and wait-for-postgres. Setting equal requests and limits enables the "Guaranteed" QoS class when combined with resource settings on the main container. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| topologySpreadConstraints | []TopologySpreadConstraint | `[]` | Topology spread constraints control how pods are distributed across topology domains (e.g., zones, nodes, regions) for high availability. Unlike affinity, these constraints provide more granular control over pod distribution. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/ |
| zitadel.autoscaling.annotations | map[string]string | `{}` | Annotations applied to the HPA object. |
| zitadel.autoscaling.behavior | HorizontalPodAutoscalerBehavior | `{}` | Configures the scaling behavior for scaling up and down. See: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior |
| zitadel.autoscaling.enabled | bool | `false` | If true, enables the Horizontal Pod Autoscaler for the Zitadel deployment. This will automatically override the `replicaCount` value. |
| zitadel.autoscaling.maxReplicas | int | `10` | The maximum number of pod replicas. |
| zitadel.autoscaling.metrics | []MetricSpec | `[]` | Advanced scaling based on custom metrics exposed by Zitadel. Zitadel exposes standard Go runtime metrics. To use these for scaling, you MUST have a metrics server (e.g., Prometheus) and a metrics adapter (e.g., prometheus-adapter) running in your cluster. Ref: https://github.com/kubernetes-sigs/prometheus-adapter |
| zitadel.autoscaling.minReplicas | int | `3` | The minimum number of pod replicas. |
| zitadel.autoscaling.targetCPU | string | `nil` | The target average CPU utilization percentage. |
| zitadel.autoscaling.targetMemory | string | `nil` | The target average memory utilization percentage. |
| zitadel.configSecretKey | string | `"config-yaml"` | The key within the configSecretName secret that contains the ZITADEL configuration YAML. The default "config-yaml" matches the expected format. |
| zitadel.configSecretName | string | `nil` | Name of an existing Kubernetes Secret containing ZITADEL configuration. Use this when you want to manage ZITADEL configuration externally (e.g., via External Secrets Operator, Sealed Secrets, or GitOps). The secret should contain YAML configuration in the same format as configmapConfig. |
| zitadel.configmapConfig | object | `{"Database":{"Postgres":{"Host":"","Port":5432}},"ExternalDomain":"","ExternalSecure":true,"FirstInstance":{"LoginClientPatPath":null,"MachineKeyPath":null,"Org":{"LoginClient":{"Machine":{"Name":"Automatically Initialized IAM Login Client","Username":"login-client"},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Machine":{"Machine":{"Name":"Automatically Initialized IAM Admin","Username":"iam-admin"},"MachineKey":{"ExpirationDate":"2029-01-01T00:00:00Z","Type":1},"Pat":{"ExpirationDate":"2029-01-01T00:00:00Z"}},"Skip":null},"PatPath":null,"Skip":false},"Machine":{"Identification":{"Hostname":{"Enabled":true},"Webhook":{"Enabled":false}}},"TLS":{"Enabled":false}}` | ZITADEL runtime configuration written to a Kubernetes ConfigMap. These values are passed directly to the ZITADEL binary and control its behavior. For the complete list of available configuration options, see: https://github.com/zitadel/zitadel/blob/main/cmd/defaults.yaml |
| zitadel.dbSslAdminCrtSecret | string | `""` | Name of a Kubernetes Secret containing the admin user's client certificate for mutual TLS (mTLS) authentication to the database. The secret must contain keys "tls.crt" (certificate) and "tls.key" (private key). Used by the init job for database setup operations that require elevated privileges. |
| zitadel.dbSslCaCrt | string | `""` | PEM-encoded CA certificate for verifying the database server's TLS certificate. Use this when your PostgreSQL/CockroachDB server uses a self-signed certificate or a certificate signed by a private CA. The certificate is stored in a Kubernetes Secret and mounted into ZITADEL pods at /db-ssl-ca-crt/ca.crt. Either provide the certificate inline here, or reference an existing secret using dbSslCaCrtSecret instead. |
| zitadel.dbSslCaCrtAnnotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the dbSslCaCrt Secret when created from the inline certificate. The default Helm hooks ensure the secret exists before pods start. |
| zitadel.dbSslCaCrtSecret | string | `""` | Name of an existing Kubernetes Secret containing the database CA certificate at key "ca.crt". Use this instead of dbSslCaCrt when the certificate is managed externally (e.g., by cert-manager or an operator). The secret must exist in the same namespace as the ZITADEL release. |
| zitadel.dbSslUserCrtSecret | string | `""` | Name of a Kubernetes Secret containing the application user's client certificate for mutual TLS (mTLS) authentication to the database. The secret must contain keys "tls.crt" (certificate) and "tls.key" (private key). Used by the main ZITADEL deployment and setup job for normal database operations. |
| zitadel.debug.annotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-weight":"1"}` | Annotations for the debug pod. The Helm hooks ensure it's created during install/upgrade and cleaned up appropriately. |
| zitadel.debug.enabled | bool | `false` | Enable or disable the debug pod. Only enable for troubleshooting; disable in production environments. |
| zitadel.debug.extraContainers | []Container | `[]` | Sidecar containers to run alongside the debug container. |
| zitadel.debug.initContainers | []Container | `[]` | Init containers to run before the debug container starts. |
| zitadel.extraContainers | []Container | `[]` | Global sidecar containers added to all ZITADEL workloads (Deployment, init job, setup job, and debug pod when enabled). Use this for shared services like database proxies (e.g., cloud-sql-proxy) that all workloads need to connect to the database. |
| zitadel.initContainers | []Container | `[]` | Global init containers added to all ZITADEL workloads (Deployment, init job, setup job, and debug pod when enabled). Use this for shared dependencies like database readiness checks or certificate initialization that all workloads need. |
| zitadel.masterkey | string | `""` | Zitadel's masterkey for symmetric encryption of sensitive data like private keys and tokens. Must be exactly 32 ASCII characters (not bytes). Do NOT use a hex-encoded or base64-encoded key, as these result in 64 or 44 characters respectively. Use literal characters only. Generate with: tr -dc A-Za-z0-9 </dev/urandom | head -c 32 IMPORTANT: Store this value securely. Loss of the masterkey means loss of all encrypted data. Either set this value or use masterkeySecretName. |
| zitadel.masterkeyAnnotations | map[string]string | `{"helm.sh/hook":"pre-install","helm.sh/hook-weight":"0"}` | Annotations for the masterkey Secret when created from zitadel.masterkey. The secret is created once on install and is immutable. |
| zitadel.masterkeySecretName | string | `""` | Name of an existing Kubernetes Secret containing the masterkey at key "masterkey". Use this for production deployments to avoid storing the masterkey in values files. The secret must exist before chart installation. Note: Either zitadel.masterkey or zitadel.masterkeySecretName must be set. |
| zitadel.podSecurityContext | PodSecurityContext | `{}` | Optional overrides for the pod security context used by Zitadel pods. If left empty, the chart-wide podSecurityContext defined below is used. |
| zitadel.revisionHistoryLimit | int | `10` | Number of old ReplicaSets to retain for rollback purposes Set to 0 to not keep any old ReplicaSets |
| zitadel.secretConfig | string | `nil` | Sensitive ZITADEL configuration values written to a Kubernetes Secret instead of a ConfigMap. Use this for database passwords, API keys, SMTP credentials, and other values that should not be stored in plain text. The secret is mounted alongside the ConfigMap and both are merged by ZITADEL at startup. Structure follows the same format as configmapConfig. See all options: https://github.com/zitadel/zitadel/blob/main/cmd/defaults.yaml Example:   secretConfig:     Database:       Postgres:         User:           Password: "my-secure-password" |
| zitadel.secretConfigAnnotations | map[string]string | `{"helm.sh/hook":"pre-install,pre-upgrade","helm.sh/hook-delete-policy":"before-hook-creation","helm.sh/hook-weight":"0"}` | Annotations for the secretConfig Secret. The default Helm hooks ensure the secret is created before the deployment and recreated on upgrades. |
| zitadel.securityContext | SecurityContext | `{}` | Optional overrides for the container security context used by Zitadel pods. If left empty, the chart-wide securityContext defined below is used. |
| zitadel.selfSignedCert.additionalDnsName | string | `""` |  |
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
