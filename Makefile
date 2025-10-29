.PHONY: all
all: test

.PHONY: test
test:
	@echo "Running Go tests..."
	go test ./...

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
