# Atlassian Assets CLI - Build and Test Automation
# ==================================================

# Variables
BINARY_NAME=assets
BUILD_DIR=bin
MAIN_PATH=./cmd/assets
VERSION?=dev
COMMIT?=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-X github.com/aaronsb/atlassian-assets/internal/version.Version=$(VERSION) \
        -X github.com/aaronsb/atlassian-assets/internal/version.Commit=$(COMMIT) \
        -X github.com/aaronsb/atlassian-assets/internal/version.Date=$(DATE)

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build build-dev build-release clean test test-help test-integration lint format check install uninstall git-tag git-push version

# Default target
all: build

# Help target
help:
	@echo "$(BLUE)Atlassian Assets CLI - Build System$(NC)"
	@echo "===================================="
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  $(YELLOW)build$(NC)         - Interactive build with options"
	@echo "  $(YELLOW)build-dev$(NC)     - Quick development build"
	@echo "  $(YELLOW)build-release$(NC) - Production build with version injection"
	@echo "  $(YELLOW)clean$(NC)         - Remove build artifacts"
	@echo "  $(YELLOW)test$(NC)          - Run all tests"
	@echo "  $(YELLOW)test-help$(NC)     - Run help system tests only"
	@echo "  $(YELLOW)test-integration$(NC) - Run integration tests (requires credentials)"
	@echo "  $(YELLOW)lint$(NC)          - Run code linting"
	@echo "  $(YELLOW)format$(NC)        - Format code"
	@echo "  $(YELLOW)check$(NC)         - Run tests + lint + format check"
	@echo "  $(YELLOW)install$(NC)       - Install binary to system PATH"
	@echo "  $(YELLOW)uninstall$(NC)     - Remove binary from system PATH"
	@echo "  $(YELLOW)git-tag$(NC)       - Create and push git tag"
	@echo "  $(YELLOW)git-push$(NC)      - Push to remote repository"
	@echo "  $(YELLOW)version$(NC)       - Show version information"
	@echo ""
	@echo "$(GREEN)Usage examples:$(NC)"
	@echo "  make build              # Interactive build"
	@echo "  make build-release VERSION=v1.0.0"
	@echo "  make test"
	@echo "  make git-tag VERSION=v1.0.0"

# Interactive build target
build:
	@echo "$(BLUE)🔨 Atlassian Assets CLI Builder$(NC)"
	@echo "================================="
	@echo ""
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "$(YELLOW)⚠️  Existing binary found: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"; \
		echo ""; \
		read -p "Delete existing binary and build clean? [y/N]: " choice; \
		case "$$choice" in \
			y|Y ) \
				echo "$(RED)🗑️  Removing existing binary...$(NC)"; \
				rm -f $(BUILD_DIR)/$(BINARY_NAME); \
				echo "$(GREEN)✅ Binary removed$(NC)"; \
				;; \
			* ) \
				echo "$(YELLOW)📦 Keeping existing binary, building anyway...$(NC)"; \
				;; \
		esac; \
		echo ""; \
	fi
	@echo "$(BLUE)🛠️  Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Build successful: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@echo ""
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@$(MAKE) test-help
	@echo ""
	@echo "$(GREEN)🎉 Build and test complete!$(NC)"
	@echo "$(BLUE)📋 Next steps:$(NC)"
	@echo "  • Test the binary: ./$(BUILD_DIR)/$(BINARY_NAME) --version"
	@echo "  • Run full tests: make test"
	@echo "  • Install to PATH: make install"

# Development build (quick)
build-dev:
	@echo "$(BLUE)🔨 Quick development build...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Development build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Release build with version injection
build-release:
	@echo "$(BLUE)🚀 Building release version $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Release build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@echo "$(BLUE)📋 Version information:$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) --version

# Clean build artifacts
clean:
	@echo "$(RED)🗑️  Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "$(GREEN)✅ Clean complete$(NC)"

# Run all tests
test:
	@echo "$(BLUE)🧪 Running all tests...$(NC)"
	@chmod +x run_tests.sh
	@./run_tests.sh

# Run help system tests only
test-help:
	@echo "$(BLUE)📋 Running help system tests...$(NC)"
	@go test -v ./cmd/assets -run TestHelpOutput
	@echo "$(GREEN)✅ Help tests complete$(NC)"

# Run integration tests (requires credentials)
test-integration:
	@echo "$(BLUE)🔗 Running integration tests...$(NC)"
	@if [ -z "$$ATLASSIAN_EMAIL" ] || [ -z "$$ATLASSIAN_API_TOKEN" ]; then \
		echo "$(RED)❌ Integration tests require ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN$(NC)"; \
		exit 1; \
	fi
	@go test -v ./cmd/assets -run TestIntegration
	@echo "$(GREEN)✅ Integration tests complete$(NC)"

# Lint code
lint:
	@echo "$(BLUE)🔍 Running code linting...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found, running go vet instead$(NC)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✅ Linting complete$(NC)"

# Format code
format:
	@echo "$(BLUE)🎨 Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatting complete$(NC)"

# Check all (tests + lint + format)
check: format lint test

# Install binary to system PATH
install: build-release
	@echo "$(BLUE)📦 Installing $(BINARY_NAME) to system PATH...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ Installation complete$(NC)"
	@echo "$(BLUE)📋 You can now run: $(BINARY_NAME) --version$(NC)"

# Uninstall binary from system PATH
uninstall:
	@echo "$(RED)🗑️  Removing $(BINARY_NAME) from system PATH...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ Uninstallation complete$(NC)"

# Create and push git tag
git-tag:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "$(RED)❌ Please specify VERSION: make git-tag VERSION=v1.0.0$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)🏷️  Creating git tag $(VERSION)...$(NC)"
	@git add .
	@git status
	@echo ""
	@read -p "Commit and tag as $(VERSION)? [y/N]: " choice; \
	case "$$choice" in \
		y|Y ) \
			git commit -m "Release $(VERSION)"; \
			git tag -a $(VERSION) -m "Release $(VERSION)"; \
			echo "$(GREEN)✅ Tag $(VERSION) created$(NC)"; \
			echo "$(BLUE)📋 To push: make git-push VERSION=$(VERSION)$(NC)"; \
			;; \
		* ) \
			echo "$(YELLOW)⚠️  Tag creation cancelled$(NC)"; \
			;; \
	esac

# Push to remote repository
git-push:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "$(BLUE)📤 Pushing main branch to remote...$(NC)"; \
		git push origin main; \
	else \
		echo "$(BLUE)📤 Pushing $(VERSION) and main to remote...$(NC)"; \
		git push origin main; \
		git push origin $(VERSION); \
	fi
	@echo "$(GREEN)✅ Push complete$(NC)"

# Show version information
version:
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "$(BLUE)📋 Current binary version:$(NC)"; \
		./$(BUILD_DIR)/$(BINARY_NAME) --version; \
	else \
		echo "$(YELLOW)⚠️  No binary found. Run 'make build' first.$(NC)"; \
	fi
	@echo ""
	@echo "$(BLUE)📋 Build configuration:$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"