# ZITADEL with CloudNativePG, Traefik, and Self-Signed Certs

## Prerequisites

- kind cluster with port mappings: 80→30080, 443→30443
- helm
- helmfile

## Deploy
```bash
helmfile apply
```

## Access

URL: https://zitadel.127.0.0.1.sslip.io
Username: zitadel-admin@zitadel.zitadel.127.0.0.1.sslip.io
Password: Password1!

Accept the self-signed certificate warning in your browser.

## Cleanup
```bash
helmfile destroy
```

## Notes

- Self-signed cert generated via Helm's `genSelfSignedCert`
- Valid for *.127.0.0.1.sslip.io wildcard domain
- Certificate stored in `traefik-default-cert` secret in ingress namespace
- CloudNativePG creates `zitadel-pg-rw` service for read-write connections
