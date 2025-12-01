.PHONY: all build test test-unit test-integration clean docker-build docker-push manifests fmt vet lint

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
REGISTRY ?= ghcr.io/wizzz
IMG ?= $(REGISTRY)/pangolin-ingress-controller:$(VERSION)
GOFLAGS ?= -trimpath

all: build

##@ Development

fmt: ## Run go fmt
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

##@ Testing

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test ./internal/... -v -coverprofile=coverage.out

test-integration: ## Run integration tests with envtest
	go test ./tests/integration/... -v

coverage: test-unit ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html

##@ Build

build: fmt vet ## Build manager binary
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/manager ./cmd/manager

run: ## Run against the configured cluster
	go run ./cmd/manager

##@ Docker

docker-build: ## Build docker image
	docker build -t $(IMG) .

docker-push: docker-build ## Push docker image
	docker push $(IMG)

##@ Deployment

manifests: ## Generate deployment manifests
	mkdir -p deploy
	cat config/rbac/service_account.yaml > deploy/install.yaml
	echo "---" >> deploy/install.yaml
	cat config/rbac/role.yaml >> deploy/install.yaml
	echo "---" >> deploy/install.yaml
	cat config/rbac/role_binding.yaml >> deploy/install.yaml
	echo "---" >> deploy/install.yaml
	cat config/manager/deployment.yaml >> deploy/install.yaml

install: manifests ## Install into cluster
	kubectl apply -f deploy/install.yaml

uninstall: ## Uninstall from cluster
	kubectl delete -f deploy/install.yaml

##@ Cleanup

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
