# ═══════════════════════════════════════════════════════════
#  GORDON WATCHER - MAKEFILE
# ═══════════════════════════════════════════════════════════

# Variables
APP_NAME := gordon-watcher
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-w -s \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)"

# Paths
BIN_DIR := bin
BUILD_DIR := build
COVERAGE_DIR := coverage
DOCKER_IMAGE := ghcr.io/fabyo/$(APP_NAME)

# Go settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: discover-ip
discover-ip: ## Discover WSL and Windows IPs for Samba access
	@echo "🔍 Descobrindo IPs..."
	@echo ""
	@WSL_IP=$$(ip addr show eth0 2>/dev/null | grep "inet " | awk '{print $$2}' | cut -d/ -f1); \
	WIN_IP=$$(cat /etc/resolv.conf 2>/dev/null | grep nameserver | awk '{print $$2}'); \
	echo "📍 IP do WSL (Docker): $$WSL_IP"; \
	echo "🪟 IP do Windows: $$WIN_IP"; \
	echo ""; \
	echo "Para conectar do Windows ao Samba:"; \
	echo "  \\\\$$WSL_IP\\incoming"; \
	echo ""; \
	echo "Credenciais:"; \
	echo "  Usuário: gordon"; \
	echo "  Senha: (verifique .env)"
	@echo ""

# ─────────────────────────────────────────────────────────
#  BUILD
# ─────────────────────────────────────────────────────────

.PHONY: build
build: clean ## Build the application
	@echo "$(COLOR_GREEN)Building $(APP_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		$(LDFLAGS) \
		-o $(BIN_DIR)/$(APP_NAME) \
		./cmd/watcher
	@echo "$(COLOR_GREEN)✓ Build complete: $(BIN_DIR)/$(APP_NAME)$(COLOR_RESET)"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "$(COLOR_GREEN)Building for all platforms...$(COLOR_RESET)"
	@$(MAKE) build GOOS=linux GOARCH=amd64
	@mv $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)-linux-amd64
	@$(MAKE) build GOOS=linux GOARCH=arm64
	@mv $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)-linux-arm64
	@$(MAKE) build GOOS=darwin GOARCH=amd64
	@mv $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)-darwin-amd64
	@$(MAKE) build GOOS=darwin GOARCH=arm64
	@mv $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)-darwin-arm64
	@$(MAKE) build GOOS=windows GOARCH=amd64
	@mv $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe
	@echo "$(COLOR_GREEN)✓ All builds complete$(COLOR_RESET)"

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@echo "$(COLOR_GREEN)Installing $(APP_NAME)...$(COLOR_RESET)"
	@cp $(BIN_DIR)/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME)
	@echo "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(APP_NAME)$(COLOR_RESET)"

# ─────────────────────────────────────────────────────────
#  DEPENDENCIES
# ─────────────────────────────────────────────────────────

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	@go mod download
	@go mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(COLOR_GREEN)Updating dependencies...$(COLOR_RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: deps-clean
deps-clean: ## Clean up dependencies
	@echo "$(COLOR_GREEN)Cleaning dependencies...$(COLOR_RESET)"
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies cleaned$(COLOR_RESET)"

# ─────────────────────────────────────────────────────────
#  TESTING
# ─────────────────────────────────────────────────────────

.PHONY: test
test: ## Run unit tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@go test -v -race -timeout 30s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(COLOR_RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(COLOR_GREEN)Running integration tests...$(COLOR_RESET)"
	@go test -v -race -tags=integration ./tests/integration/...

.PHONY: test-bench
test-bench: ## Run benchmarks
	@echo "$(COLOR_GREEN)Running benchmarks...$(COLOR_RESET)"
	@go test -bench=. -benchmem -run=^$$ ./...

.PHONY: test-stress
test-stress: ## Run stress tests (file flood simulation)
	@echo "$(COLOR_GREEN)Running stress tests...$(COLOR_RESET)"
	@./scripts/stress-test.sh

# ─────────────────────────────────────────────────────────
#  CODE QUALITY
# ─────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run linter
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed$(COLOR_RESET)"; \
		echo "Install: https://golangci-lint.run/usage/install/"; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	@go fmt ./...
	@gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	@go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

# ─────────────────────────────────────────────────────────
#  DEVELOPMENT
# ─────────────────────────────────────────────────────────

.PHONY: run
run: build ## Build and run the application
	@echo "$(COLOR_GREEN)Running $(APP_NAME)...$(COLOR_RESET)"
	@$(BIN_DIR)/$(APP_NAME)

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@echo "$(COLOR_GREEN)Starting development server with hot reload...$(COLOR_RESET)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(COLOR_YELLOW)⚠ air not installed$(COLOR_RESET)"; \
		echo "Install: go install github.com/cosmtrek/air@latest"; \
		$(MAKE) run; \
	fi

.PHONY: dev-docker
dev-docker: ## Run development environment with Docker Compose
	@echo "$(COLOR_GREEN)Starting development environment...$(COLOR_RESET)"
	@docker compose up --build

.PHONY: dev-down
dev-down: ## Stop development environment
	@docker compose down

# ─────────────────────────────────────────────────────────
#  DOCKER
# ─────────────────────────────────────────────────────────

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_GREEN)Building Docker image...$(COLOR_RESET)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest \
		-f docker/Dockerfile .
	@echo "$(COLOR_GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(VERSION)$(COLOR_RESET)"

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "$(COLOR_GREEN)Pushing Docker image...$(COLOR_RESET)"
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest
	@echo "$(COLOR_GREEN)✓ Docker image pushed$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@docker run --rm -it \
		-v $(PWD)/data:/opt/gordon/data \
		-p 8081:8081 \
		-p 9100:9100 \
		$(DOCKER_IMAGE):latest

# ─────────────────────────────────────────────────────────
#  DEPLOYMENT
# ─────────────────────────────────────────────────────────

.PHONY: deploy-dev
deploy-dev: ## Deploy to development environment
	@echo "$(COLOR_GREEN)Deploying to development...$(COLOR_RESET)"
	@cd ansible && ./scripts/deploy.sh development

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	@echo "$(COLOR_GREEN)Deploying to staging...$(COLOR_RESET)"
	@cd ansible && ./scripts/deploy.sh staging

.PHONY: deploy-prod
deploy-prod: ## Deploy to production environment
	@echo "$(COLOR_YELLOW)⚠ Deploying to PRODUCTION$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd ansible && ./scripts/deploy.sh production; \
	else \
		echo "Deployment cancelled"; \
	fi

# ─────────────────────────────────────────────────────────
#  UTILITIES
# ─────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_GREEN)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BIN_DIR)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf vendor/
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: version
version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(shell go version)"

.PHONY: setup
setup: ## Setup development environment
	@echo "$(COLOR_GREEN)Setting up development environment...$(COLOR_RESET)"
	@$(MAKE) deps
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(COLOR_GREEN)✓ Development environment ready$(COLOR_RESET)"

.PHONY: generate-data
generate-data: ## Generate test data files
	@echo "$(COLOR_GREEN)Generating test data...$(COLOR_RESET)"
	@./scripts/generate-test-files.sh

.PHONY: logs
logs: ## Show logs (if running with systemd)
	@sudo journalctl -u gordon-watcher -f

.PHONY: status
status: ## Show service status
	@sudo systemctl status gordon-watcher


# ─────────────────────────────────────────────────────────
#  DEFAULT
# ─────────────────────────────────────────────────────────

.DEFAULT_GOAL := help
