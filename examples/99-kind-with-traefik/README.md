# Local KinD Cluster with Traefik Ingress Controller

By running the commands below, you run a local Kubernetes cluster using Kubernetes in Docker.
You configure KinD so your local machines ports 80 and 443 are mapped to the KinD container and handled by a Traefik ingress controller.
Additionally, you deploy a docker registry, so you can easily run pods from your locally built images.

> [!WARNING]  
> This example is for development purposes only. 

```bash
# Run Kubernetes in Docker
curl -fsSL -o create-cluster.sh https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/99-kind-with-traefik/create-cluster.sh
chmod 700 create-cluster.sh
./create-cluster.sh

# Install Traefik
helm repo add traefik https://traefik.github.io/charts
helm install --namespace ingress --create-namespace --wait traefik traefik/traefik --version 36.3.0 --values https://raw.githubusercontent.com/zitadel/zitadel-charts/main/examples/99-kind-with-traefik/traefik-values.yaml
```

Now, you are fully set up to try one of the examples for installing the Zitadel chart on your local machine.
