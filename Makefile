.PHONY: all
all: test

.PHONY: test
test:
	@echo "Running Go tests..."
	@cd ./charts/zitadel/acceptance_test/ && go test -v -p 1 -timeout 30m
