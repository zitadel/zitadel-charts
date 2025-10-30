# ZITADEL with CloudNativePG and Traefik

This example demonstrates a complete ZITADEL deployment using CloudNativePG for PostgreSQL database management, Traefik as the ingress controller with TLS termination, and self-signed certificates generated via Helm.

#### What This Example Does

When you deploy this example, the following happens automatically:

First, Traefik is deployed as the ingress controller with a NodePort service that maps your local machine's ports 80 and 443 to the cluster. It also generates a self-signed wildcard certificate for `*.127.0.0.1.sslip.io` that will be used for HTTPS.

Next, the CloudNativePG operator is installed to manage PostgreSQL clusters. It then creates a single-instance PostgreSQL cluster with a database named `zitadel` owned by a user also named `zitadel`. A separate `postgres` superuser is also created with its own credentials.

Finally, ZITADEL is deployed and configured to connect to the PostgreSQL cluster. It runs its initialization and setup jobs to create the database schema and bootstrap the first organization with default admin credentials. The setup also creates machine users that can be used for API access.

Throughout this process, ZITADEL is configured to accept external HTTPS traffic through Traefik, which handles TLS termination. This means ZITADEL itself runs without TLS enabled, but users access it securely via HTTPS through the ingress controller.

## Prerequisites

You'll need a kind cluster with port mappings configured for 80→30080 and 443→30443. You'll also need helm (v3 or later) and helmfile installed on your machine.

## Deploy
```bash
cd examples/cloudnativepg
helmfile sync
```

If you're starting fresh or want to clean up an existing deployment first, run `helmfile destroy` before deploying.

**Note:** Deployment takes approximately 2-3 minutes. All components deploy in sequence: Traefik (15s), CloudNativePG operator (15s), PostgreSQL cluster (15s), and finally ZITADEL with its init, setup, and main deployment (60-90s).

## Access ZITADEL

Once deployment is complete, open your browser and navigate to https://zitadel.127.0.0.1.sslip.io

You'll see a security warning because the certificate is self-signed. This is expected for local development. Click "Advanced" and proceed to the site.

Log in using these default admin credentials:
- **Username:** `zitadel-admin@zitadel.zitadel.127.0.0.1.sslip.io`
- **Password:** `Password1!`

## Cleanup

When you're done and want to remove everything, simply run:
```bash
helmfile destroy
```

This will cleanly remove all deployed components including ZITADEL, the PostgreSQL cluster, CloudNativePG operator, and Traefik.
