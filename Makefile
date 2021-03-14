BINARY_DIR := .bin/

APP_NAME=kvstore
VERSION?=v0.6.1
BUILD=$(shell git rev-parse HEAD)

GO=go
GOOSS=darwin linux windows freebsd netbsd openbsd dragonfly
GOARCHS=386 arm arm64 amd64
LDFLAGS=-ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Build=${BUILD}'"

DOCKER_SERVER_REPO_NAME=hotpotatoc/kvstore_server
DOCKER_CLI_REPO_NAME=hotpotatoc/kvstore_cli

DOCKER_SERVER_IMG=${DOCKER_SERVER_REPO_NAME}:${VERSION}
DOCKER_SERVER_LATEST=${DOCKER_SERVER_REPO_NAME}:latest

DOCKER_CLI_IMG=${DOCKER_SERVER_REPO_NAME}:${VERSION}
DOCKER_CLI_LATEST=${DOCKER_SERVER_REPO_NAME}:latest

.DEFAULT_GOAL := help

.PHONY: help
# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Displays all the available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: fmt
fmt: ## Format go files
	@$(GO) fmt ./...

.PHONY: vet
vet: ## go vet
	@$(GO) vet ./...

.PHONY: clean
clean: ## Deletes all compiled / executable files
	@find .bin -type f -name '*' -print0 | xargs -0 rm --
	@echo ">> Deleted all build files!"

.PHONY: install
install: ## Installs the package
	@$(GO) install ${LDFLAGS} ./...

.PHONY: install-deps
install-deps: ## Install dependencies
	@$(GO) mod download

.PHONY: server
server: ## Compile the server
	@$(GO) build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_$(VERSION)_server cmd/$(APP_NAME)_server/main.go

.PHONY: cli
cli: ## Compile the cli
	@$(GO) build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_$(VERSION)_cli cmd/$(APP_NAME)_cli/main.go

.PHONY: all-server
all-server: ## Cross-compile the server
	@$(foreach GOOS, $(GOOSS),\
		$(foreach GOARCH, $(GOARCHS),\
			$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH);\
			$(GO) build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_$(VERSION)_server-$(GOOS)-$(GOARCH) cmd/$(APP_NAME)_server/main.go)))

.PHONY: all-cli
all-cli: ## Cross-compile the cli
	@$(foreach GOOS, $(GOOSS),\
		$(foreach GOARCH, $(GOARCHS),\
			$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH);\
			$(GO) build $(LDFLAGS) -v -o $(BINARY_DIR)/$(APP_NAME)_$(VERSION)_cli-$(GOOS)-$(GOARCH) cmd/$(APP_NAME)_cli/main.go)))

.PHONY: docker-server
docker-server: ## Builds the kvstore server docker image
	@docker build --rm -t $(DOCKER_SERVER_IMG) \
		-f build/package/container/kvstore_server/Dockerfile \
		--build-arg LDFLAGS=$(LDFLAGS) \
		--build-arg GIT_COMMIT=$(BUILD) \
		--build-arg VERSION=$(VERSION) \
		--no-cache .
	@docker tag $(DOCKER_SERVER_IMG) $(DOCKER_SERVER_LATEST)

.PHONY: docker-cli
docker-cli: ## Builds the kvstore cli app docker image
	@docker build --rm -t $(DOCKER_CLI_IMG) \
		-f build/package/container/kvstore_cli/Dockerfile \
		--build-arg LDFLAGS=$(LDFLAGS) \
		--build-arg GIT_COMMIT=$(BUILD) \
		--build-arg VERSION=$(VERSION) \
		--no-cache .
	@docker tag $(DOCKER_CLI_IMG) $(DOCKER_CLI_LATEST)

.PHONY: all
all: ## Cross-compile all the commands
	@echo ">> Building go files..."
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) all-server
	@$(MAKE) all-cli
	@echo ">> Finished building"
	@ls -hl -d $(BINARY_DIR)* $(BINARY_DIR)
