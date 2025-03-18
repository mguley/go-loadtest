## help: Print this help message
.PHONY: help
help:
	@echo 'Usage':
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## install/linter: Install GolangCI-Lint
.PHONY: install/linter
install/linter:
	@echo "Installing GolangCI-Lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

## lint: Run linter on all Go files in each module directory
.PHONY: lint
lint: install/linter
	@echo "Running GolangCI-Lint on all Go files..."
	@$(shell go env GOPATH)/bin/golangci-lint run ./...

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@for service in pkg; do \
		echo "Tidying and verifying $$service..."; \
		(cd $$service && go mod tidy && go mod verify); \
	done
	@echo "Vendoring workspace dependencies..."
	go work vendor
