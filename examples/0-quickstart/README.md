# Quickstart — Full Stack in One Command

Deploy ZITADEL with a bundled PostgreSQL database on any Kubernetes cluster.
No separate database install. No manual secret wiring. One `helm install`.

> [!WARNING]
> This configuration is for **local development and evaluation only**.
> Credentials are hardcoded, TLS is disabled, and PostgreSQL data is not persisted.
> For production deployments see the [Kubernetes guide](https://zitadel.com/docs/self-hosting/deploy/kubernetes).

## Prerequisites

- A Kubernetes cluster (1.30+) with an ingress controller
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/) 3.x or 4.x

> **No cluster yet?** [k3d](https://k3d.io/) is a quick local option — it runs k3s in Docker
> and includes Traefik as the ingress controller:
> ```bash
> k3d cluster create zitadel --port "8080:80@loadbalancer"
> ```

## Steps

### Install the full stack

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

> Adjust `ExternalDomain`, `ExternalPort`, and `ingress.className` in `quickstart-values.yaml`
> to match your cluster's domain/IP and ingress controller.

## What just deployed?

```
<your ingress controller> → ZITADEL API (Go) + ZITADEL Login (Next.js) → PostgreSQL
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
