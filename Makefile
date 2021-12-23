.DEFAULT_GOAL := help

VERSION?=$(shell cat VERSION)
BUILD=$(shell git rev-parse HEAD)

VERSION_PACKAGE=github.com/HotPotatoC/kvstore/build
LDFLAGS=-X ${VERSION_PACKAGE}.Version=${VERSION} -X ${VERSION_PACKAGE}.Build=${BUILD}

GOPATH = $(shell go env GOPATH)

.PHONY: lint
lint: ## Lint the source code
	revive ./...

.PHONY: fmt
fmt: ## Format the source code
	go fmt ./...

.PHONY: vet
vet: ## Lint the source code
	go vet ./...

.PHONY: test
test: ## Run the tests
	go test -v ./...

.PHONY: bench
bench: ## Run the benchmarks
	go test -bench=. -benchmem ./...

.PHONY: coverage
coverage: ## Generate coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: install
install: ## Install the binary
	go install -ldflags "${LDFLAGS}" ./...

.PHONY: help
# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Displays all the available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)