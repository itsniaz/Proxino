# LAN Relay Makefile

.PHONY: all build build-frontend build-backend clean test install uninstall package help

# Variables
BINARY_NAME := lan-relay
VERSION := 1.0.0
BUILD_DIR := dist
BACKEND_DIR := backend
FRONTEND_DIR := frontend

# Go build flags
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION)"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
all: clean build

# Help target
help:
	@echo "LAN Relay Build System"
	@echo "====================="
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-frontend - Build frontend only"
	@echo "  build-backend  - Build backend only"
	@echo "  package        - Build for all platforms"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install locally"
	@echo "  uninstall      - Uninstall"
	@echo "  help           - Show this help"

# Build frontend
build-frontend:
	@echo "🎨 Building frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build
	@echo "✅ Frontend built"

# Prepare backend static files
prepare-backend: build-frontend
	@echo "📦 Preparing backend static files..."
	rm -rf $(BACKEND_DIR)/static
	cp -r $(FRONTEND_DIR)/build $(BACKEND_DIR)/static
	@echo "✅ Static files prepared"

# Build backend for current platform
build-backend: prepare-backend
	@echo "🚀 Building backend..."
	cd $(BACKEND_DIR) && go mod tidy
	cd $(BACKEND_DIR) && go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✅ Backend built: $(BACKEND_DIR)/$(BINARY_NAME)"

# Build everything for current platform
build: build-backend

# Test the application
test:
	@echo "🧪 Running tests..."
	cd $(BACKEND_DIR) && go test ./...
	cd $(FRONTEND_DIR) && npm test -- --coverage --silent --watchAll=false
	@echo "✅ Tests completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(BACKEND_DIR)/$(BINARY_NAME)
	rm -rf $(BACKEND_DIR)/static
	rm -rf $(FRONTEND_DIR)/build
	@echo "✅ Clean completed"

# Install locally
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@if [ -w /usr/local/bin ]; then \
		cp $(BACKEND_DIR)/$(BINARY_NAME) /usr/local/bin/; \
		chmod +x /usr/local/bin/$(BINARY_NAME); \
		echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "⚠️  Need sudo to install to /usr/local/bin"; \
		sudo cp $(BACKEND_DIR)/$(BINARY_NAME) /usr/local/bin/; \
		sudo chmod +x /usr/local/bin/$(BINARY_NAME); \
		echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi

# Uninstall
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "✅ Uninstalled"

# Package for all platforms
package: prepare-backend
	@echo "📦 Building packages for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		platform_split=($${platform//\// }); \
		GOOS=$${platform_split[0]}; \
		GOARCH=$${platform_split[1]}; \
		output_name=$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building $$output_name..."; \
		cd $(BACKEND_DIR) && env GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o ../$(BUILD_DIR)/$$output_name .; \
		if [ $$? -ne 0 ]; then \
			echo "❌ Failed to build $$output_name"; \
			exit 1; \
		fi; \
	done
	@echo "✅ All packages built in $(BUILD_DIR)/"

# Development server (for testing)
dev:
	@echo "🚀 Starting development server..."
	./scripts/start-dev.sh

# Quick test of the built binary
test-binary: build
	@echo "🧪 Testing built binary..."
	$(BACKEND_DIR)/$(BINARY_NAME) version
	$(BACKEND_DIR)/$(BINARY_NAME) --help
	@echo "✅ Binary test completed" 