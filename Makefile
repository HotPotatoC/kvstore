BINARY_DIR := .bin/

APP_NAME=kvstore
VERSION?=v0.0.3
BUILD=$(shell git rev-parse HEAD)

PLATFORMS=freebsd darwin linux windows
ARCHS=386 arm arm64 amd64

LDFLAGS = -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Build=${BUILD}'"

.DEFAULT_GOAL := help

# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Displays all the available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format go files
	@go fmt ./...

vet: ## go vet
	@go vet ./...

clean: ## Deletes all compiled / executable files
	@find .bin -type f -name '*' -print0 | xargs -0 rm --
	@echo ">> Deleted all build files!"

install: ## Installs the package
	@go install ${LDFLAGS} ./...

install-deps: ## Install dependencies
	@go mod download

server: ## Compile the server
	@go build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_server cmd/$(APP_NAME)_server/main.go

cli: ## Compile the cli
	@go build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_cli cmd/$(APP_NAME)_cli/main.go

all-server: ## Cross-compile the server
	@$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHS),\
			$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH);\
			go build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_server-$(GOOS)-$(GOARCH) cmd/$(APP_NAME)_server/main.go)))

all-cli: ## Cross-compile the cli
	@$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHS),\
			$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH);\
			go build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_cli-$(GOOS)-$(GOARCH) cmd/$(APP_NAME)_cli/main.go)))

all: ## Cross-compile all the commands
	@echo ">> Building go files..."
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) all-server
	@$(MAKE) all-cli
	@echo ">> Finished building"
	@ls -hl -d $(BINARY_DIR)* $(BINARY_DIR)

.PHONY: help install-deps fmt vet clean server cli all-server all-cli all