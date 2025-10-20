.PHONY: all
all: test

.PHONY: test
test:
	@echo "Running Go tests..."
	@cd ./charts/zitadel/acceptance_test/ && go test -v -p 1 -timeout 30m

# Validate Helm chart manifests using kubeconform. This renders the Helm chart
# templates into Kubernetes YAML manifests and validates them against the K8s
# API schemas. The masterkey is required for template rendering but is just a
# placeholder value since we're only validating YAML structure, not deploying.
.PHONY: check
check:
	@echo "Validating Helm charts..."
	@helm template zitadel charts/zitadel \
		--values charts/zitadel/values.yaml \
		--set zitadel.masterkey="dGVzdC1tYXN0ZXJrZXktZm9yLXZhbGlkYXRpb24=" \
		| kubeconform -strict -ignore-missing-schemas -summary

.PHONY: start stop
start stop: NS ?= $(shell basename "$$(git rev-parse --show-toplevel 2>/dev/null || pwd)" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9-]+/-/g;s/^-+//;s/-+$$//;s/(.{1,63}).*/\1/')
start stop: POSTGRES_RELEASE ?= db
start stop: ZITADEL_RELEASE ?= zitadel
start stop: CHART_PATH ?= charts/zitadel
start stop: PG_CHART ?= bitnami/postgresql
start stop: PG_VERSION ?= 15.2.10

# Local ports:
# - 8443: HTTPS to Zitadel (TLS terminated by Zitadel itself using chart self-signed cert)
# - 3000: HTTPS to Login UI (TLS terminated by NGINX sidecar; Next.js still listens on 3000 internally)
ZITADEL_LOCAL_HTTPS ?= 8443
LOGIN_LOCAL_PORT ?= 8444

KUBEFLAGS :=
HELMFLAGS :=
ifneq ($(strip $(KUBECONFIG)),)
KUBEFLAGS += --kubeconfig=$(KUBECONFIG)
HELMFLAGS += --kubeconfig=$(KUBECONFIG)
endif
ifneq ($(strip $(KCTX)),)
KUBEFLAGS += --context=$(KCTX)
HELMFLAGS += --kube-context=$(KCTX)
endif

start stop: kubectl-ns = kubectl $(KUBEFLAGS) --namespace=$(NS)
start stop: helm-ns    = helm $(HELMFLAGS) --namespace=$(NS)

start:
	@echo "==> verifying kubectl access"
	@if ! kubectl $(KUBEFLAGS) cluster-info >/dev/null 2>&1; then echo "kubectl cannot reach a cluster (set KUBECONFIG or KCTX)"; exit 1; fi
	@echo "==> ensuring local ports are free ($(ZITADEL_LOCAL_HTTPS), $(LOGIN_LOCAL_PORT))"
	@# free $(ZITADEL_LOCAL_HTTPS) if our own kubectl PF has it; otherwise fail
	@if lsof -iTCP:$(ZITADEL_LOCAL_HTTPS) -sTCP:LISTEN -Pn >/dev/null 2>&1; then \
		if ps -o command= -p $$(lsof -tiTCP:$(ZITADEL_LOCAL_HTTPS) -sTCP:LISTEN) | grep -qE "kubectl .*port-forward"; then \
			pkill -f "kubectl .*(--namespace=$(NS)|-n $(NS)).*port-forward" >/dev/null 2>&1 || true; \
			sleep 1; \
			lsof -iTCP:$(ZITADEL_LOCAL_HTTPS) -sTCP:LISTEN -Pn >/dev/null 2>&1 && { echo "port $(ZITADEL_LOCAL_HTTPS) still busy"; exit 1; } || true; \
		else \
			echo "port $(ZITADEL_LOCAL_HTTPS) in use by another process; free it or set a different port"; exit 1; \
		fi; \
	fi
	@# free $(LOGIN_LOCAL_PORT) if our own kubectl PF has it; otherwise fail
	@if lsof -iTCP:$(LOGIN_LOCAL_PORT) -sTCP:LISTEN -Pn >/dev/null 2>&1; then \
		if ps -o command= -p $$(lsof -tiTCP:$(LOGIN_LOCAL_PORT) -sTCP:LISTEN) | grep -qE "kubectl .*port-forward"; then \
			pkill -f "kubectl .*(--namespace=$(NS)|-n $(NS)).*port-forward" >/dev/null 2>&1 || true; \
			sleep 1; \
			lsof -iTCP:$(LOGIN_LOCAL_PORT) -sTCP:LISTEN -Pn >/dev/null 2>&1 && { echo "port $(LOGIN_LOCAL_PORT) still busy"; exit 1; } || true; \
		else \
			echo "port $(LOGIN_LOCAL_PORT) in use by another process; free it or set a different port"; exit 1; \
		fi; \
	fi
	@echo "==> creating namespace '$(NS)'"
	@if kubectl $(KUBEFLAGS) get namespace "$(NS)" >/dev/null 2>&1; then echo "namespace '$(NS)' already exists"; exit 1; fi
	@kubectl $(KUBEFLAGS) create namespace "$(NS)"
	@echo "==> installing Postgres ($(PG_CHART) $(PG_VERSION))"
	@helm repo add bitnami https://charts.bitnami.com/bitnami >/dev/null 2>&1 || true
	@helm repo update >/dev/null
	$(helm-ns) upgrade --install $(POSTGRES_RELEASE) $(PG_CHART) \
		--version=$(PG_VERSION) \
		--set=primary.persistence.enabled=false \
		--set-string=primary.pgHbaConfiguration="host all all all trust" \
		--set=image.repository=bitnamilegacy/postgresql \
		--set=metrics.image.repository=bitnamilegacy/postgres-exporter \
		--set=volumePermissions.image.repository=bitnamilegacy/os-shell \
		--wait --atomic --timeout=5m --hide-notes
	@echo "==> installing ZITADEL chart with self-signed TLS (no ingress)"
	$(helm-ns) upgrade --install $(ZITADEL_RELEASE) $(CHART_PATH) \
		--set=zitadel.masterkey=x123456789012345678901234567891y \
		--set=zitadel.configmapConfig.ExternalSecure=true \
		--set=zitadel.configmapConfig.ExternalDomain=$(NS).dev.mrida.ng \
		--set=zitadel.configmapConfig.ExternalPort=$(ZITADEL_LOCAL_HTTPS) \
		--set=zitadel.configmapConfig.TLS.Enabled=true \
		--set=zitadel.selfSignedCert.enabled=true \
		--set=zitadel.configmapConfig.Database.Postgres.Host=$(POSTGRES_RELEASE)-postgresql \
		--set=zitadel.configmapConfig.Database.Postgres.Port=5432 \
		--set=zitadel.configmapConfig.Database.Postgres.Database=zitadel \
		--set=zitadel.configmapConfig.Database.Postgres.MaxOpenConns=20 \
		--set=zitadel.configmapConfig.Database.Postgres.MaxIdleConns=10 \
		--set=zitadel.configmapConfig.Database.Postgres.MaxConnLifetime=30m \
		--set=zitadel.configmapConfig.Database.Postgres.MaxConnIdleTime=5m \
		--set=zitadel.configmapConfig.Database.Postgres.User.Username=postgres \
		--set=zitadel.configmapConfig.Database.Postgres.User.SSL.Mode=disable \
		--set=zitadel.configmapConfig.Database.Postgres.Admin.Username=postgres \
		--set=zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode=disable \
		--set=login.enabled=true \
		--set=login.selfSignedCert.enabled=true \
		--set=login.service.port=$(LOGIN_LOCAL_PORT) \
		--wait --atomic --timeout=5m --hide-notes
	@echo "==> waiting for deployments to be available"
	@kubectl $(KUBEFLAGS) -n "$(NS)" wait --for=condition=available --timeout=300s deployment/$(ZITADEL_RELEASE)
	@kubectl $(KUBEFLAGS) -n "$(NS)" wait --for=condition=available --timeout=300s deployment/$(ZITADEL_RELEASE)-login
	@echo "==> starting local port-forwards"
	@{ lsof -iTCP:$(ZITADEL_LOCAL_HTTPS) -sTCP:LISTEN -Pn >/dev/null 2>&1 && echo "port $(ZITADEL_LOCAL_HTTPS) in use, skipping PF"; } || nohup $(kubectl-ns) port-forward svc/$(ZITADEL_RELEASE) $(ZITADEL_LOCAL_HTTPS):8080 >/dev/null 2>&1 &
	@{ lsof -iTCP:$(LOGIN_LOCAL_PORT) -sTCP:LISTEN -Pn >/dev/null 2>&1 && echo "port $(LOGIN_LOCAL_PORT) in use, skipping PF"; } || nohup $(kubectl-ns) port-forward svc/$(ZITADEL_RELEASE)-login $(LOGIN_LOCAL_PORT):$(LOGIN_LOCAL_PORT) >/dev/null 2>&1 &
	@sleep 1
	@echo "==> URLs"
	@echo "    1. Go to the API/Console URL: https://$(NS).dev.mrida.ng:$(ZITADEL_LOCAL_HTTPS)/"
	@echo "       Username: zitadel-admin@zitadel.$(NS).dev.mrida.ng"
	@echo "       Password: Password1!"
	@echo "    Login UI:    https://localhost:$(LOGIN_LOCAL_PORT)/ui/v2/login"
	@kubectl $(KUBEFLAGS) -n "$(NS)" get secret iam-admin -o jsonpath='{.data.iam-admin\.json}' 2>/dev/null | base64 -D 2>/dev/null > sa.json || kubectl $(KUBEFLAGS) -n "$(NS)" get secret iam-admin -o jsonpath='{.data.iam-admin\.json}' 2>/dev/null | base64 -d > sa.json || true
	@echo "    Login client PAT (secret 'login-client'):"
	@PAT=$$(kubectl $(KUBEFLAGS) -n "$(NS)" get secret login-client -o jsonpath='{.data.pat}' 2>/dev/null | base64 -d 2>/dev/null || kubectl $(KUBEFLAGS) -n "$(NS)" get secret login-client -o jsonpath='{.data.pat}' 2>/dev/null | base64 -D 2>/dev/null); \
	if [ -n "$$PAT" ]; then echo "      $$PAT"; else echo "      (not found yet)"; fi

stop:
	@echo "==> stopping port-forwards (if any)"
	-@pkill -f "kubectl .*(--namespace=$(NS)|-n $(NS)).*port-forward" >/dev/null 2>&1 || true
	@echo "==> deleting namespace '$(NS)'"
	@kubectl $(KUBEFLAGS) delete namespace "$(NS)" --ignore-not-found --wait=true
