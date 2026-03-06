# Quickstart — Full Stack in One Command

Deploy ZITADEL with a bundled PostgreSQL database and Traefik ingress controller
on any Kubernetes cluster. No separate installs. One `helm install`.

> [!WARNING]
> This configuration is for **local development and evaluation only**.
> Credentials are hardcoded, TLS is disabled, and PostgreSQL data is not persisted.
> For production deployments see the [Kubernetes guide](https://zitadel.com/docs/self-hosting/deploy/kubernetes).

## Prerequisites

- A Kubernetes cluster (1.30+)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/) 3.x or 4.x

> **No cluster yet?** [k3d](https://k3d.io/) is a quick local option — it runs k3s in Docker:
> ```bash
> k3d cluster create zitadel --port "8080:80@loadbalancer"
> ```
> The default values work with k3d out of the box.

## Steps

```bash
mkdir zitadel-helm && cd zitadel-helm &&
curl -fsSLO https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/0-quickstart/quickstart-values.yaml
```

```bash
helm repo add zitadel https://charts.zitadel.com &&
helm repo add bitnami https://charts.bitnami.com/bitnami &&
helm repo add traefik https://traefik.github.io/charts &&
helm repo update &&
helm upgrade --install zitadel zitadel/zitadel --values quickstart-values.yaml --wait
```

That's it. Visit [http://localhost:8080/ui/console](http://localhost:8080/ui/console) and log in
with username `zitadel-admin@zitadel.localhost` and password `Password1!`.

## What just deployed?

```
Traefik (bundled) → ZITADEL API (Go) + ZITADEL Login (Next.js) → PostgreSQL (bundled)
```

All components are managed by a single Helm release:

```bash
kubectl get pods
helm status zitadel
```

The bundled Traefik uses `isDefaultClass: false`, so it only handles ZITADEL's own
Ingresses and does not conflict with any existing ingress controller in the cluster.

## What's next?

| Goal | How |
|------|-----|
| Bring your own database | Set `postgresql.enabled: false` and configure `zitadel.configmapConfig.Database` |
| Bring your own ingress controller | Set `traefik.enabled: false` and set `ingress.className` |
| Restrict Traefik to this namespace | Set `traefik.providers.kubernetesIngress.namespaces: ["<ns>"]` |
| Add Redis caching | Configure `zitadel.configmapConfig.Caches` |
| Go to production | Follow the [Production Checklist](https://zitadel.com/docs/self-hosting/manage/productionchecklist) |
