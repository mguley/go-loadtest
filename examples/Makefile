# NATS example Makefile
include .envrc

## help: Print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# =============================================================================== #
# LINTER
# =============================================================================== #

## install/linter: Install GolangCI-Lint
.PHONY: install/linter
install/linter:
	@echo "Installing GolangCI-Lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

## lint: Run linter on all Go files in each module directory
.PHONY: lint
lint: install/linter
	@echo "Running GolangCI-Lint on all Go files..."
	@$(shell go env GOPATH)/bin/golangci-lint run ./nats/...

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@for service in nats; do \
		echo "Tidying and verifying $$service..."; \
		(cd $$service && go mod tidy && go mod verify); \
	done
	@echo "Vendoring workspace dependencies..."
	go work vendor

# =============================================================================== #
# LOAD TEST RUNNERS
# =============================================================================== #

## run/publish: Run NATS publish load tests.
.PHONY: run/publish
run/publish:
	LOAD_TEST_TYPE=publish go run ./nats/cmd

## run/subscribe: Run NATS subscribe load tests.
.PHONY: run/subscribe
run/subscribe:
	LOAD_TEST_TYPE=subscribe go run ./nats/cmd

## run/combined: Run combined NATS pub/sub load tests.
.PHONY: run/combined
run/combined:
	LOAD_TEST_TYPE=combined go run ./nats/cmd

## run/high-concurrency: Run with high concurrency (100 clients).
.PHONY: run/high-concurrency
run/high-concurrency:
	LOAD_TEST_CONCURRENCY=100 go run ./nats/cmd

## run/long-duration: Run extended duration test (5 minutes).
.PHONY: run/long-duration
run/long-duration:
	LOAD_TEST_DURATION=5m go run ./nats/cmd

# =============================================================================== #
# DOCKER
# =============================================================================== #

## docker/nats: Start a NATS server in Docker for testing
.PHONY: docker/nats
docker/nats:
	docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats
