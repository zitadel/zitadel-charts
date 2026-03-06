# Quickstart — Full Stack in One Command

Deploy ZITADEL with a bundled PostgreSQL database on a local Kubernetes cluster.
No separate database install. No manual secret wiring. One `helm install`.

> [!WARNING]
> This configuration is for **local development and evaluation only**.
> Credentials are hardcoded, TLS is disabled, and PostgreSQL data is not persisted.
> For production deployments see the [Kubernetes guide](https://zitadel.com/docs/self-hosting/deploy/kubernetes).

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) Engine 24+
- [k3d](https://k3d.io/#installation) 5+
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/) 3.x or 4.x

## Steps

### 1 — Create a local cluster

```bash
k3d cluster create zitadel --port "8080:80@loadbalancer"
```

### 2 — Install the full stack

```bash
helm repo add zitadel https://charts.zitadel.com
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install zitadel zitadel/zitadel \
  --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/0-quickstart/quickstart-values.yaml \
  --wait
```

That's it. Visit [http://localhost:8080/ui/console](http://localhost:8080/ui/console) and log in
with username `zitadel-admin@zitadel.localhost` and password `Password1!`.

## What just deployed?

```
Traefik (built into k3d) → ZITADEL API (Go) + ZITADEL Login (Next.js) → PostgreSQL
```

All three components are managed by a single Helm release. You can inspect them:

```bash
kubectl get pods
helm status zitadel
```

## What's next?

| Goal | How |
|------|-----|
| Bring your own database | Set `postgresql.enabled: false` and configure `zitadel.configmapConfig.Database` |
| Bring your own ingress controller | Set `ingress.className` to your controller (e.g., `nginx`) |
| Add Redis caching | Configure `zitadel.configmapConfig.Caches` |
| Go to production | Follow the [Production Checklist](https://zitadel.com/docs/self-hosting/manage/productionchecklist) |
