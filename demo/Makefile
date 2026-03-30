# Demo environment for ZITADEL on k3d with Traefik and self-signed TLS.
# Domain: zitadeldemo.dev.mrida.ng (loopback domain resolving to 127.0.0.1)
#
# Quick start:
#   make -C demo deploy
#
# Tear down:
#   make -C demo destroy

SHELL := /bin/bash
.DEFAULT_GOAL := deploy

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
CLUSTER_NAME    := zitadel-demo
K3S_IMAGE       := rancher/k3s:v1.31.4-k3s1
RELEASE_NAME    := zitadel
NAMESPACE       := default
DOMAIN          := zitadeldemo.dev.mrida.ng

# ---------------------------------------------------------------------------
# init: Create a k3d cluster with the built-in k3s Traefik.
#
# - Maps host ports 80 (HTTP) and 443 (HTTPS) to the k3d load balancer.
# - Mounts traefik-config.yaml into k3s manifests so Traefik picks up
#   our customizations (self-signed TLS, dashboard, HTTP→HTTPS redirect)
#   automatically on startup.
# - Idempotent: skips creation if the cluster already exists.
# ---------------------------------------------------------------------------
.PHONY: init
init:
	@if k3d cluster list -o json 2>/dev/null | grep -q '"name":"$(CLUSTER_NAME)"'; then \
		echo "Cluster '$(CLUSTER_NAME)' already exists — skipping creation."; \
	else \
		echo "Creating k3d cluster '$(CLUSTER_NAME)'..."; \
		k3d cluster create $(CLUSTER_NAME) \
			--image $(K3S_IMAGE) \
			--port "80:80@loadbalancer" \
			--port "443:443@loadbalancer" \
			--volume "$$(pwd)/traefik-config.yaml:/var/lib/rancher/k3s/server/manifests/traefik-config.yaml@server:0" \
			--wait; \
	fi

# ---------------------------------------------------------------------------
# install: Install or upgrade the ZITADEL chart from Artifact Hub.
#
# - Uses helm upgrade --install so it works for both first install and
#   subsequent upgrades.
# - 10-minute timeout accounts for PostgreSQL startup + init/setup hooks.
# ---------------------------------------------------------------------------
.PHONY: install
install:
	@helm repo add zitadel https://charts.zitadel.com 2>/dev/null || true
	@helm repo update zitadel
	@echo "Installing/upgrading ZITADEL..."
	@helm upgrade --install $(RELEASE_NAME) zitadel/zitadel \
		--namespace $(NAMESPACE) \
		--values zitadel-values.yaml \
		--wait --timeout 10m

# ---------------------------------------------------------------------------
# deploy: Full setup — create cluster, install ZITADEL.
#
# This is the default target. It is fully idempotent and safe to run
# multiple times.
# ---------------------------------------------------------------------------
.PHONY: deploy
deploy: init install
	@echo ""
	@echo "======================================================"
	@echo " ZITADEL is ready!"
	@echo ""
	@echo " Login at https://$(DOMAIN)/ui/console"
	@echo " with username zitadel-admin@zitadel.$(DOMAIN)"
	@echo " and password Password1!"
	@echo ""
	@echo " Your browser will warn about the self-signed"
	@echo " certificate — accept it to continue."
	@echo "======================================================"

# ---------------------------------------------------------------------------
# uninstall: Remove the ZITADEL release only.
#
# The k3d cluster and Traefik remain intact so you can re-run
# "make install" without recreating the cluster.
# ---------------------------------------------------------------------------
.PHONY: uninstall
uninstall:
	@echo "Uninstalling ZITADEL..."
	@helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE) --wait || true

# ---------------------------------------------------------------------------
# destroy: Tear down the entire k3d cluster.
#
# This removes everything — Traefik, ZITADEL, PostgreSQL, and the
# cluster itself. Use "make deploy" to start fresh.
# ---------------------------------------------------------------------------
.PHONY: destroy
destroy:
	@echo "Deleting k3d cluster '$(CLUSTER_NAME)'..."
	@k3d cluster delete $(CLUSTER_NAME) || true

# ---------------------------------------------------------------------------
# dashboard: Port-forward the Traefik dashboard to localhost:9000.
#
# Open http://localhost:9000/dashboard/ in your browser.
# Press Ctrl+C to stop the port-forward.
# ---------------------------------------------------------------------------
.PHONY: dashboard
dashboard:
	@echo "Traefik dashboard: http://localhost:9000/dashboard/"
	@echo "Press Ctrl+C to stop."
	@kubectl port-forward -n kube-system svc/traefik 9000:9000
